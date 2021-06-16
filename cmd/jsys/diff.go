package main

// Goes through diffs to see if there's any interesting state we forgot to persist

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/sliceutil"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

func diffEntrypoint() *cobra.Command {
	maxDiffFilesFind := 100

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Goes through diffs to see if there's any interesting state we forgot to persist",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(diff(maxDiffFilesFind))
		},
	}

	cmd.Flags().IntVarP(&maxDiffFilesFind, "max-diff-files-find", "m", maxDiffFilesFind, "Maximum # of diff files to report before bailing out")

	return cmd
}

func diffOneEntrypoint() *cobra.Command {
	maxDiffFilesFind := 100

	cmd := &cobra.Command{
		Use: "diff1 [path]",
		// Short: "Goes through diffs to see if there's any interesting state we forgot to persist",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(diffOne(args[0]))
		},
	}

	cmd.Flags().IntVarP(&maxDiffFilesFind, "max-diff-files-find", "m", maxDiffFilesFind, "Maximum # of diff files to report before bailing out")

	return cmd
}

func revertEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "revert [path]",
		Short: "Revert a file from diffs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(revert(args[0]))
		},
	}
}

type diffReport struct {
	output          io.Writer
	totallyNewFiles int
}

func (d *diffReport) TotallyNewFile(entryPathCanonical string) {
	d.totallyNewFiles++

	fmt.Fprintf(d.output, " N %s\n", entryPathCanonical)
}

func (d *diffReport) ModifiedFromImage(entryPathCanonical string) {
	d.totallyNewFiles++

	fmt.Fprintf(d.output, "IM %s\n", entryPathCanonical)
}

func (d *diffReport) OverrideEqual(entryPathCanonical string) {
	// fmt.Fprintf(d.output, "OVERRIDE equal: %s\n", entryPathCanonical)
}

func (d *diffReport) OverrideOutOfDate(entryPathCanonical string) {
	// "override modified"
	fmt.Fprintf(d.output, "OM %s\n", entryPathCanonical)
}

func dirBasesForRunningSystemId() (*dirBases, error) {
	// resolving diffs for currently running system. if needed, we could add a CLI flag for override
	runningSysId, err := readRunningSystemId()
	if err != nil {
		return nil, err
	}

	return &dirBases{
		// root differences to system N descend from here
		diff:             fmt.Sprintf("/persist/apps/SYSTEM_nobackup/%s-diff/", runningSysId),
		overridesWorkdir: "/persist/work/joonas-sys/overrides",
		image:            "/mnt/sys-current-rom",
	}, nil
}

// TODO: show file content diffs
func diff(maxDiffFilesFind int) error {
	if err := requireRoot(); err != nil {
		return err
	}

	_, allowedChangeSubtrees, allowedChangeFiles, err := loadConf()
	if err != nil {
		return err
	}

	dirs, err := dirBasesForRunningSystemId()
	if err != nil {
		return err
	}

	exists, err := osutil.Exists(dirs.diff)
	if err != nil || !exists { // fine if nil error
		return fmt.Errorf("diff dir doesn't exist: %s: %w", dirs.diff, err)
	}

	report := &diffReport{output: os.Stdout}

	diffFilesFound := 0

	var walkOneDir func(string, string) error
	walkOneDir = func(dirCanonical string, dirDiff string) error {
		entries, err := ioutil.ReadDir(dirDiff)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if report.totallyNewFiles == maxDiffFilesFind {
				return errors.New("too many diffing files found - bailing out")
			}

			entryPath := newEntryPathBuilder(entry.Name(), dirCanonical, dirs)

			// TODO: it might not be safe to ignore symlinks. they're data also?
			if (entry.Mode() & os.ModeSymlink) != 0 {
				continue
			}

			if entry.IsDir() {
				if sliceutil.ContainsString(allowedChangeSubtrees, entryPath.Canonical()) { // allowed to be changed
					continue
				}

				if err := walkOneDir(entryPath.Canonical(), entryPath.Diff()); err != nil {
					return err
				}
			} else {
				if sliceutil.ContainsString(allowedChangeFiles, entryPath.Canonical()) { // allowed to be changed
					continue
				}

				overrideExists, err := osutil.Exists(entryPath.OverridesWorkdir())
				if err != nil {
					return err
				}

				imageExists, err := osutil.Exists(entryPath.Image())
				if err != nil {
					return err
				}

				if overrideExists {
					equal, err := compareFiles(entryPath.Diff(), entryPath.OverridesWorkdir())
					if err != nil {
						return err
					}

					if equal {
						report.OverrideEqual(entryPath.Canonical())
					} else {
						report.OverrideOutOfDate(entryPath.Canonical())
					}
				} else if imageExists {
					report.ModifiedFromImage(entryPath.Canonical())
				} else {
					report.TotallyNewFile(entryPath.Canonical())
				}
			}
		}

		return nil
	}

	if err := walkOneDir("/", dirs.diff); err != nil {
		return err
	}

	if diffFilesFound > 0 {
		_ = os.Stdout.Sync()

		return fmt.Errorf("unclean system: %d file(s) with diffs found", diffFilesFound)
	}

	return nil
}

