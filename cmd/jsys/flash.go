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

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/backend/file"
	"github.com/diskfs/go-diskfs/partition/gpt"
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
	espMountpoint        = "/boot/efi"
	tmpMountpointEsp     = "/tmp/jsys-esp"
	tmpMountpointSystem  = "/tmp/jsys-system"
	tmpMountpointPersist = "/tmp/jsys-persist"
)

func flashEFIEntrypoint() *cobra.Command {
	writeBootloaderEntry := true
	sanityCheckBeforeFlash := true
	bootByKexec := false

	cmd := &cobra.Command{
		Use:   "flash",
		Short: "Flash EFI boot partition with target version",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			if _, err := userutil.RequireRoot(); err != nil {
				return err
			}

			if sanityCheckBeforeFlash {
				if err := sanityCheck(ctx); err != nil {
					return fmt.Errorf("sanity check: %w", err)
				}
			}

			bootloaderDestination := filepath.Join(espMountpoint, "EFI/BOOT/BOOTx64.efi")

			espMounted, err := osutil.Exists(bootloaderDestination)
			if err != nil || !espMounted {
				const mountSource = "/dev/disk/by-partlabel/EFI\\x20system\\x20partition"

				slog.Warn("ESP not mounted; mounting",
					"to", espMountpoint,
					"from", mountSource)

				if output, err := exec.CommandContext(ctx, "mount", mountSource, espMountpoint).CombinedOutput(); err != nil {
					return fmt.Errorf("mount: %w: %s", err, string(output))
				}
			}

			sysrootCheckouts, err := ostree.ListVersions(filelocations.Sysroot)
			if err != nil {
				return err
			}
			idx, _, err := promptUISelect("Version", lo.Map(sysrootCheckouts, func(x ostree.CheckoutWithLabel, _ int) string { return x.Label }))
			if err != nil {
				return err
			}

			if _, err = ostree.EnsureCheckedOut(ctx, sysrootCheckouts[idx]); err != nil {
				return err
			}

			sysID := sysrootCheckouts[idx].CommitShort

			// create diff dir (system is unbootable without this)
			if err := os.MkdirAll(filelocations.Sysroot.Diff(sysID), 0755); err != nil {
				return err
			}

			if err := os.MkdirAll(filelocations.Sysroot.DiffWork(), 0755); err != nil {
				return err
			}

			vol1 := fmt.Sprintf("--volume=%s:/sysroot", filelocations.Sysroot.Checkout(sysID))
			vol2 := "--volume=/tmp/ukifybuild:/workspace"

			ukifyBuild := exec.CommandContext(ctx, "docker", "run", "--rm", "-t", vol1, vol2, "ghcr.io/joonas-fi/joonas-sys-ukify:latest", "build",
				"--linux=/sysroot/boot/vmlinuz",
				"--initrd=/sysroot/boot/initrd.img",
				"--cmdline="+strings.Join(createKernelCmdline(sysID), " "),
				"--output=/workspace/BOOTx64.efi")
			if output, err := ukifyBuild.CombinedOutput(); err != nil {
				return fmt.Errorf("ukify: %w: %s", err, string(output))
			}

			if writeBootloaderEntry {
				if err := os.Rename(bootloaderDestination, bootloaderDestination+".old"); err != nil {
					return err
				}

				if err := osutil.CopyFile("/tmp/ukifybuild/BOOTx64.efi", bootloaderDestination); err != nil {
					return err
				}

				if err := os.WriteFile(activeSystemVersionPath(), []byte(sysID), 0644); err != nil {
					return err
				}

				slog.Info("new bootloader now live", "bootloader", bootloaderDestination)

				if bootByKexec {
					if err := kexecLoad(ctx, sysID); err != nil {
						return err
					}

					slog.Info("kexec loaded. about to boot - hold on to your butts.")

					if err := kexecBoot(ctx); err != nil {
						return err
					}
				} else {
					fmt.Printf("pro-tip:\n  $ %s %s\n", os.Args[0], bootEntrypoint().Use)
				}
			} else {
				fmt.Println("pro-tip: (NOTE: take backup of target first)")
				fmt.Println("  $ cp /tmp/ukifybuild/BOOTx64.efi " + bootloaderDestination)
			}

			return nil
		}),
	}

	cmd.Flags().BoolVarP(&writeBootloaderEntry, "write-bootloader", "", writeBootloaderEntry, "Write the bootloader entry, effectively making the change live")
	cmd.Flags().BoolVarP(&sanityCheckBeforeFlash, "sanity", "", sanityCheckBeforeFlash, "Do sanity check before flash")
	cmd.Flags().BoolVarP(&bootByKexec, "boot", "", bootByKexec, "Boot right now (using kexec)")

	cmd.AddCommand(assertDiscoverablePartitionsSpecEntrypoint())

	return cmd
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
	// discover by https://uapi-group.org/specifications/specs/discoverable_partitions_specification/
	// specifically SD_GPT_ROOT_X86_64 (semantics = root partition for x86-64)
	//
	// Linux kernel implements a few configurable strategies for early boot resolving of the root filesystem:
	// - "PARTUUID=..." for using GPT part UUID (this is different than GPT part type)
	// - "PARTLABEL=..." for using GPT part label
	//   (this is not codified by the discoverable partitions spec but this is the best we can do)
	//   https://github.com/torvalds/linux/blob/059dd502b263d8a4e2a84809cf1068d6a3905e6f/block/early-lookup.c#L236
	//
	// WARN: cannot use quotes in the arg (even though kernel docs mention it being supported), it will break boot.
	howToResolveRoot := fmt.Sprintf("root=PARTLABEL=%s", rootPartitionGPTLabel)

	cmdline := append(createKernelCmdlineWithoutRootDiskOption(sysID), howToResolveRoot)
	return cmdline
}

