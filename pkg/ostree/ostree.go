// root filesystem storage and transport in OSTree
package ostree

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/joonas-fi/joonas-sys/pkg/gostree"
	"github.com/joonas-fi/joonas-sys/pkg/xdgcommonextendedattributes"
	"github.com/pkg/xattr"
	"github.com/samber/lo"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
)

const (
	ostreeBranchNameX8664 = "deploy/app/fi.joonas.os/x86_64/stable"
	ostreeRemoteName      = "fi.joonas.os"
)

func PullEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull updates from joonas-sys",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			if _, err := userutil.RequireRoot(); err != nil {
				return err
			}

			logOutput := exec.CommandContext(ctx, "ostree", "pull", ostreeRemoteName, ostreeBranchNameX8664)
			logOutput.Stdout = os.Stdout
			logOutput.Stderr = os.Stderr
			if err := logOutput.Run(); err != nil {
				return err
			}

			repo := gostree.Open(filelocations.Sysroot.App(common.AppOSRepo))

			commitID, err := repo.ResolveRef(ostreeBranchNameX8664, remoteNameParam())
			if err != nil {
				return err
			}

			commit, err := repo.ReadCommit(commitID)
			if err != nil {
				return err
			}

			fmt.Printf(
				"done. got head:\n  %s %s\n",
				commit.GetTimestamp().Format(time.RFC3339),
				commit.Subject)

			return nil
		}),
	}
}

func Entrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ostree",
		Short: "OSTree + jsys management",
	}

	cmd.AddCommand(PullEntrypoint())

	cmd.AddCommand(&cobra.Command{
		Use:   "log",
		Short: "Show commits",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			repo := gostree.Open(filelocations.Sysroot.App(common.AppOSRepo))

			commitID, err := repo.ResolveRef(ostreeBranchNameX8664, remoteNameParam())
			if err != nil {
				return err
			}
			commitLog, err := repo.ReadParentCommits(commitID)
			if err != nil {
				return err
			}

			commitLogTbl := termtables.CreateTable()

			commitLogTbl.AddHeaders("Date", "Subject", "Subject")

			for _, commit := range commitLog {
				commitLogTbl.AddRow(
					commitShort(commit.ID),
					commit.GetTimestamp().Format("2006-01-02 15:04"),
					commit.Subject,
				)
			}

			fmt.Println(commitLogTbl.Render())

			return nil
		}),
	})

	cmd.AddCommand(commitEntrypoint())
	cmd.AddCommand(pushEntrypoint())

	cmd.AddCommand(&cobra.Command{
		Use:   "checkout <commit>",
		Short: "Checks out a root filesystem from a commit",
		Args:  cobra.ExactArgs(1),
		Run: cli.WrapRun(func(ctx context.Context, args []string) error {
			if err := checkoutRootFS(ctx, args[0]); err != nil {
				return err
			}

			fmt.Printf("pro-tip:\n  $ %s test-in-vm\nOR\n  $ %s flash\n", os.Args[0], os.Args[0])

			return nil
		}),
	})

	cmd.AddCommand(checkoutsCleanupEntrypoint())

	return cmd
}

func commitEntrypoint() *cobra.Command {
	checkout := false

	cmd := &cobra.Command{
		Use:   "commit <subject>",
		Short: "Commit current build to OSTree",
		Args:  cobra.ExactArgs(1),
		Run: cli.WrapRun(func(ctx context.Context, args []string) error {
			subject := args[0]

			if _, err := userutil.RequireRoot(); err != nil {
				return err
			}

			devEntries, err := os.ReadDir(filepath.Join(common.BuildTreeLocation, "/dev"))
			if err != nil {
				return fmt.Errorf("/dev: %w", err)
			}

			// files under /dev are special (created by deboostrap for backwards compat reasons?
			// usually devtmpfs is mounted on /dev (so those are not even stored on disk).
			// OSTree only supports regular files and symlinks.
			if len(devEntries) > 0 {
				slog.Info("removing files in /dev")

				for _, devEntry := range devEntries {
					slog.Debug("removing file in /dev", "file", devEntry.Name())

					if err := os.RemoveAll(filepath.Join(common.BuildTreeLocation, "/dev", devEntry.Name())); err != nil {
						return err
					}
				}
			}

			commitOutput := exec.CommandContext(ctx, "ostree", "commit", "--branch="+ostreeBranchNameX8664, "--subject="+subject, common.BuildTreeLocation)
			// stdout would only show created commit ID
			commitOutput.Stderr = os.Stderr
			if err := commitOutput.Run(); err != nil {
				return err
			}

			repo := gostree.Open(filelocations.Sysroot.App(common.AppOSRepo))

			commitID, err := repo.ResolveRef(ostreeBranchNameX8664, remoteNameParam())
			if err != nil {
				return err
			}

			if checkout {
				return checkoutRootFS(ctx, commitID)
			} else {
				slog.Info("committed", "commitID", commitID)
			}

			return nil
		}),
	}

	cmd.Flags().BoolVarP(&checkout, "checkout", "", checkout, "Run checkout right after commit")

	return cmd
}

