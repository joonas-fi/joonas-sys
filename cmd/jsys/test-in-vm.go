package main

// Utility for testing a system (that exists either on a partition or in-RAM) in a VM

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/function61/gokit/app/cli"
	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/joonas-fi/joonas-sys/pkg/ostree"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func testInVMEntrypoint() *cobra.Command {
	rescue := false
	wipe := false

	cmd := &cobra.Command{
		Use:   "test-in-vm",
		Short: "Tests new system version in a VM",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			return testInVM(ctx, rescue, wipe)
		}),
	}

	cmd.Flags().BoolVarP(&rescue, "rescue", "", rescue, "Enter rescue (AKA single-user) mode: no GUI or network.")
	cmd.Flags().BoolVarP(&wipe, "wipe", "", wipe, "Wipe state (diffs) before boot")

	return cmd
}

func testInVM(ctx context.Context, rescue bool, wipe bool) error {
	if _, err := userutil.RequireRoot(); err != nil {
		return err
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

	checkout := filelocations.Sysroot.Checkout(sysID)

	// cannot be in /tmp because then our topology would be:
	// host: overlayfs -> virtiofsd
	// guest: virtiofs -> overlay
	// (overlayFS cannot be hosted on overlayFS)
	//
	// VM's sysroot is actually encapsulated in our sysroot's app directory
	vmSysroot := filelocations.WithRoot(filelocations.Sysroot.App("OS-test-in-VM"))

	if wipe {
		if err := os.RemoveAll(vmSysroot.Diff(sysID)); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(vmSysroot.DiffWork(), 0755); err != nil {
		return err
	}

	if err := writeBoilerplateFiles(vmSysroot, sysID); err != nil {
		return err
	}

	if len(Must(os.ReadDir(vmSysroot.CheckoutsDir()))) == 0 { // no checkouts => not mounted
		if err := syscall.Mount(filelocations.Sysroot.CheckoutsDir(), vmSysroot.CheckoutsDir(), "", uintptr(syscall.MS_BIND), ""); err != nil {
			return fmt.Errorf("bind-mount of checkouts to VM: %w", err)
		}

		// ChatGPT:
		// > The MS_BIND flag is used to mirror the directory tree, but it doesn't affect the permissions of the original mount point.
		// > To change the permissions (make it read-only), you need to remount the bind mount with the MS_REMOUNT and MS_RDONLY flags.
		//
		// additional context: https://unix.stackexchange.com/a/128388

		if err := syscall.Mount("", vmSysroot.CheckoutsDir(), "", uintptr(syscall.MS_BIND|syscall.MS_RDONLY|syscall.MS_REMOUNT), ""); err != nil {
			return fmt.Errorf("bind-mount of checkouts to VM remount r/o: %w", err)
		}
	}

	virtioFSDSockPath := "/tmp/jsys-virtiofsd.sock"

	tasks := taskrunner.New(ctx, slog.Default())

	tasks.Start("FS", func(ctx context.Context) error {
		virtiofsd := exec.CommandContext(ctx, "virtiofsd",
			"--socket-path="+virtioFSDSockPath,
			"--cache=never",
			"-o", "modcaps=+sys_admin", // needed to support overlayfs
			"--xattr", // needed to support overlayfs
			"--shared-dir="+vmSysroot.Root())
		virtiofsd.Stdout = os.Stdout
		virtiofsd.Stderr = os.Stderr
		return virtiofsd.Run()
	})

	kernelCmdline := append([]string{"rootfstype=virtiofs", "root=vroot"}, createKernelCmdline(sysID)...)
	if rescue {
		kernelCmdline = append(kernelCmdline, "systemd.unit=rescue.target")
	}

	tasks.Start("VM", func(ctx context.Context) error {
		// if we want some accelerated display:
		//  -device virtio-vga-gl,xres=1920,yres=1080 -display gtk,gl=on

		// if we want to boot with UEFI:
		// "-drive", "if=pflash,format=raw,unit=0,readonly=on,file=misc/uefi-files/OVMF_CODE-pure-efi.fd",

		vm := exec.CommandContext(ctx, "qemu-system-x86_64",
			"-machine", "type=q35,accel=kvm",
			"-m", "3G",
			"-smp", "4",
			"-chardev", "socket,id=char0,path="+virtioFSDSockPath,
			"-device", "vhost-user-fs-pci,queue-size=1024,chardev=char0,tag=vroot",
			"-object", "memory-backend-file,id=mem,size=3G,mem-path=/dev/shm,share=on", // required by virtiofsd
			"-numa", "node,memdev=mem",
			"-usb",
			"-device", "qemu-xhci",
			"-device", "usb-host,vendorid=0x04f9,productid=0x009a", // printer
			"-kernel", filepath.Join(checkout, "/boot/vmlinuz"),
			"-initrd", filepath.Join(checkout, "/boot/initrd.img"),
			"-append", strings.Join(kernelCmdline, " "),
			// "-drive", "format=raw,file="+volatilePersistPartition,
		)
		vm.Stdout = os.Stdout
		vm.Stderr = os.Stderr

		return vm.Run()
	})

	return tasks.Wait()
}

// these minimum amount of files need to exist in order for the system to be usable
func writeBoilerplateFiles(root filelocations.Root, sysVersion string) error {
	withErr := func(err error) error { return fmt.Errorf("writeBoilerplateFiles: %w", err) }

	path := func(p string) string { return filepath.Join(root.Root(), p) }

	writeFile := func(pathRelative string, content string, mode fs.FileMode) error {
		pathInPersist := path(pathRelative)

		if err := os.MkdirAll(filepath.Dir(pathInPersist), 0775); err != nil {
			return err
		}

		if err := os.WriteFile(pathInPersist, []byte(content), mode); err != nil {
			return fmt.Errorf("write %s: %w", pathInPersist, err)
		}

		return nil
	}

	if err := writeFile("apps/SYSTEM/hostname", "j-sys-test-vm", 0660); err != nil {
		return withErr(err)
	}

	// many places blow up without this. needs to be readable to all users.
	// https://xkcd.com/221/
	if err := writeFile("apps/SYSTEM/machine-id", "f5610b0c906aa304e98ea0fa6609649c\n", 0664); err != nil {
		return withErr(err)
	}

	if err := copyBackgroundFromCurrentSystemIfExistsTo(path("apps/SYSTEM/background.png")); err != nil {
		return withErr(err)
	}

	// we need this to comply with `App()` "factory" to just add the `apps/` prefix (not `/sysroot/apps/`)
	dummyRoot := filelocations.WithRoot("/")

	app := func(appName string) string { return dummyRoot.App(appName) }

	for _, dirToCreate := range []string{
		"apps/SYSTEM/backlight-state",
		"apps/SYSTEM/rfkill-state",
		"apps/SYSTEM/lowdiskspace-check-rules",
		app(common.AppOSCheckout), // most likely this will be a mountpoint so we can piggyback off of host checkouts
		fmt.Sprintf("apps/OS-diff/%s", sysVersion),
		fmt.Sprintf("apps/OS-diff/%s-work", sysVersion),
		"apps/docker/data",
		"apps/docker/config",
		app("zoxide"),
		app("varasto"),
		app("Desktop"),
		app("mcfly"),
	} {
		alreadyExists, err := osutil.Exists(path(dirToCreate))
		if err != nil {
			return withErr(err)
		}

		if alreadyExists {
			continue
		}

		if err := os.MkdirAll(path(dirToCreate), 0777); err != nil {
			return withErr(err)
		}

		// umask doesn't give us 0777 from above (FIXME)
		if err := os.Chmod(path(dirToCreate), 0777); err != nil {
			return withErr(err)
		}
	}

	for _, symlink := range []struct {
		from string
		to   string
	}{
		{path("apps/docker/cli-plugins"), "/etc/docker-cli-plugins/"},
		{path(app("SYSTEM_nobackup")), common.AppSYSTEM}, // backwards compat
	} {
		exists, err := osutil.ExistsNoLinkFollow(symlink.from)
		if err != nil {
			return withErr(fmt.Errorf("symlink %s: %w", symlink.from, err))
		}

		// symlink creation errors if it already exists
		if exists {
			continue
		}

		if err := os.Symlink(symlink.to, symlink.from); err != nil {
			return withErr(err)
		}
	}

	// the regular `/etc/fstab` file that instructs to mount an actual block device partition with some heuristic
	// is not compatible with our VM since we use virtiofs to expose not a block device but the FS directly.
	if err := writeFile(fmt.Sprintf("apps/OS-diff/%s/etc/fstab", sysVersion), "# file purposefully empty\n", 0600); err != nil {
		return withErr(err)
	}

	return nil
}
