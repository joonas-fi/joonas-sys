package main

// Flashes systree to a system partition (and makes corresponding changes to ESP)

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/joonas-fi/joonas-sys/pkg/ostree"
	"github.com/prometheus/procfs"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

const (
	tmpMountpointEsp     = "/tmp/jsys-esp"
	tmpMountpointSystem  = "/tmp/jsys-system"
	tmpMountpointPersist = "/tmp/jsys-persist"
)

func flashEntrypoint() *cobra.Command {
	ignoreWarnings := false
	remote := false
	autoRemove := false
	ostreeRef := ""

	cmd := &cobra.Command{
		Use:   "flash [system]",
		Short: "Flashes systree to storage",
		Args:  cobra.ExactArgs(1),
		Run: cli.WrapRun(func(ctx context.Context, args []string) error {
			return flash(
				ctx,
				args[0],
				ostreeRef,
				ignoreWarnings,
				remote,
				autoRemove)
		}),
	}

	cmd.Flags().BoolVarP(&ignoreWarnings, "ignore-warnings", "", ignoreWarnings, "Ignore any warnings")
	cmd.Flags().BoolVarP(&remote, "remote", "", remote, "Use known remote (192.168.1.104)")
	cmd.Flags().BoolVarP(&autoRemove, "auto-remove", "", autoRemove, "Automatically remove previous diff")
	cmd.Flags().StringVarP(&ostreeRef, "ostree", "", ostreeRef, "OSTree ref to checkout")

	cmd.AddCommand(flashEFIEntrypoint())

	return cmd
}

func flashEFIEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "efi",
		Short: "Flash EFI boot partition with target sysid",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			espMounted, err := osutil.Exists("/boot/efi/EFI/")
			if err != nil || !espMounted {
				return errors.New("/boot/efi not mounted")
			}

			sysrootCheckouts, err := ostree.GetCheckoutsSortedByDate(filelocations.Sysroot)
			if err != nil {
				return err
			}
			idx, _, err := promptUISelect("Version", lo.Map(sysrootCheckouts, func(x ostree.CheckoutWithLabel, _ int) string { return x.Label }))
			if err != nil {
				return err
			}

			sysID := sysrootCheckouts[idx].Dir

			// create diff dir (system is unbootable without this)
			if err := os.MkdirAll(filelocations.Sysroot.Diff(sysID), 0755); err != nil {
				return err
			}

			// TODO: discover by https://uapi-group.org/specifications/specs/discoverable_partitions_specification/
			cmdline := append(createKernelCmdline(sysID), "root=LABEL=persist")

			vol1 := fmt.Sprintf("--volume=%s:/sysroot", filelocations.Sysroot.Checkout(sysID))
			vol2 := "--volume=/tmp/ukifybuild:/workspace"

			ukifyBuild := exec.CommandContext(ctx, "docker", "run", "--rm", "-t", vol1, vol2, "ghcr.io/joonas-fi/joonas-sys-ukify:latest", "build",
				"--linux=/sysroot/boot/vmlinuz",
				"--initrd=/sysroot/boot/initrd.img",
				"--cmdline="+strings.Join(cmdline, " "),
				"--output=/workspace/BOOTx64.efi")
			if output, err := ukifyBuild.CombinedOutput(); err != nil {
				return fmt.Errorf("ukify: %w: %s", err, string(output))
			}

			fmt.Println("pro-tip: (NOTE: take backup of target first)")
			fmt.Println("  $ cp /tmp/ukifybuild/BOOTx64.efi /boot/efi/EFI/BOOT/BOOTx64.efi")

			return nil
		}),
	}
}

