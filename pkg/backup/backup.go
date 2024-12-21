// Tools for backing up the system
package backup

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/function61/gokit/sliceutil"
	"github.com/function61/gokit/time/timeutil"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/joonas-fi/joonas-sys/pkg/xdgcommonextendedattributes"
	"github.com/pkg/xattr"
	"github.com/spf13/cobra"
)

var (
	appPath            = filelocations.Sysroot.App("backup-rsync")
	backupExcludesFile = filepath.Join(appPath, "excludes.conf")
	backupPasswordFile = filepath.Join(appPath, "password")
)

func Entrypoint() *cobra.Command {
	dryRun := false
	refreshExcludes := false

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup files using rsync",
		Args:  cobra.NoArgs,
		Run: cli.RunnerNoArgs(func(ctx context.Context, logger *log.Logger) error {
			if refreshExcludes {
				if err := backupExcludedDirs(ctx, logger); err != nil {
					return err
				}
			}

			return backup(ctx, dryRun)
		}),
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "", dryRun, "Dry run")
	cmd.Flags().BoolVarP(&refreshExcludes, "refresh-excludes", "", refreshExcludes, "Refresh excludes list first")

	cmd.AddCommand(excludedDirsEntrypoint())

	return cmd
}

func backup(ctx context.Context, dryRun bool) error {
	password, err := os.ReadFile(backupPasswordFile)
	if err != nil {
		return err
	}

	// if persist hierarchy would be empty due to an error (say, HDD crash),
	// it could also trigger removing backups. this would make us pretty sad pandas.
	sane, err := osutil.Exists(filelocations.Sysroot.App("work"))
	if err != nil {
		return err
	}

	if !sane {
		return errors.New("Sanity check failed; it would be dangerous to continue")
	}

	excludeExists, err := osutil.Exists(backupExcludesFile)
	if err != nil || !excludeExists {
		return fmt.Errorf("the exclude file must exist: %s", backupExcludesFile)
	}

	if err := os.Setenv("RSYNC_PASSWORD", string(password)); err != nil {
		return err
	}

	args := []string{"docker", "run", "--rm",
		"--env=RSYNC_PASSWORD",
		"--user=0:0", // root (the rsync image defaults to unprivileged user)
		"-v", backupExcludesFile + ":/excludes.conf:ro",
		"-v", fmt.Sprintf("%[1]s:%[1]s:ro", filelocations.Sysroot.Root()),
		"joonas/rsync",
		"-av",
		"--delete",
		"--exclude-from=/excludes.conf"}

	if dryRun {
		args = append(args, "--dry-run")
	}

	args = append(args,
		filelocations.Sysroot.Root()+"/",             // source
		"rsync://joonas@192.168.1.105/volume/backup", // destination
	)

	rsync := exec.CommandContext(ctx, args[0], args[1:]...)
	rsync.Stdout = os.Stdout
	rsync.Stderr = os.Stderr

	// TODO: realtime progress updates?
	if output, err := exec.CommandContext(ctx, "notify-send", "Backup starting").CombinedOutput(); err != nil {
		log.Printf("notify-send: %v: %s", err, string(output))
	}

	started := time.Now()

	if err := rsync.Run(); err != nil {
		return err
	}

	notifyMsg := fmt.Sprintf("Backup completed in %s", timeutil.HumanizeDuration(time.Since(started)))

	if output, err := exec.CommandContext(ctx, "notify-send", notifyMsg).CombinedOutput(); err != nil {
		log.Printf("notify-send: %v: %s", err, string(output))
	}

	return nil
}

func excludedDirsEntrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "excluded-dirs",
		Short: "Scans the tree-to-backup for excluded dirs & writes them to excludes.conf file",
		Args:  cobra.NoArgs,
		Run:   cli.RunnerNoArgs(backupExcludedDirs),
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "add [path]",
		Short: "Add a directory (or a file) to backup excludes",
		Args:  cobra.ExactArgs(1),
		Run: cli.Runner(func(_ context.Context, args []string, _ *log.Logger) error {
			return backupExcludeObject(args[0])
		}),
	})

	return cmd
}

func backupExcludedDirs(ctx context.Context, _ *log.Logger) error {
	// f.ex. reading lost+found xattrs requires root
	if _, err := userutil.RequireRoot(); err != nil {
		return err
	}

	// start scanning under this tree. don't follow symlinks.
	excluded, files, dirs, err := discoverExcludedDirs(filelocations.Sysroot.Root())
	if err != nil {
		return err
	}

	for _, item := range excluded {
		fmt.Printf("- exclude %s\n", item)
	}

	excludesConf := strings.Join(excluded, "\n") + "\n"

	if err := osutil.WriteFileAtomicFromReader(backupExcludesFile, strings.NewReader(excludesConf)); err != nil {
		return err
	}

	fmt.Printf("files = %d dirs = %d, wrote %s\n", files, dirs, backupExcludesFile)

	return nil
}

func backupExcludeObject(path string) error {
	return xattr.Set(path, xdgcommonextendedattributes.RobotsBackup, []byte("false"))
}

// to opt out a dir, run:
//
//	$ setfattr -n user.xdg.robots.backup -v false /dir
func IsExcluded(path string) (bool, error) {
	// we can' directly "get" the xattr because if it doesn't exist it errors with an error whose
	// concrete type we can't assert (was it "not exist" error or an actual error?)
	attrs, err := xattr.LList(path)
	if err != nil {
		return false, err
	}

	// xattr not present => no need to dig further
	if !sliceutil.ContainsString(attrs, xdgcommonextendedattributes.RobotsBackup) {
		return false, nil
	}

	value, err := xattr.Get(path, xdgcommonextendedattributes.RobotsBackup)
	if err != nil {
		return false, nil
	}

	return string(value) == "false", nil
}

// discovers directories / files from a file tree with XDG-recommended backup exclusion xattr
func discoverExcludedDirs(root string) ([]string, int, int, error) {
	files := 0
	dirs := 0

	excluded := []string{}

	if err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rootRelative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirs++

			excludedFromBackup, err := IsExcluded(path)
			if err != nil {
				return err
			}

			if excludedFromBackup {
				excluded = append(excluded, "/"+rootRelative)
				return fs.SkipDir // no sense in descending there, as the whole sub-tree is blocked from backing up
			}

			// TODO: remove old mechanism
			if strings.Contains(rootRelative, "_nobackup") {
				return fmt.Errorf("%s contains _nobackup but not excluded via XDG xattr", rootRelative)
			}

			return nil
		} else {
			files++
			return nil
		}
	}); err != nil {
		return nil, 0, 0, err
	}

	sort.Strings(excluded) // so config file is stable

	return excluded, files, dirs, nil
}
