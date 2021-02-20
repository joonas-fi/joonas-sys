package main

// Flashes systree to a system partition (and makes corresponding changes to ESP)

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/function61/gokit/os/osutil"
	"github.com/prometheus/procfs"
	"github.com/spf13/cobra"
)

const (
	tmpMountpointEsp    = "/tmp/jsys-esp"
	tmpMountpointSystem = "/tmp/jsys-system"
)

var (
	sysInRam = systemSpec{
		sysId: "a", // in-RAM system is always known as system A

		systemDevice: "/dev/shm/joonas-os-ram-image",

		espDevice: "/dev/disk/by-label/ESP-VM",
	}
)

type systemSpec struct {
	sysId string

	systemDevice string

	espDevice string
}

func (s systemSpec) espDeviceLabel() (string, error) {
	diskByLabelPrefix := "/dev/disk/by-label/"

	if strings.HasPrefix(s.espDevice, diskByLabelPrefix) {
		return strings.TrimPrefix(s.espDevice, diskByLabelPrefix), nil
	}

	return "", fmt.Errorf(
		"ESP device does not start with '"+diskByLabelPrefix+"', cannot deduce label for %s",
		s.espDevice)
}

func flashEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "flash-to-ram",
		Short: "Flashes systree to RAM",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(flashToInRamDisk(
				osutil.CancelOnInterruptOrTerminate(nil),
				sysInRam))
		},
	}
}

func flashToInRamDisk(ctx context.Context, system systemSpec) error {
	if err := requireRoot(); err != nil {
		return err
	}

	if err := mountSystem(system); err != nil {
		return fmt.Errorf("mountSystem: %w", err)
	}

	if err := copySystree(system); err != nil {
		return fmt.Errorf("copySystree: %w", err)
	}

	if err := stampSysId(system); err != nil {
		return fmt.Errorf("stampSysId: %w", err)
	}

	if err := mountEsp(system); err != nil {
		return fmt.Errorf("mountEsp: %w", err)
	}

	if err := copyKernelAndInitrdToEsp(system); err != nil {
		return fmt.Errorf("copyKernelAndInitrdToEsp: %w", err)
	}

	if err := unmount(tmpMountpointEsp); err != nil {
		return fmt.Errorf("unmount ESP: %w", err)
	}

	if err := unmount(tmpMountpointSystem); err != nil {
		return fmt.Errorf("unmount system: %w", err)
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

func copySystree(paths systemSpec) error {
	rsync := exec.Command(
		"rsync",
		"-ah",
		"--delete",
		"--info=progress2",
		"/mnt/j-os-inmem-staging/",
		tmpMountpointSystem,
	)
	rsync.Stdout = os.Stdout
	rsync.Stderr = os.Stderr

	return rsync.Run()
}

// stamps the system partition with a /etc/sys-id file so it knows which system instance we booted
// TODO: since we're already passing a kernel label for partition name, could we deduce it from that?
func stampSysId(system systemSpec) error {
	return ioutil.WriteFile(
		filepath.Join(tmpMountpointSystem, "/etc/sys-id"),
		[]byte(system.sysId),
		osutil.FileMode(osutil.OwnerRW, osutil.GroupRW, osutil.OtherR))
}

func copyKernelAndInitrdToEsp(system systemSpec) error {
	sys := func(file string) string { // shorthand
		return filepath.Join(tmpMountpointSystem, file)
	}
	uefiAppDir := func(file string) string { // shorthand
		return filepath.Join(tmpMountpointEsp, "EFI", "system"+system.sysId, file)
	}

	if err := copyFile(sys("/boot/vmlinuz"), uefiAppDir("/vmlinuz")); err != nil {
		return err
	}

	if err := copyFile(sys("/boot/initrd.img"), uefiAppDir("/initrd.img")); err != nil {
		return err
	}

	return nil
}

func mountIfNeeded(device string, mountpoint string) error {
	if exists, err := osutil.Exists(device); !exists || err != nil {
		return fmt.Errorf("mount source %s does not exist: %w", device, err)
	}

	if err := os.MkdirAll(mountpoint, osutil.FileMode(osutil.OwnerRWX, osutil.GroupRWX, osutil.OtherNone)); err != nil {
		return err
	}

	if is, err := isMounted(mountpoint); is || err != nil { // already mounted?
		if err != nil {
			return err
		} else {
			log.Printf("already mounted: %s", mountpoint)
			return nil
		}
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
