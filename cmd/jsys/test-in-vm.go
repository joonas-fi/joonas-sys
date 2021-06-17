package main

// Utility for testing a system (that exists either on a partition or in-RAM) in a VM

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

const (
	cowNameEsp  = "misc/vm-test-disks/esp.qcow2"
	cowNameRoot = "misc/vm-test-disks/root-ro.qcow2"
)

func testInVmEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "test-in-vm [system]",
		Short: "Tests systree in a VM",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(testInVm(
				osutil.CancelOnInterruptOrTerminate(nil),
				args[0]))
		},
	}
}

func testInVm(ctx context.Context, sysLabel string) error {
	if err := requireRoot(); err != nil {
		return err
	}

	sys, err := getSystemNotCurrent(sysLabel)
	if err != nil {
		return err
	}

	volatilePersistPartition, err := createEmptyRamBackedPersistPartition(sys)
	if err != nil {
		return err
	}

	// there is no way in QEMU to mark drives as readonly (apparently IDE doesn't have the concept),
	// so the next best thing we can do is create copy-on-write layer whose changes we discard

	if err := qemuCreatePseudoReadonlyDisk(sys.systemDevice, cowNameRoot); err != nil {
		return fmt.Errorf("qemuCreatePseudoReadonlyDisk %s: %w", cowNameRoot, err)
	}

	if err := qemuCreatePseudoReadonlyDisk(sys.espDevice, cowNameEsp); err != nil {
		return fmt.Errorf("qemuCreatePseudoReadonlyDisk %s: %w", cowNameEsp, err)
	}

	// the various OVMF_VARS files decide which system we'll boot automatically (UEFI vars remembering
	// last selected boot option)
	uefiVars := fmt.Sprintf("misc/uefi-files/OVMF_VARS-boot-%s.fd", sys.lieAboutLabelIfVirtualMachine())

	// RNG device supposedly speeds up Ubuntu boot

	vm := exec.CommandContext(ctx, "qemu-system-x86_64",
		"-machine", "type=q35,accel=kvm",
		"-drive", "file="+cowNameEsp,
		"-drive", "file="+cowNameRoot,
		"-drive", "format=raw,file="+volatilePersistPartition,
		"-drive", "if=pflash,format=raw,unit=0,readonly,file=misc/uefi-files/OVMF_CODE-pure-efi.fd",
		"-drive", "if=pflash,format=raw,unit=1,readonly,file="+uefiVars,
		"-m", "4G",
		"-smp", "4",
	)
	vm.Stdout = os.Stdout
	vm.Stderr = os.Stderr

	return vm.Run()
}

// requires root
func qemuCreatePseudoReadonlyDisk(realDevice string, cowFile string) error {
	exists, err := osutil.Exists(realDevice)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("realDevice does not exist: %s", realDevice)
	}

	// remove existing, so we start with empty diff disk
	if err := removeIfExists(cowFile); err != nil {
		return err
	}

	return exec.Command(
		"qemu-img",
		"create",
		"-f", "qcow2",
		"-b", realDevice,
		cowFile).Run()
}

func createEmptyRamBackedPersistPartition(sys systemSpec) (string, error) {
	volatilePersistPartition := fmt.Sprintf("/dev/shm/%s-persist-volatile", sys.label)

	if err := removeIfExists(volatilePersistPartition); err != nil {
		return "", err
	}

	if err := createAndTruncateFile(volatilePersistPartition, 4*gb); err != nil {
		return "", err
	}

	if err := exec.Command("mkfs.ext4", "-L", "persist", volatilePersistPartition).Run(); err != nil {
		return "", err
	}

	if err := mount(volatilePersistPartition, tmpMountpointPersist); err != nil {
		return "", err
	}
	defer func() {
		if err := unmount(tmpMountpointPersist); err != nil {
			panic(err)
		}
	}()

	writeFile := func(path string, content string) error {
		pathInPersist := filepath.Join(tmpMountpointPersist, path)

		if err := os.MkdirAll(filepath.Dir(pathInPersist), 0775); err != nil {
			return err
		}

		if err := os.WriteFile(pathInPersist, []byte(content), 0660); err != nil {
			return fmt.Errorf("write %s: %v", path, err)
		}

		return nil
	}

	if err := writeFile("apps/SYSTEM_nobackup/active_sys_id", sys.lieAboutLabelIfVirtualMachine()); err != nil {
		return "", err
	}

	if err := writeFile("apps/SYSTEM_nobackup/hostname", "j-sys-test-vm"); err != nil {
		return "", err
	}

	if err := copyBackgroundFromCurrentSystemIfExistsTo(filepath.Join(tmpMountpointPersist, "apps/SYSTEM_nobackup/background.png")); err != nil {
		return "", err
	}

	_ = os.Chmod("apps/SYSTEM_nobackup/background.png", 0661)

	if err := os.MkdirAll(filepath.Join(tmpMountpointPersist, "apps/docker/data_nobackup"), 0770); err != nil {
		return "", err
	}

	return volatilePersistPartition, nil
}
