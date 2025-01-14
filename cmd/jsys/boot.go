package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/spf13/cobra"
)

func bootEntrypoint() *cobra.Command {
	now := true

	cmd := &cobra.Command{
		Use:   "boot",
		Short: "Do quick kexec-based boot into currently active system",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			if _, err := userutil.RequireRoot(); err != nil {
				return err
			}

			sysIDbytes, err := os.ReadFile(activeSystemVersionPath())
			if err != nil {
				return err
			}

			sysID := string(sysIDbytes)

			if err := kexecLoad(ctx, sysID); err != nil {
				return err
			}

			if now {
				if err := kexecBoot(ctx); err != nil {
					return err
				}
			} else {
				fmt.Printf("succeeded. to reboot, issue (with sudo):\n    $ systemctl kexec\n")
			}

			return nil
		}),
	}

	cmd.Flags().BoolVarP(&now, "now", "", now, "Boot immediately (if not do just $ kexec --load ...)")

	return cmd
}

func kexecLoad(ctx context.Context, sysID string) error {
	root := filelocations.Sysroot.Checkout(sysID)

	kexecOutput, err := exec.CommandContext(ctx, "kexec", "--load",
		"--command-line="+strings.Join(createKernelCmdline(sysID), " "),
		"--initrd="+filepath.Join(root, "/boot/initrd.img"),
		filepath.Join(root, "/boot/vmlinuz"),
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("kexec --load: %w: %s", err, kexecOutput)
	}

	return nil
}

func kexecBoot(ctx context.Context) error {
	if output, err := exec.CommandContext(ctx, "systemctl", "kexec").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl kexec: %w: %s", err, string(output))
	}

	return nil
}
