// Goes through diffs to see if there's any interesting state we forgot to persist
package sysdiff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/function61/gokit/sliceutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

func Entrypoints() []*cobra.Command {
	return []*cobra.Command{
		entrypoint(),
		repoDiffEntrypoint(),
		diffOneEntrypoint(),
		&cobra.Command{
			Use:   "revert [path]",
			Short: "Revert a file from diffs",
			Args:  cobra.ExactArgs(1),
			Run: cli.Runner(func(_ context.Context, args []string, _ *log.Logger) error {
				return revert(args[0])
			}),
		},
	}
}

func entrypoint() *cobra.Command {
	maxDiffFilesFind := 100
	ignoreDeleted := false

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Goes through diffs to see if there's any interesting state we forgot to persist",
		Args:  cobra.NoArgs,
		Run: cli.RunnerNoArgs(func(_ context.Context, _ *log.Logger) error {
			return diff(maxDiffFilesFind, ignoreDeleted)
		}),
	}

	cmd.Flags().IntVarP(&maxDiffFilesFind, "max-diff-files-find", "m", maxDiffFilesFind, "Maximum # of diff files to report before bailing out")
	cmd.Flags().BoolVarP(&ignoreDeleted, "ignore-deleted", "", ignoreDeleted, "Ignore deleted files")

	return cmd
}

func diffOneEntrypoint() *cobra.Command {
	maxDiffFilesFind := 100

	cmd := &cobra.Command{
		Use:   "diff1 [path]",
		Short: "Diffs one file from current running system to the baseline",
		Args:  cobra.ExactArgs(1),
		Run: cli.Runner(func(_ context.Context, args []string, _ *log.Logger) error {
			return diffOne(args[0])
		}),
	}

	cmd.Flags().IntVarP(&maxDiffFilesFind, "max-diff-files-find", "m", maxDiffFilesFind, "Maximum # of diff files to report before bailing out")

	return cmd
}

type diffReport struct {
	output          io.Writer
	totallyNewFiles int
	overrideEqual   int
	imageEqual      int
	ignoreDeleted   bool
}

func (d *diffReport) Deleted(entryPathCanonical string) {
	if !d.ignoreDeleted {
		fmt.Fprintf(d.output, " D %s\n", entryPathCanonical)
	}
}

func (d *diffReport) TotallyNewFile(entryPathCanonical string) {
	d.totallyNewFiles++

	fmt.Fprintf(d.output, " N %s\n", entryPathCanonical)
}

func (d *diffReport) ImageModified(entryPathCanonical string) {
	d.totallyNewFiles++

	fmt.Fprintf(d.output, "IM %s\n", entryPathCanonical)
}

func (d *diffReport) ImageEqual(entryPathCanonical string) {
	d.imageEqual++
}

func (d *diffReport) OverrideEqual(entryPathCanonical string) {
	d.overrideEqual++
	// fmt.Fprintf(d.output, "OVERRIDE equal: %s\n", entryPathCanonical)
}

func (d *diffReport) OverrideOutOfDate(entryPathCanonical string) {
	// "override modified"
	fmt.Fprintf(d.output, "OM %s\n", entryPathCanonical)
}

func dirBasesForRunningSystemId() (*dirBases, error) {
	// resolving diffs for currently running system. if needed, we could add a CLI flag for override
	runningSysID, err := common.ReadRunningSystemId()
	if err != nil {
		return nil, err
	}

	return &dirBases{
		// root differences to system N descend from here
		diff:             filelocations.Sysroot.Diff(runningSysID),
		overridesWorkdir: "/persist/work/joonas-sys/overrides",
		image:            filelocations.Sysroot.Checkout(runningSysID),
	}, nil
}

