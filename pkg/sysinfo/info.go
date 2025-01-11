// Shows information about the installed system
package sysinfo

import (
	"context"
	"fmt"
	"time"

	"github.com/acobaugh/osrelease"
	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/time/timeutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/joonas-fi/joonas-sys/pkg/ostree"
	"github.com/samber/lo"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
)

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show system information",
		Args:  cobra.NoArgs,
		Run:   cli.WrapRun(info),
	}
}

func info(ctx context.Context, _ []string) error {
	sysID, err := common.ReadRunningSystemId()
	if err != nil {
		return err
	}

	updatedTime, err := getSystemUpdatedTime(sysID)
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
	infoTbl.AddRow("OS", fmt.Sprintf("%s %s", osRelease["NAME"], osRelease["VERSION"]))

	infoTbl.AddRow("System ID", sysID)

	fmt.Println(infoTbl.Render())

	return nil
}

// due to how we use the system, update time is the same as install time
func getSystemUpdatedTime(sysID string) (time.Time, error) {
	withErr := func(err error) (time.Time, error) { return time.Time{}, fmt.Errorf("getSystemUpdatedTime: %w", err) }

	checkouts, err := ostree.GetCheckoutsSortedByDate(filelocations.Sysroot)
	if err != nil {
		return withErr(err)
	}

	checkout, found := lo.Find(checkouts, func(checkout ostree.CheckoutWithLabel) bool { return checkout.Dir == sysID })
	if !found {
		return withErr(fmt.Errorf("checkout not found for %s", sysID))
	}

	return checkout.Timestamp, nil
}
