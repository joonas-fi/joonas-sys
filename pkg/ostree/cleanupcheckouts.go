package ostree

import (
	"context"
	"log/slog"
	"os"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/os/user/userutil"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/spf13/cobra"
)

func checkoutsCleanupEntrypoint() *cobra.Command {
	execute := false

	cmd := &cobra.Command{
		Use:   "checkouts-cleanup",
		Short: "Cleanup checkouts that are unused (don't have corresponding diff)",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, args []string) error {
			return checkoutsCleanup(ctx, execute)
		}),
	}

	cmd.Flags().BoolVarP(&execute, "execute", "x", execute, "Remove the checkouts for real instead of just listing them")

	return cmd
}

func checkoutsCleanup(ctx context.Context, execute bool) error {
	if _, err := userutil.RequireRoot(); err != nil {
		return err
	}

	checkouts, err := os.ReadDir(filelocations.Sysroot.CheckoutsDir())
	if err != nil {
		return err
	}

	for _, checkout := range checkouts {
		hasDiff, err := osutil.Exists(filelocations.Sysroot.Diff(checkout.Name()))
		if err != nil {
			return err
		}

		slog.Info("checkout", "name", checkout.Name(), "hasDiff", hasDiff)

		if !hasDiff && execute {
			slog.Info("deleting unused checkout", "checkout", filelocations.Sysroot.Checkout(checkout.Name()))

			if err := os.RemoveAll(filelocations.Sysroot.Checkout(checkout.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}