func createKernelCmdlineWithoutRootDiskOption(sysID string) []string {
	return []string{"sysid=" + sysID, "rw"}
}

func activeSystemVersionPath() string {
	return filepath.Join(espMountpoint, "active-system-version.txt")
}

func assertDiscoverablePartitionsSpecEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "assert-discoverable-partitions-spec [devPath]",
		Short: "Check that partitions have correct partition types",
		Args:  cobra.ExactArgs(1),
		Run: cli.WrapRun(func(ctx context.Context, args []string) error {
			if err := assertDiscoverablePartitionsSpec(args[0]); err != nil {
				return err
			}

			return nil
		}),
	}
}

const (
	// GPT labels are not strictly codified in the spec (though it has a name column), it primarily speaks
	// of using GPT partition type as the "foreign key". also it seems that using PARTLABEL with space
	// for Linux to boot the support was either broken in the kernel or the initrd (it failed parsing at space's position)
	// (even though kernel docs indicate space support: https://www.kernel.org/doc/html/v4.14/admin-guide/kernel-parameters.html).
	//
	// so let's just use a simple no-frills label.
	rootPartitionGPTLabel = "root"
)

func assertDiscoverablePartitionsSpec(devicePath string) error {
	dev, err := file.OpenFromPath(devicePath, true)
	if err != nil {
		return err
	}
	defer dev.Close()

	disk, err := diskfs.OpenBackend(dev, diskfs.WithOpenMode(diskfs.ReadOnly))
	if err != nil {
		return err
	}

	partTable := disk.Table.(*gpt.Table)

	if !lo.SomeBy(partTable.Partitions, func(part *gpt.Partition) bool {
		return part.Type == gpt.LinuxRootX86_64 && part.Name == rootPartitionGPTLabel
	}) {
		return errors.New("no partition found matching LinuxRootX86_64 (" + string(gpt.LinuxRootX86_64) + ") and the label " + rootPartitionGPTLabel + "\n  pro-tip: can be changed with gdisk (use type code 8304)")
	}

	return nil

}