type dirBases struct {
	diff             string // "/persist/apps/SYSTEM_nobackup/system_a-diff"
	overridesWorkdir string // "/persist/work/joonas-sys/overrides"
	image            string // "/mnt/sys-current-rom"
}

// same file's paths in different places (canonical / diff tree / overrides workdir)
type entryPathBuilder struct {
	basename     string // "timezone"
	dirCanonical string // "/etc"
	dirs         *dirBases
}

func newEntryPathBuilder(basename string, dirCanonical string, dirs *dirBases) *entryPathBuilder {
	return &entryPathBuilder{
		basename:     basename,
		dirCanonical: dirCanonical,
		dirs:         dirs,
	}
}

func newEntryPathBuilderFromPath(entryPathCanonical string, dirs *dirBases) *entryPathBuilder {
	return newEntryPathBuilder(filepath.Base(entryPathCanonical), filepath.Dir(entryPathCanonical), dirs)
}

// "/etc/timezone"
func (e *entryPathBuilder) Canonical() string {
	return filepath.Join(e.dirCanonical, e.basename)
}

// "/mnt/sys-current-rom/etc/timezone"
func (e *entryPathBuilder) Image() string {
	return filepath.Join(e.dirs.image, e.Canonical())
}

// "/persist/apps/SYSTEM_nobackup/system_a-diff/etc/timezone"
func (e *entryPathBuilder) Diff() string {
	return filepath.Join(e.dirs.diff, e.Canonical())
}

// "/persist/work/joonas-sys/overrides/etc/timezone"
func (e *entryPathBuilder) OverridesWorkdir() string {
	return filepath.Join(e.dirs.overridesWorkdir, e.Canonical())
}

func compareFiles(pathA string, pathB string) (bool, error) {
	contentA, err := os.ReadFile(pathA)
	if err != nil {
		return false, err
	}

	contentB, err := os.ReadFile(pathB)
	if err != nil {
		return false, err
	}

	return bytes.Equal(contentA, contentB), nil
}

func diffOne(entryPathCanonical string) error {
	dirs, err := dirBasesForRunningSystemId()
	if err != nil {
		return err
	}

	entryPath := newEntryPathBuilderFromPath(entryPathCanonical, dirs)

	overrideExists, err := osutil.Exists(entryPath.OverridesWorkdir())
	if err != nil {
		return err
	}

	/*
		Image
		└── Overrides workdir
		    └── Current
	*/
	pathPrevious, pathUpdated := func() (string, string) {
		if overrideExists {
			return entryPath.OverridesWorkdir(), entryPath.Canonical()
		} else {
			return entryPath.Image(), entryPath.Canonical()
		}
	}()

	contentPrevious, err := os.ReadFile(pathPrevious)
	if err != nil {
		return err
	}

	contentUpdated, err := os.ReadFile(pathUpdated)
	if err != nil {
		return err
	}

	diffText(string(contentPrevious), string(contentUpdated))

	// fmt.Println(diff)

	return nil
}

// https://gist.github.com/ik5/d8ecde700972d4378d87#gistcomment-3074524
const (
	RedColor   = "\033[1;31m%s\033[0m"
	GreenColor = "\033[1;32m%s\033[0m"
)

func diffText(previous, updated string) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(previous),
		B:        difflib.SplitLines(updated),
		FromFile: "previous",
		ToFile:   "updated",
		Context:  5,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)

	// ugly hack to color the output
	for _, line := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			fmt.Println(fmt.Sprintf(GreenColor, line))
		case strings.HasPrefix(line, "-"):
			fmt.Println(fmt.Sprintf(RedColor, line))
		default:
			fmt.Println(line)
		}
	}

	// fmt.Printf(text)
	/*
	   		dmp := diffmatchpatch.New()

	   // https://github.com/sergi/go-diff/issues/69#issuecomment-688602689
	   previousDMP, updatedDMP, dmpStrings := dmp.DiffLinesToChars(previous, updated)
	   diffs := dmp.DiffMain(previousDMP, updatedDMP, false)
	   diffs2 := dmp.DiffCharsToLines(diffs, dmpStrings)
	   diffs3 := dmp.DiffCleanupSemantic(diffs2)

	   		return dmp.DiffPrettyText(diffs3)
	*/
}

// WARNING: you may need to do also: $ echo 2 > /proc/sys/vm/drop_caches
func revert(entryPathCanonical string) error {
	dirs, err := dirBasesForRunningSystemId()
	if err != nil {
		return err
	}

	entryPath := newEntryPathBuilderFromPath(entryPathCanonical, dirs)

	exists, err := osutil.Exists(entryPath.Diff())
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("file to revert is not in diffs: %s", entryPathCanonical)
	}

	return os.Remove(entryPath.Diff())
}
