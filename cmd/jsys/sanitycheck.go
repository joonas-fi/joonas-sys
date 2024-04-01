package main

import (
	"context"
	"fmt"
	"os"

	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/joonas-fi/joonas-sys/pkg/backup"
	"github.com/spf13/cobra"
)

func sanityCheckEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "sanitycheck",
		Short: "Check that important files are where the should be",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(sanityCheck(
				osutil.CancelOnInterruptOrTerminate(nil)))
		},
	}
}

var (
	shouldBeSymlink = []string{
		"/sysroot/apps/SYSTEM_nobackup", // backwards compat
		"/persist/apps",                 // backwards compat
		"/persist/work",                 // backwards compat
	}
	filesThatShouldNotExist = []string{
		"/sysroot/apps/docker/data_nobackup", // deprecated
		"/sysroot/apps/docker/cli-plugins",   // deprecated
		"/sysroot/apps/SYSTEM/cpu_temp",      // deprecated
		"/sysroot/apps/SYSTEM/active_sys_id", // deprecated
		"/sysroot/apps/ostree",               // deprecated
	}
)

func sanityCheck(ctx context.Context) error {
	// some places of access, like /sysroot/lost+found, require root
	if _, err := userutil.RequireRoot(); err != nil {
		return err
	}

	filesThatShouldExist := []string{
		"/sysroot/apps/SYSTEM/backlight-state",
		"/sysroot/apps/SYSTEM/rfkill-state",
		"/sysroot/apps/SYSTEM/hostname",
		"/sysroot/apps/SYSTEM/machine-id", // $ dbus-uuidgen --ensure=/sysroot/apps/SYSTEM/machine-id
		"/sysroot/apps/SYSTEM/background.png",
		"/sysroot/apps/SYSTEM/lowdiskspace-check-rules/root",
		"/sysroot/apps/Desktop",
		"/sysroot/apps/ssh-server/ssh_host_ed25519_key",
		"/sysroot/apps/varasto/varasto-work",
		"/sysroot/apps/zoxide", // zoxide directory history
		"/sysroot/apps/docker/config",
		"/sysroot/apps/docker/data",
		"/sysroot/apps/flatpak",
		"/sysroot/apps/flatpak-appdata",
		"/sysroot/apps/netplan",
		"/sysroot/apps/mcfly",
		"/sysroot/apps/git config", // usually symlink to Varasto, sometimes may be local "fork" (custom Git author for client work on client laptop)
		"/sysroot/swapfile",
		"/dev/cpu_temp",
		"/dev/kvm", // assert KVM support enabled. details: ticket #38
	}

	errorCount := 0
	errorCountOfWhichNotExcludedFromBackup := 0

	emitError := func(msg string, subCounter *int) {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)

		errorCount++

		if subCounter != nil {
			*subCounter = (*subCounter) + 1
		}
	}

	for _, fileThatShouldExist := range filesThatShouldExist {
		exists, err := osutil.Exists(fileThatShouldExist)
		if err != nil {
			return err
		}

		if !exists {
			emitError(fmt.Sprintf("file missing: %s\n", fileThatShouldExist), nil)
		}
	}

	// at least lost+found requires root to access (but things like /sto don't work with root currently)
	for _, shouldExcludeFromBackup := range []string{
		"/sysroot/lost+found",
		"/sysroot/apps/SYSTEM",
		"/sysroot/apps/OS-checkout",
		"/sysroot/apps/OS-diff",
		"/sysroot/apps/OS-repo",
		"/sysroot/apps/docker/data",
		"/sysroot/apps/flatpak",
		"/sysroot/apps/flatpak-appdata",
	} {
		excludedFromBackup, err := backup.IsExcluded(shouldExcludeFromBackup)
		if err != nil {
			return err
		}

		if !excludedFromBackup {
			emitError(fmt.Sprintf("not excluded from backup: %s", shouldExcludeFromBackup), &errorCountOfWhichNotExcludedFromBackup)
		}
	}

	for _, shouldBeSymlinkItem := range shouldBeSymlink {
		isSymlink, err := func() (bool, error) {
			info, err := os.Lstat(shouldBeSymlinkItem)
			if err != nil {
				if os.IsNotExist(err) {
					return false, nil
				} else {
					return false, err
				}
			}

			return info.Mode()&os.ModeSymlink != 0, nil
		}()
		if err != nil {
			return err
		}

		if !isSymlink {
			emitError(fmt.Sprintf("not symlink: %s", shouldBeSymlinkItem), nil)
		}
	}
	for _, fileThatShouldNotExist := range filesThatShouldNotExist {
		exists, err := osutil.Exists(fileThatShouldNotExist)
		if err != nil {
			return err
		}

		if exists {
			emitError(fmt.Sprintf("should not exist: %s", fileThatShouldNotExist), nil)
		}
	}

	if errorCount > 0 {
		if errorCountOfWhichNotExcludedFromBackup > 0 {
			fmt.Println("  pro-tip: $ setfattr -n user.xdg.robots.backup -v false <dir>")
		}
		return fmt.Errorf("%d errors encountered", errorCount)
	} else {
		return nil
	}
}
