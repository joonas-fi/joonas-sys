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

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
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
			return errors.New("regressed")
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
