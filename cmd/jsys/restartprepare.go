package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/spf13/cobra"
)

func restartPrepareEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "restart-prepare [system]",
		Short: "Prepare quick kexec-based restart into new active system",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			if _, err := userutil.RequireRoot(); err != nil {
				return err
			}

			return errors.New("TODO") // regressed

			/*
				sysLabel := args[0]

				system, err := getSystemNotCurrent(sysLabel)
				if err != nil {
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
				kexecOutput, err := exec.CommandContext(ctx, "kexec", "--load",
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
			*/

			fmt.Println(
				"succeeded. to reboot, issue (with sudo):\n    $ systemctl kexec")

			return nil
		}),
	}
}

func restartPrepareCurrentEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "restart-current-prepare",
		Short: "Prepare quick kexec-based restart into same system",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			if _, err := userutil.RequireRoot(); err != nil {
				return err
			}

			systemLabel, err := common.ReadRunningSystemId()
			if err != nil {
				return err
			}

			kexecOutput, err := exec.CommandContext(ctx, "kexec", "--load",
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
		}),
	}
}
