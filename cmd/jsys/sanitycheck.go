package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

func sanityCheckEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "sanitycheck",
		Short: "Check that important files are where the should be",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(sanityCheck(
				osutil.CancelOnInterruptOrTerminate(nil)))
		},
	}
}

func sanityCheck(ctx context.Context) error {
	filesThatShouldExist := []string{
		"/persist/apps/SYSTEM_nobackup/backlight-state",
		"/persist/apps/SYSTEM_nobackup/rfkill-state",
		"/persist/apps/SYSTEM_nobackup/hostname",
		"/persist/apps/SYSTEM_nobackup/active_sys_id",
		"/persist/apps/SYSTEM_nobackup/background.png",
		"/persist/apps/SYSTEM_nobackup/lowdiskspace-check-rules/root",
		"/persist/apps/Desktop",
		"/persist/apps/ssh-server/ssh_host_ed25519_key",
		"/persist/apps/varasto",
		"/persist/apps/docker/data_nobackup",
		"/persist/apps/mcfly",
		"/persist/apps/git config", // usually symlink to Varasto, sometimes may be local "fork" (custom Git author for client work on client laptop)
		"/dev/cpu_temp",
	}

	missingFileCount := 0

	for _, fileThatShouldExist := range filesThatShouldExist {
		exists, err := osutil.Exists(fileThatShouldExist)
		if err != nil {
			return err
		}

		if !exists {
			fmt.Printf("file missing: %s\n", fileThatShouldExist)
			missingFileCount++
		}
	}

	if missingFileCount > 0 {
		return errors.New("missing files")
	} else {
		return nil
	}
}
