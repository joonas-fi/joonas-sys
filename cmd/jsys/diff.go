package main

// Goes through diffs to see if there's any interesting state we forgot to persist

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/sliceutil"
	"github.com/spf13/cobra"
)

func diffEntrypoint() *cobra.Command {
	maxDiffFilesFind := 100

	cmd := &cobra.Command{
		Use:     "diff",
		Short:   "Goes through diffs to see if there's any interesting state we forgot to persist",
		Args:    cobra.NoArgs,
		Version: dynversion.Version,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(diff(maxDiffFilesFind))
		},
	}

	cmd.Flags().IntVarP(&maxDiffFilesFind, "max-diff-files-find", "m", maxDiffFilesFind, "Maximum # of diff files to report before bailing out")

	return cmd
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

	// resolving diffs for currently running system. if needed, we could add a CLI flag for override
	runningSysId, err := readRunningSystemId()
	if err != nil {
		return err
	}

	// root differences to system N descend from here
	sysDiffsRoot := fmt.Sprintf("/persist/sys-%s-diff", runningSysId)

	exists, err := osutil.Exists(sysDiffsRoot)
	if err != nil || !exists { // fine if nil error
		return fmt.Errorf("system diffs for %s don't exist: %w", runningSysId, err)
	}

	diffFilesFound := 0

	var walkOneDir func(string, string) error
	walkOneDir = func(dirCanonical string, dirInternal string) error {
		entries, err := ioutil.ReadDir(dirInternal)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			entryPathDiff := filepath.Join(dirInternal, entry.Name())
			entryPathCanonical := filepath.Join(dirCanonical, entry.Name())

			// TODO: it might not be safe to ignore symlinks. they're data also?
			if (entry.Mode() & os.ModeSymlink) != 0 {
				continue
			}

			if entry.IsDir() {
				if sliceutil.ContainsString(allowedChangeSubtrees, entryPathCanonical) { // allowed to be changed
					continue
				}

				if err := walkOneDir(entryPathCanonical, entryPathDiff); err != nil {
					return err
				}
			} else {
				if sliceutil.ContainsString(allowedChangeFiles, entryPathCanonical) { // allowed to be changed
					continue
				}

				diffFilesFound++

				fmt.Fprintf(os.Stdout, "%s\n", entryPathCanonical)

				if diffFilesFound == maxDiffFilesFind {
					return errors.New("too many diffing files found - bailing out")
				}
			}
		}

		return nil
	}

	if err := walkOneDir("/", sysDiffsRoot); err != nil {
		return err
	}

	if diffFilesFound > 0 {
		_ = os.Stdout.Sync()

		return fmt.Errorf("unclean system: %d file(s) with diffs found", diffFilesFound)
	}

	return nil
}
