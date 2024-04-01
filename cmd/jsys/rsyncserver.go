package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	rsyncdconfig "github.com/gokrazy/rsync/pkg/config"
	"github.com/gokrazy/rsync/pkg/rsyncd"
	"github.com/joonas-fi/joonas-sys/pkg/common"
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

func rsyncServerV2(ctx context.Context) error {
	listener, err := net.Listen("tcp", "0.0.0.0:873")
	if err != nil {
		return err
	}

	fmt.Println("Pro-tip: on remote run $ jsys flash --remote ${SYSTEM_ID}")

	rsyncServer := rsyncd.NewServer(
		rsyncdconfig.NewModule("jsys", common.BuildTreeLocation, nil),
		rsyncdconfig.NewModule("EFI", "misc/esp/EFI", nil))

	if err := rsyncServer.Serve(ctx, listener); err != nil {
		return err
	}

	return nil
}

func rsyncServer(ctx context.Context) error {
	if os.Getenv("OLD_RSYNC") == "" {
		return rsyncServerV2(ctx)
	} else {
		return rsyncServerLegacy(ctx)
	}
}

// TODO: serve on top of tailscale userspace networking as a temporary tagged service?
func rsyncServerLegacy(ctx context.Context) error {
	// after starting this, the remote computer can be flashed with:
	//   $ REMOTE=rsync://192.168.1.123/ ./jsys_linux-amd64 flash system_b

	// reading some files in the systree require root access
	if err := requireRoot(); err != nil {
		return err
	}

	fmt.Println("Pro-tip: on remote run $ jsys flash --remote ${SYSTEM_ID}")

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
