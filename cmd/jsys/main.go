package main

import (
	"os"

	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/os/osutil"
	"github.com/joonas-fi/joonas-sys/pkg/backlight"
	"github.com/joonas-fi/joonas-sys/pkg/backup"
	"github.com/joonas-fi/joonas-sys/pkg/calendar"
	"github.com/joonas-fi/joonas-sys/pkg/debug"
	"github.com/joonas-fi/joonas-sys/pkg/discoverremotemachines"
	"github.com/joonas-fi/joonas-sys/pkg/lowdiskspacechecker"
	"github.com/joonas-fi/joonas-sys/pkg/notificationdedup"
	"github.com/joonas-fi/joonas-sys/pkg/ostree"
	"github.com/joonas-fi/joonas-sys/pkg/statusbar"
	"github.com/joonas-fi/joonas-sys/pkg/sysdiff"
	"github.com/joonas-fi/joonas-sys/pkg/sysinfo"
	"github.com/joonas-fi/joonas-sys/pkg/tui"
	"github.com/spf13/cobra"
)

func main() {
	app := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Management tools for joonas-sys",
		Version: dynversion.Version,
	}

	for _, entrypoint := range sysdiff.Entrypoints() {
		app.AddCommand(entrypoint)
	}

	app.AddCommand(buildEntrypoint())
	app.AddCommand(flashEntrypoint())
	app.AddCommand(testInVMEntrypoint())
	app.AddCommand(espEntrypoint())
	app.AddCommand(restartPrepareEntrypoint())
	app.AddCommand(restartPrepareCurrentEntrypoint())
	app.AddCommand(versionReportEntrypoint())
	app.AddCommand(rsyncServerEntrypoint())
	app.AddCommand(backlight.Entrypoint())
	app.AddCommand(lowdiskspacechecker.Entrypoint())
	app.AddCommand(sanityCheckEntrypoint())
	app.AddCommand(sysinfo.Entrypoint())
	app.AddCommand(discoverremotemachines.Entrypoint())
	app.AddCommand(backup.Entrypoint())
	app.AddCommand(statusbar.Entrypoint())
	app.AddCommand(calendar.Entrypoint())
	app.AddCommand(tui.HREntrypoint())
	app.AddCommand(debug.Entrypoint())
	app.AddCommand(ostree.Entrypoint())

	app.AddCommand(notificationdedup.Entrypoint())
	app.AddCommand(notificationdedup.WorkerEntrypoint())

	osutil.ExitIfError(app.Execute())
}
