// root filesystem storage and transport in OSTree
package ostree

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/joonas-fi/joonas-sys/pkg/xdgcommonextendedattributes"
	"github.com/pkg/xattr"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

const (
	ostreeBranchNameX8664 = "deploy/app/fi.joonas.os/x86_64/stable"
)

func Entrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ostree",
		Short: "OSTree + jsys management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "pull",
		Short: "Pull updates from joonas-sys",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			if _, err := userutil.RequireRoot(); err != nil {
				return err
			}

			logOutput := exec.CommandContext(ctx, "ostree", "pull", "fi.joonas.os", ostreeBranchNameX8664)
			logOutput.Stdout = os.Stdout
			logOutput.Stderr = os.Stderr
			if err := logOutput.Run(); err != nil {
				return err
			}

			fmt.Println("done. pro-tip:")
			fmt.Println("  $ jsys ostree log")
			fmt.Println("  $ jsys ostree checkout <commit>")

			return nil
		}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "log",
		Short: "Show commits",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			logOutput := exec.CommandContext(ctx, "ostree", "log", ostreeBranchNameX8664)
			logOutput.Stdout = os.Stdout
			logOutput.Stderr = os.Stderr
			return logOutput.Run()
		}),
	})

	cmd.AddCommand(commitEntrypoint())
	cmd.AddCommand(pushEntrypoint())

	cmd.AddCommand(&cobra.Command{
		Use:   "checkout <commit>",
		Short: "Checks out a root filesystem from a commit",
		Args:  cobra.ExactArgs(1),
		Run: cli.WrapRun(func(ctx context.Context, args []string) error {
			return checkoutRootFS(ctx, args[0])
		}),
	})

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

			// commit writes the revision to stdout
			revisionOutput := &bytes.Buffer{}

			commitOutput := exec.CommandContext(ctx, "ostree", "commit", "--branch="+ostreeBranchNameX8664, "--subject="+subject, common.BuildTreeLocation)
			commitOutput.Stdout = io.MultiWriter(os.Stdout, revisionOutput)
			commitOutput.Stderr = os.Stderr
			if err := commitOutput.Run(); err != nil {
				return err
			}

			commidID := revisionOutput.String()

			if checkout {
				return checkoutRootFS(ctx, commidID)
			} else {
				slog.Info("committed", "commidID", commidID)
				fmt.Printf("pro-tip: run $ %s ostree checkout %s\n", os.Args[0], commidID)
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

	commitShort := commit[0:7] // 7 hexits is unique enough

	checkoutPath := filelocations.Sysroot.Checkout(commitShort)

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

	showOutput, err := exec.CommandContext(ctx, "ostree", "show", commit).Output()
	if err != nil {
		return err
	}

	showCommentMatch := showCommentRe.FindStringSubmatch(string(showOutput))
	commitMessage := func() string {
		if showCommentMatch != nil {
			return showCommentMatch[1]
		} else {
			return ""
		}
	}()

	if err := xattr.Set(checkoutPath, xdgcommonextendedattributes.Comment, []byte(commitMessage)); err != nil {
		return err
	}

	slog.Info("checked out to", "checkoutPath", checkoutPath)

	fmt.Printf("pro-tip:\n  $ %s test-in-vm\nOR\n  $ %s flash efi\n", os.Args[0], os.Args[0])

	return nil
}

var showCommentRe = regexp.MustCompile(`\n    (.+)`)

type CheckoutWithLabel struct {
	Dir       string // "ae39405"
	Label     string // "ae39405 - 2023-05-28T13:06:07+03:00 - fix virtio-fsd"
	Timestamp time.Time
}

func GetCheckoutsSortedByDate(root filelocations.Root) ([]CheckoutWithLabel, error) {
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

	sort.Slice(checkoutsWithTimestamps, func(i, j int) bool {
		// newest to oldest
		return checkoutsWithTimestamps[i].timestamp.After(checkoutsWithTimestamps[j].timestamp)
	})

	return lo.Map(checkoutsWithTimestamps, func(x direntryWithTimestamp, _ int) CheckoutWithLabel {
		labelComponents := []string{x.Name()}

		labelComponents = append(labelComponents, x.timestamp.Format(time.RFC3339))

		comment, err := xattr.Get(filepath.Join(root.CheckoutsDir(), x.Name()), xdgcommonextendedattributes.Comment)
		if err == nil {
			labelComponents = append(labelComponents, string(comment))
		}

		return CheckoutWithLabel{
			Dir:       x.Name(),
			Label:     strings.Join(labelComponents, " - "),
			Timestamp: x.timestamp,
		}
	}), nil
}
