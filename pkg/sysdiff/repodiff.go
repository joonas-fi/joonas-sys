// Compare changes to running system from the build baseline in repo
package sysdiff

import (
	"bytes"
	"context"
	"crypto/sha1"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/spf13/cobra"
)

func Entrypoint() *cobra.Command {
	verbose := false

	cmd := &cobra.Command{
		Use:   "repo-diff",
		Short: "Show differences in current system compared to in repo's overrides/ dir",
		Args:  cobra.NoArgs,
		Run: cli.RunnerNoArgs(func(_ context.Context, _ *log.Logger) error {
			return repoDiff(verbose)
		}),
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", verbose, "Log more details")

	return cmd
}

func repoDiff(verbose bool) error {
	sysID, err := common.ReadRunningSystemId()
	if err != nil {
		return err
	}

	sysDiffPath := filelocations.Sysroot.Diff(sysID)

	return filepath.WalkDir("overrides/", func(pathInOverrides string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		isActualFile := func() bool {
			if dirEntry.IsDir() { // only process files
				return false
			}

			if dirEntry.Name() == ".empty_dir" {
				return false
			}

			return true
		}

		if !isActualFile() {
			return nil
		}

		// overrides/etc/wireguard => etc/wireguard
		canonicalName := strings.TrimPrefix(pathInOverrides, "overrides")

		fileInDiffs := filepath.Join(sysDiffPath, canonicalName)

		exists, err := osutil.Exists(fileInDiffs)
		if err != nil {
			return err
		}

		if !exists {
			if verbose {
				log.Printf("not overridden: %s", canonicalName)
			}
		} else {
			same, err := isSameFile(pathInOverrides, fileInDiffs)
			if err != nil {
				return err
			}

			if same {
				if verbose {
					log.Printf("overridden but same: %s", canonicalName)
				}
			} else {
				log.Printf("overridden and different: %s", canonicalName)
			}
		}

		return nil
	})
}

func isSameFile(path1, path2 string) (bool, error) {
	stat1, err := os.Lstat(path1)
	if err != nil {
		return false, err
	}
	stat2, err := os.Lstat(path2)
	if err != nil {
		return false, err
	}

	isSymlink := func(fi fs.FileInfo) int {
		if fi.Mode()&fs.ModeSymlink != 0x00 {
			return 1
		} else {
			return 0
		}
	}

	switch isSymlink(stat1) + isSymlink(stat2) {
	case 0: // neither are -> continue investigating
		// no-op
	case 1: // only one is a symlink -> definitely not equal
		return false, nil
	case 2: // both are a symlink -> compare destination paths (*not* destinations themselfes)
		link1, err := os.Readlink(path1)
		if err != nil {
			return false, err
		}

		link2, err := os.Readlink(path2)
		if err != nil {
			return false, err
		}

		if link1 == link2 { // point to the same destination
			return true, nil
		} else {
			return false, nil
		}
	default:
		panic("shouldn't be here")
	}

	if stat1.Size() != stat2.Size() { // different sizes -> must be different
		return false, nil
	}

	file1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer file1.Close()

	file2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer file2.Close()

	hash1, hash2 := sha1.New(), sha1.New()

	if _, err := io.Copy(hash1, file1); err != nil {
		return false, err
	}

	if _, err := io.Copy(hash2, file2); err != nil {
		return false, err
	}

	return bytes.Equal(hash1.Sum(nil), hash2.Sum(nil)), nil
}
