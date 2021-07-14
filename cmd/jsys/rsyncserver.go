package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

func rsyncServerEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "rsync-server",
		Short: "Serve the systree over rsync so it can be flashed from a remote computer",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(rsyncServer(
				osutil.CancelOnInterruptOrTerminate(nil)))
		},
	}
}

func rsyncServer(ctx context.Context) error {
	// after starting this, the remote computer can be flashed with:
	//   $ REMOTE=rsync://192.168.1.123/ ./jsys_linux-amd64 flash system_b

	// reading some files in the systree require root access
	if err := requireRoot(); err != nil {
		return err
	}

	tempConfFile := "/tmp/jsys-rsyncd.conf"
	if err := os.WriteFile(tempConfFile, []byte(`
log file = /dev/stdout

# doesn't work without root
# TODO: might work now that we're root. test it.
use chroot = no

# preserve root. without this, rsync tries to downgrade privileges
uid = 0
gid = 0

[jsys]
path = /mnt/j-os-inmem-staging
read only = true
timeout = 300

[EFI]
path = misc/esp/EFI
read only = true
timeout = 300

`), 0660); err != nil {
		return err
	}
	defer os.Remove(tempConfFile)

	rsync := exec.CommandContext(ctx, "rsync",
		"--daemon",
		"--no-detach",
		"--config", tempConfFile)
	rsync.Stdout = os.Stdout
	rsync.Stderr = os.Stderr

	return rsync.Run()
}
