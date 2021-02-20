package main

// Utility for testing a system (that exists either on a partition or in-RAM) in a VM

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

const (
	cowNameEsp  = "misc/vm-test-disks/esp.qcow2"
	cowNameRoot = "misc/vm-test-disks/root-ro.qcow2"
)

func testInVmEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "test-in-vm",
		Short: "Tests in-RAM systree in a VM",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(testInVm(
				osutil.CancelOnInterruptOrTerminate(nil),
				sysInRam))
		},
	}
}

func testInVm(ctx context.Context, sys systemSpec) error {
	if err := requireRoot(); err != nil {
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
	uefiVars := fmt.Sprintf("misc/uefi-files/OVMF_VARS-boot-system-%s.fd", sys.sysId)

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
	volatilePersistPartition := fmt.Sprintf("/dev/shm/sys-%s-persist-volatile", sys.sysId)

	if err := removeIfExists(volatilePersistPartition); err != nil {
		return "", err
	}

	// Truncate() needs file to exist
	if err := createEmptyFile(volatilePersistPartition); err != nil {
		return "", err
	}

	// creates sparse file (i.e. will only take space for blocks that are actually used)
	if err := os.Truncate(volatilePersistPartition, 4*gb); err != nil {
		return "", err
	}

	if err := exec.Command("mkfs.ext4", "-L", "persist", volatilePersistPartition).Run(); err != nil {
		return "", err
	}

	return volatilePersistPartition, nil
}