func flash(ctx context.Context, sysLabel string, ostreeRef string, ignoreWarnings bool, remote bool, autoRemove bool) error {
	if _, err := userutil.RequireRoot(); err != nil {
		return err
	}

	if remote {
		// dirty implementation
		_ = os.Setenv("REMOTE", "rsync://192.168.1.104/")
	}

	system, err := func() (systemSpec, error) {
		// ignore would be needed in preinstallation environment where we don't have persist hierarchy available
		// (and by extension, active_sys_id)
		if ignoreWarnings {
			return getSystemNoEditCheck(sysLabel)
		} else {
			return getSystemNotCurrent(sysLabel)
		}
	}()
	if err != nil {
		return err
	}

	exists, err := osutil.Exists(system.systemDevice)
	if err != nil {
		return err
	}

	diffTreeExists, err := osutil.Exists(system.diffPath())
	if err != nil {
		return err
	}

	if diffTreeExists {
		if autoRemove {
			if err := os.RemoveAll(system.diffPath()); err != nil {
				return fmt.Errorf("automatic remove of diff: %w", err)
			}
		} else {
			if !ignoreWarnings {
				return fmt.Errorf(
					"safety: bailing out because diff tree exists!\n(safely do this first:) $ rm -rf %s",
					system.diffPath())
			}
		}
	}

	if !exists {
		if system.systemDeviceCanCreateIfNotFound {
			slog.Info("RAM device doesn't exist - creating & formatting")

			if err := createAndTruncateFile(system.systemDevice, 10*gb); err != nil {
				return err
			}

			if err := exec.Command("mkfs.ext4", "-L", system.lieAboutLabelIfVirtualMachine(), system.systemDevice).Run(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("system partition not found: %s", system.systemDevice)
		}
	}

	if err := mountSystem(system); err != nil {
		return fmt.Errorf("mountSystem: %w", err)
	}
	defer func() {
		if err := unmount(tmpMountpointSystem); err != nil {
			slog.Error("unmount system", "err", err)
		}
	}()

	if ostreeRef != "" {
		checkout := exec.CommandContext(ctx, "ostree", "checkout", "--union", "--force-copy", ostreeRef, tmpMountpointSystem)
		checkout.Stdout = os.Stdout
		checkout.Stderr = os.Stderr

		if err := checkout.Run(); err != nil {
			return fmt.Errorf("ostree checkout: %w", err)
		}
	} else {
		copySystreeFrom := func() string {
			if remote := os.Getenv("REMOTE"); remote != "" {
				if !strings.HasSuffix(remote, "/") { // made this accident once
					panic("remote must end in slash")
				}

				return remote + "jsys/"
			} else {
				return common.BuildTreeLocation + "/"
			}
		}()

		if err := copySystree(copySystreeFrom, system); err != nil {
			return fmt.Errorf("copySystree: %w", err)
		}
	}

	if system.espDeviceCanCreateIfNotFound {
		espDeviceExists, err := osutil.Exists(system.espDevice)
		if err != nil {
			return err
		}

		if !espDeviceExists {
			slog.Info("ESP doesn't exist and we are allowed to create it - creating")

			if err := espFormatInternal(ctx, system); err != nil {
				return err
			}
		}
	}

	if err := mountEsp(system); err != nil {
		return fmt.Errorf("mountEsp: %w", err)
	}
	defer func() {
		if err := unmount(tmpMountpointEsp); err != nil {
			slog.Error("unmount ESP", "err", err)
		}
	}()

	if err := copyKernelAndInitrdToEsp(system); err != nil {
		return fmt.Errorf("copyKernelAndInitrdToEsp: %w", err)
	}

	syscall.Sync() // no return value

	if sysLabel != "in-ram" { // no sense for in-ram flash
		// pro-tip
		fmt.Printf(
			"flashing complete. to prepare boot into the new system:\n  $ %s restart-prepare %s\n",
			os.Args[0],
			sysLabel)
	}

	return nil
}

func mountSystem(paths systemSpec) error {
	return mountIfNeeded(paths.systemDevice, tmpMountpointSystem)
}

func mountEsp(paths systemSpec) error {
	return mountIfNeeded(paths.espDevice, tmpMountpointEsp)
}

func unmount(mountpoint string) error {
	if err := syscall.Unmount(mountpoint, 0); err != nil {
		return fmt.Errorf("unmount: %w", err)
	}

	return nil
}

func copySystree(from string, paths systemSpec) error {
	rsync := exec.Command(
		"rsync",
		"-ah",
		"--delete",
		"--info=progress2",
		from,
		tmpMountpointSystem,
	)
	rsync.Stdout = os.Stdout
	rsync.Stderr = os.Stderr

	return rsync.Run()
}

func copyKernelAndInitrdToEsp(system systemSpec) error {
	sys := func(file string) string { // shorthand
		return filepath.Join(tmpMountpointSystem, file)
	}
	uefiAppDir := func(file string) string { // shorthand
		return filepath.Join(tmpMountpointEsp, "EFI", system.lieAboutLabelIfVirtualMachine(), file)
	}

	dummyPerms := osutil.FileMode(osutil.OwnerRWX, osutil.GroupRWX, osutil.OtherNone) // ESP partition doesn't support perms
	if err := os.MkdirAll(uefiAppDir(""), dummyPerms); err != nil {
		return err
	}

	if err := osutil.CopyFile(sys("/boot/vmlinuz"), uefiAppDir("/vmlinuz")); err != nil {
		return err
	}

	if err := osutil.CopyFile(sys("/boot/initrd.img"), uefiAppDir("/initrd.img")); err != nil {
		return err
	}

	if err := copyBackgroundFromCurrentSystemIfExistsTo(filepath.Join(tmpMountpointEsp, "EFI", "background.png")); err != nil {
		return err
	}

	/* Production EFI dir will look like this:

	EFI
	├── refind
	├── system_a
	├── system_b
	├── background.png
	└── tools

	However our EFI template tree doesn't contain system_ + ("a" | "b") or background so we've to
	exclude them from rsync so they won't get deleted (b/c --delete flag)
	*/
	copyESPFrom := func() string {
		if remote := os.Getenv("REMOTE"); remote != "" {
			return remote + "EFI/"
		} else {
			return "misc/esp/EFI/"
		}
	}()

	// can't use -a flag because it would try to copy permissions, which FAT doesn't support
	if err := exec.Command("rsync",
		"-h",
		"--recursive",
		"--delete",
		"--exclude=system_*",
		"--exclude=background.png",
		copyESPFrom,
		filepath.Join(tmpMountpointEsp, "EFI"),
	).Run(); err != nil {
		return fmt.Errorf("ESP rsync: %v", err)
	}

	return nil
}

func mountIfNeeded(device string, mountpoint string) error {
	if is, err := isMounted(mountpoint); is || err != nil { // already mounted?
		if err != nil {
			return err
		} else {
			slog.Info("already mounted", "mountpoint", mountpoint)
			return nil
		}
	}

	return mount(device, mountpoint)
}

func mount(device string, mountpoint string) error {
	if exists, err := osutil.Exists(device); !exists || err != nil {
		return fmt.Errorf("mount source %s does not exist: %w", device, err)
	}

	if err := os.MkdirAll(mountpoint, osutil.FileMode(osutil.OwnerRWX, osutil.GroupRWX, osutil.OtherNone)); err != nil {
		return err
	}

	// TODO: does it work without -o loop?
	return exec.Command("mount", device, mountpoint).Run()
}

func isMounted(mountpoint string) (bool, error) {
	// FIXME: procfs contains bug: https://github.com/prometheus/node_exporter/issues/1672
	// "couldn't find enough fields in mount string: 1570 29 0:25 /snapd/ns /run/snapd/ns rw,nosuid,nodev,noexec,relatime - tmpfs tmpfs rw,size=3081272k,mode=755"
	if true {
		// this is a bad and dangerous hack
		// status 0 if mounted
		err := exec.Command("mountpoint", "-q", mountpoint).Run()
		return err == nil, nil
	}

	mounts, err := procfs.GetMounts()
	if err != nil {
		return false, err
	}

	for _, mount := range mounts {
		if mount.MountPoint == mountpoint {
			return true, nil
		}
	}

	return false, nil
}

func copyBackgroundFromCurrentSystemIfExistsTo(to string) error {
	backgroundFromCurrentSystem := filepath.Join(filelocations.Sysroot.App(common.AppSYSTEM), "background.png")

	backgroundExists, err := osutil.Exists(backgroundFromCurrentSystem)
	if err != nil {
		return err
	}

	if backgroundExists {
		if err := osutil.CopyFile(backgroundFromCurrentSystem, to); err != nil {
			return err
		}
	}

	return nil
}

func createKernelCmdline(sysID string) []string {
	return []string{"sysid=" + sysID, "rw"}
}
