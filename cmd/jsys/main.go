package main

import (
	"os"

	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/os/osutil"
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
	app.AddCommand(currentSysIdEntrypoint())
	app.AddCommand(versionReportEntrypoint())
	app.AddCommand(rsyncServerEntrypoint())

	osutil.ExitIfError(app.Execute())
}