// TODO: show file content diffs
func diff(maxDiffFilesFind int, ignoreDeleted bool) error {
	if _, err := userutil.RequireRoot(); err != nil {
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

	report := &diffReport{output: os.Stdout, ignoreDeleted: ignoreDeleted}

	// we're running as root, but even root cannot access everything.
	// one such case is FUSE filesystems.
	errCannotAccess := 0

	errDirChangedToSymlink := []string{}

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

			if sliceutil.ContainsString(allowedChangeFiles, entryPath.Canonical()) { // allowed to be changed
				continue
			}

			switch {
			case entry.IsDir():
				if sliceutil.ContainsString(allowedChangeSubtrees, entryPath.Canonical()) { // allowed to be changed
					continue
				}

				if err := walkOneDir(entryPath.Canonical(), entryPath.Diff()); err != nil {
					return err
				}
			case (entry.Mode() & os.ModeCharDevice) != 0: // assuming whiteout file (= deleted marker)
				report.Deleted(entryPath.Canonical())
			case isSymlink(entry):
				// deleted symlink is already handled by a whiteout file, so we have create / modify to handle

				overrideExists, err := osutil.Exists(entryPath.OverridesWorkdir())
				if err != nil {
					return fmt.Errorf("Exists: %w", err)
				}

				imageExists, err := osutil.Exists(entryPath.Image())
				if err != nil {
					return fmt.Errorf("Exists: %w", err)
				}

				if overrideExists || imageExists { // modified or modified-then-reverted-to-no-change
					equal, err := func() (bool, error) {
						imageOrOverride := func() string {
							if overrideExists {
								return entryPath.OverridesWorkdir()
							} else {
								return entryPath.Image()
							}
						}()

						// diff is a symlink, and image exists.
						// in some rare cases image is not a symlink (think dir replaced with symlink)
						imageStat, err := os.Lstat(imageOrOverride)
						if err != nil {
							return false, err
						}

						if !isSymlink(imageStat) {
							errDirChangedToSymlink = append(errDirChangedToSymlink, entryPath.Canonical())
							return false, nil
						}

						// if equal => modified-then-reverted-to-no-change
						return symlinksEqual(entryPath.Diff(), imageOrOverride)
					}()
					if err != nil {
						return err
					}

					if equal {
						report.ImageEqual(entryPath.Canonical())
					} else {
						report.ImageModified(entryPath.Canonical())
					}
				} else {
					report.TotallyNewFile(entryPath.Canonical())
				}
			default: // regular file
				overrideExists, err := osutil.Exists(entryPath.OverridesWorkdir())
				if err != nil {
					errCannotAccess++
					continue
				}

				imageExists, err := osutil.Exists(entryPath.Image())
				if err != nil {
					return err
				}

				if overrideExists {
					equal, err := filesEqual(entryPath.Diff(), entryPath.OverridesWorkdir())
					if err != nil {
						return fmt.Errorf("filesEqual: %w", err)
					}

					if equal {
						report.OverrideEqual(entryPath.Canonical())
					} else {
						report.OverrideOutOfDate(entryPath.Canonical())
					}
				} else if imageExists {
					equal, err := filesEqual(entryPath.Diff(), entryPath.Image())
					if err != nil {
						return err
					}

					if equal {
						report.ImageEqual(entryPath.Canonical())
					} else {
						report.ImageModified(entryPath.Canonical())
					}
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

	if errCannotAccess > 0 {
		log.Printf("WARN: errCannotAccess=%d", errCannotAccess)
	}

	for _, item := range errDirChangedToSymlink {
		log.Printf("WARN: errDirChangedToSymlink (changes not analyzed): %s", item)
	}

	if report.overrideEqual > 0 {
		log.Printf("WARN: overrideEqual=%d", report.overrideEqual)
	}

	if report.imageEqual > 0 {
		log.Printf("WARN: imageEqual=%d (modified without actually modified? just modification timestamp?)", report.imageEqual)
	}

	return nil
}

type dirBases struct {
	diff             string // "/sysroot/apps/OS-diff/<sysID>"
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

// "/sysroot/apps/OS-diff/<sysID>/etc/timezone"
func (e *entryPathBuilder) Diff() string {
	return filepath.Join(e.dirs.diff, e.Canonical())
}

// "/persist/work/joonas-sys/overrides/etc/timezone"
func (e *entryPathBuilder) OverridesWorkdir() string {
	return filepath.Join(e.dirs.overridesWorkdir, e.Canonical())
}

func filesEqual(pathA string, pathB string) (bool, error) {
	// TODO: implementation could be optimized to stop reading on first different byte
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

func symlinksEqual(pathA string, pathB string) (bool, error) {
	linkA, err := os.Readlink(pathA)
	if err != nil {
		return false, err
	}

	linkB, err := os.Readlink(pathB)
	if err != nil {
		return false, err
	}

	return linkA == linkB, nil
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

	if string(contentUpdated) == string(contentPrevious) {
		fmt.Println("files are binary equal")
		return nil
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

func isSymlink(info fs.FileInfo) bool {
	return (info.Mode() & os.ModeSymlink) != 0
}
