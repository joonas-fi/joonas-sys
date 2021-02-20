package main

// ESP partition creation

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

func espEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "esp-create-in-ram",
		Short: "Creates ESP partition",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(espCreate(
				osutil.CancelOnInterruptOrTerminate(nil),
				sysInRam))
		},
	}
}

const (
	mb = 1024 * 1024
	gb = 1024 * mb
)

func espCreate(ctx context.Context, sys systemSpec) error {
	if err := requireRoot(); err != nil {
		return err
	}

	exists, err := osutil.Exists(sys.espDevice)
	if err != nil {
		return err
	}

	if !exists {
		log.Println("ESP device doesn't exist - creating")

		if err := createEmptyInRamEspDevice(ctx, sys); err != nil {
			return err
		}
	}

	if err := mountEsp(sys); err != nil {
		return fmt.Errorf("mountEsp: %w", err)
	}

	if err := copyEspTemplateToEsp(ctx); err != nil {
		return err
	}

	return unmount(tmpMountpointEsp)
}

// pretty much summed up by:
//   $ cp -r misc/esp/ /tmp/jsys-esp
func copyEspTemplateToEsp(ctx context.Context) error {
	// can't use -a flag because it would try to copy permissions, which FAT doesn't support
	rsync := exec.CommandContext(ctx, "rsync",
		"--recursive",
		"--times",
		"--human-readable",
		"--info=progress2",
		"misc/esp/",
		tmpMountpointEsp,
	)
	rsync.Stdout = os.Stdout
	rsync.Stderr = os.Stderr

	return rsync.Run()
}

func createEmptyInRamEspDevice(ctx context.Context, system systemSpec) error {
	volatileEspBackingFile := "/dev/shm/esp-staging.img"

	// file needs to exist before we can call truncate
	if err := createEmptyFile(volatileEspBackingFile); err != nil {
		return err
	}

	if err := os.Truncate(volatileEspBackingFile, 512*mb); err != nil {
		return err
	}

	// usually "ESP"  (might be "ESP-VM" when testing in a VM)
	espFilesystemLabel, err := system.espDeviceLabel()
	if err != nil {
		return err
	}

	// TODO: use pregenerated template to be more portable?
	if err := exec.Command("mkfs.fat", "-F32", "-n", espFilesystemLabel, volatileEspBackingFile).Run(); err != nil {
		return err
	}

	// because we named our filesystem that we just now made, after this "$ losetup" call there should be
	// symlink "/dev/disk/by-label/ESP-VM" waiting for us ...
	if err := exec.Command("losetup", "--find", "--partscan", volatileEspBackingFile).Run(); err != nil {
		return err
	}

	// ... but it is not synchronous
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return waitForFileAvailable(ctx, system.espDevice)
}
