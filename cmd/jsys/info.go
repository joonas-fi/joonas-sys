package main

// Shows information about the installed system

import (
	"fmt"
	"os"
	"time"

	"github.com/acobaugh/osrelease"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/time/timeutil"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
)

func infoEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show system information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(info())
		},
	}
}

func info() error {
	updatedTime, err := getSystemUpdatedTime()
	if err != nil {
		return err
	}

	osRelease, err := osrelease.Read()
	if err != nil {
		return err
	}

	infoTbl := termtables.CreateTable()

	infoTbl.AddRow("Updated", fmt.Sprintf(
		"%s (%s ago)",
		updatedTime.Format("2006-01-02"),
		timeutil.HumanizeDuration(time.Since(updatedTime))))

	// "Ubuntu 20.04.3 LTS (Focal Fossa)"
	infoTbl.AddRow("OS", fmt.Sprintf("OS: %s %s", osRelease["NAME"], osRelease["VERSION"]))

	fmt.Println(infoTbl.Render())

	return nil
}

// due to how we use the system, update time is the same as install time
func getSystemUpdatedTime() (time.Time, error) {
	stat, err := os.Stat("/mnt/sys-current-rom/tmp/.joonas-os-install")
	if err != nil {
		return time.Time{}, err
	}

	installedTime := stat.ModTime()

	return installedTime, nil
}