func checkoutRootFS(ctx context.Context, commit string) error {
	if _, err := userutil.RequireRoot(); err != nil {
		return err
	}

	checkoutPath := filelocations.Sysroot.Checkout(commitShort(commit))

	exists, err := osutil.Exists(checkoutPath)
	if err != nil {
		return err
	}

	if !exists {
		commitOutput := exec.CommandContext(ctx, "ostree", "checkout", commit, checkoutPath)
		commitOutput.Stdout = os.Stdout
		commitOutput.Stderr = os.Stderr

		if err := commitOutput.Run(); err != nil {
			return err
		}
	}

	commitObj, err := gostree.Open(filelocations.Sysroot.App(common.AppOSRepo)).ReadCommit(commit)
	if err != nil {
		return err
	}

	if err := xattr.Set(checkoutPath, xdgcommonextendedattributes.Comment, []byte(commitObj.Subject)); err != nil {
		return err
	}

	slog.Info("checked out to", "checkoutPath", checkoutPath)

	return nil
}

type CheckoutWithLabel struct {
	CommitShort string // "ae39405"
	Commit      string // "ae39405..." (set only if sourced from ostree commit log and not from actual checkouts)
	Label       string // "ae39405 - 2023-05-28T13:06:07+03:00 - fix virtio-fsd"
	Timestamp   time.Time
}

// returns true if this represents a commit that has been checked out.
// if not, returns the commit id so we can checkout based on that.
func (c CheckoutWithLabel) CheckedOut() (string, bool) {
	checkedOut := c.Commit == ""
	return c.Commit, checkedOut
}

func listVersionsFromCheckouts(root filelocations.Root) ([]CheckoutWithLabel, error) {
	versionsEntries, err := os.ReadDir(root.CheckoutsDir())
	if err != nil {
		return nil, err
	}

	type direntryWithTimestamp struct {
		fs.DirEntry
		timestamp time.Time
	}

	// decorate with timestamps
	checkoutsWithTimestamps := lo.Map(versionsEntries, func(e fs.DirEntry, _ int) direntryWithTimestamp {
		info, err := e.Info()
		if err != nil {
			panic(err)
		}

		// modtime is set to 1970 by OSTree. need to dig deeper.
		tim := func() time.Time {
			allTimes := times.Get(info)

			if allTimes.HasChangeTime() {
				return allTimes.ChangeTime()
			} else {
				return info.ModTime()
			}
		}()

		return direntryWithTimestamp{
			DirEntry:  e,
			timestamp: tim,
		}
	})

	return lo.Map(checkoutsWithTimestamps, func(x direntryWithTimestamp, _ int) CheckoutWithLabel {
		labelComponents := []string{x.Name()}

		labelComponents = append(labelComponents, x.timestamp.Format(time.RFC3339))

		comment, err := xattr.Get(filepath.Join(root.CheckoutsDir(), x.Name()), xdgcommonextendedattributes.Comment)
		if err == nil {
			labelComponents = append(labelComponents, string(comment))
		}

		return CheckoutWithLabel{
			CommitShort: x.Name(),
			Commit:      "", // needs to be empty for checkouts
			Label:       checkedOutIndicator(true) + strings.Join(labelComponents, " - "),
			Timestamp:   x.timestamp,
		}
	}), nil
}

func remoteNameParam() []string {
	return []string{ostreeRemoteName}
}
