package main

// Flashes systree to a system partition (and makes corresponding changes to ESP)

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/function61/gokit/os/osutil"
	"github.com/prometheus/procfs"
	"github.com/spf13/cobra"
)

const (
	tmpMountpointEsp     = "/tmp/jsys-esp"
	tmpMountpointSystem  = "/tmp/jsys-system"
	tmpMountpointPersist = "/tmp/jsys-persist"
)

func flashEntrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flash [system]",
		Short: "Flashes systree to storage",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(flash(
				osutil.CancelOnInterruptOrTerminate(nil),
				args[0]))
		},
	}

	return cmd
}

func flash(ctx context.Context, sysLabel string) error {
	if err := requireRoot(); err != nil {
		return err
	}

	system, err := getSystemNotCurrent(sysLabel)
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
		return fmt.Errorf(
			"safety: bailing out because diff tree exists!\n(safely do this first:) $ rm -rf %s",
			system.diffPath())
	}

	if !exists {
		if system.systemDeviceCanCreateIfNotFound {
			log.Println("RAM device doesn't exist - creating & formatting")

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
			log.Printf("unmount system: %v", err)
		}
	}()

	copySystreeFrom := func() string {
		if remote := os.Getenv("REMOTE"); remote != "" {
			return remote
		} else {
			return treeLocation + "/"
		}
	}()

	if err := copySystree(copySystreeFrom, system); err != nil {
		return fmt.Errorf("copySystree: %w", err)
	}

	if system.espDeviceCanCreateIfNotFound {
		espDeviceExists, err := osutil.Exists(system.espDevice)
		if err != nil {
			return err
		}

		if !espDeviceExists {
			log.Println("ESP doesn't exist and we are allowed to create it - creating")

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
			log.Printf("unmount ESP: %v", err)
		}
	}()

	if err := copyKernelAndInitrdToEsp(system); err != nil {
		return fmt.Errorf("copyKernelAndInitrdToEsp: %w", err)
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
	// TODO: use syscall
	return exec.Command("umount", mountpoint).Run()
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

	if err := copyFile(sys("/boot/vmlinuz"), uefiAppDir("/vmlinuz")); err != nil {
		return err
	}

	if err := copyFile(sys("/boot/initrd.img"), uefiAppDir("/initrd.img")); err != nil {
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

	// can't use -a flag because it would try to copy permissions, which FAT doesn't support
	if err := exec.Command("rsync",
		"-h",
		"--recursive",
		"--delete",
		"--exclude=system_*",
		"--exclude=background.png",
		"misc/esp/EFI/",
		filepath.Join(tmpMountpointEsp, "EFI"),
	).Run(); err != nil {
		return fmt.Errorf("rsync: %v", err)
	}

	return nil
}

func mountIfNeeded(device string, mountpoint string) error {
	if is, err := isMounted(mountpoint); is || err != nil { // already mounted?
		if err != nil {
			return err
		} else {
			log.Printf("already mounted: %s", mountpoint)
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
	backgroundFromCurrentSystem := "/persist/apps/SYSTEM_nobackup/background.png"

	backgroundExists, err := osutil.Exists(backgroundFromCurrentSystem)
	if err != nil {
		return err
	}

	if backgroundExists {
		if err := copyFile(backgroundFromCurrentSystem, to); err != nil {
			return err
		}
	}

	return nil
}
