package main

import (
	"os"

	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/os/osutil"
	"github.com/joonas-fi/joonas-sys/pkg/calendar"
	"github.com/joonas-fi/joonas-sys/pkg/debug"
	"github.com/joonas-fi/joonas-sys/pkg/discoverremotemachines"
	"github.com/joonas-fi/joonas-sys/pkg/tui"
	"github.com/spf13/cobra"
)

func main() {
	app := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Management tools for joonas-sys",
		Version: dynversion.Version,
	}

	app.AddCommand(buildEntrypoint())
	app.AddCommand(flashEntrypoint())
	app.AddCommand(testInVmEntrypoint())
	app.AddCommand(diffEntrypoint())
	app.AddCommand(diffOneEntrypoint())
	app.AddCommand(revertEntrypoint())
	app.AddCommand(espEntrypoint())
	app.AddCommand(restartPrepareEntrypoint())
	app.AddCommand(restartPrepareCurrentEntrypoint())
	app.AddCommand(versionReportEntrypoint())
	app.AddCommand(rsyncServerEntrypoint())
	app.AddCommand(backlightEntrypoint())
	app.AddCommand(lowDiskSpaceCheckerEntrypoint())
	app.AddCommand(sanityCheckEntrypoint())
	app.AddCommand(infoEntrypoint())
	app.AddCommand(discoverremotemachines.Entrypoint())
	app.AddCommand(statusbar.Entrypoint())
	app.AddCommand(calendar.Entrypoint())
	app.AddCommand(tui.HREntrypoint())
	app.AddCommand(debug.Entrypoint())

	osutil.ExitIfError(app.Execute())
}
