// Debug tools
package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/function61/gokit/app/cli"
	"github.com/joonas-fi/joonas-sys/pkg/sysfs"
	"github.com/spf13/cobra"
)

func Entrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug tools",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(_ context.Context, _ []string) error {
			// we rely on udev rules setting up /dev/powermonitor-<device name> symlinks that
			// link to e.g. /sys/class/power_supply/hidpp_battery_0.
			// from that dir we can read the uevent file that describes its power status etc.
			dentries, err := filepath.Glob("/dev/powermonitor-*")
			if err != nil {
				return err
			}

			for _, dentry := range dentries {
				status, err := sysfs.PowerSupplyDir(dentry).ReadUevent()
				if err != nil {
					return err
				}

				fmt.Printf("%s = %s\n", status.MODEL_NAME, status.CAPACITY_LEVEL)
			}

			return nil
		}),
	}

	cmd.AddCommand(udevadmWalk())

	return cmd
}

func udevadmWalk() *cobra.Command {
	return &cobra.Command{
		Use:   "udevadm-walk [syspath]",
		Short: "Report the udev attributes for a device",
		Args:  cobra.ExactArgs(1),
		Run: cli.WrapRun(func(ctx context.Context, args []string) error {
			attributeWalk := exec.CommandContext(ctx, "udevadm", "info", "--attribute-walk", args[0])
			attributeWalk.Stdout = os.Stdout
			attributeWalk.Stderr = os.Stderr
			return attributeWalk.Run()
		}),
	}
}
