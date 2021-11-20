package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

func restartPrepareEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "restart-prepare [system]",
		Short: "Prepare quick kexec-based restart into new active system",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func(sysLabel string) error {
				if err := requireRoot(); err != nil {
					return err
				}

				system, err := getSystemNotCurrent(sysLabel)
				if err != nil {
					return err
				}

				if err := os.WriteFile("/persist/apps/SYSTEM_nobackup/active_sys_id", []byte(sysLabel), 0770); err != nil {
					return err
				}

				// TODO: grab these from ESP instead?
				if err := mountSystem(system); err != nil {
					return fmt.Errorf("mountSystem: %w", err)
				}

				newKernelLocation := filepath.Join(tmpMountpointSystem, "/boot/vmlinuz")
				newInitrdLocation := filepath.Join(tmpMountpointSystem, "/boot/initrd.img")

				kernelCommandLine := fmt.Sprintf("root=LABEL=%s ro", system.label)

				// loads new image/initrd/cmdline into the old kernel for kexec'ing later, not just yet, but soon.
				kexecOutput, err := exec.Command("kexec", "--load",
					"--initrd="+newInitrdLocation,
					"--command-line="+kernelCommandLine,
					newKernelLocation,
				).CombinedOutput()

				if err != nil {
					return fmt.Errorf("kexec: %w: %s", err, kexecOutput)
				}

				if err := unmount(tmpMountpointSystem); err != nil {
					return fmt.Errorf("unmount system: %w", err)
				}

				fmt.Printf(
					"succeeded. to reboot, issue (with sudo):\n    $ systemctl kexec\nremember to enter %s!",
					system.label)

				return nil
			}(args[0]))
		},
	}
}

func restartPrepareCurrentEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "restart-current-prepare",
		Short: "Prepare quick kexec-based restart into same system",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func() error {
				if err := requireRoot(); err != nil {
					return err
				}

				systemLabel, err := readRunningSystemId()
				if err != nil {
					return err
				}

				kexecOutput, err := exec.Command("kexec", "--load",
					"--initrd=/boot/initrd.img",
					"--reuse-cmdline",
					"/boot/vmlinuz",
				).CombinedOutput()
				if err != nil {
					return fmt.Errorf("kexec: %w: %s", err, kexecOutput)
				}

				fmt.Printf(
					"succeeded. to reboot, issue (with sudo):\n    $ systemctl kexec\nremember to enter %s!\n",
					systemLabel)

				return nil
			}())
		},
	}
}
