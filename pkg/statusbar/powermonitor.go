package statusbar

// shows a warning in the status bar

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/sync/syncutil"
	"github.com/joonas-fi/joonas-sys/pkg/sysfs"
)

func powerMonitor(ctx context.Context, requestRefresh func()) error {
	checkOnceLogError := func() {
		defer requestRefresh()

		batteryLowItems, err := powerMonitorGetWarnings(ctx)
		if err != nil {
			log.Printf("powerMonitor: %v", err)
			return
		}

		if len(batteryLowItems) > 0 {
			// we don't yet support multiple.
			if len(batteryLowItems) > 1 {
				log.Printf("WARN: got > batteryLowItems: %d", len(batteryLowItems))
			}

			setBatteryLowItem(&batteryLowItems[0]) // FIXME
		} else {
			setBatteryLowItem(nil) // clear
		}
	}

	checkOnceLogError() // initial sync

	tick := time.NewTicker(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			checkOnceLogError()
		}
	}

	return nil
}

// returns `barItem` for each power-monitorable device that is low on power.
func powerMonitorGetWarnings(ctx context.Context) ([]barItem, error) {
	// we rely on udev rules setting up /dev/powermonitor-<device name> symlinks that
	// link to e.g. /sys/class/power_supply/hidpp_battery_0.
	// from that dir we can read the uevent file that describes its power status etc.
	powerMonitorSymlinks, err := filepath.Glob("/dev/powermonitor-*")
	if err != nil {
		return nil, err
	}

	batteryLowItems := []barItem{}

	for _, powerMonitorSymlink := range powerMonitorSymlinks {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		status, err := sysfs.PowerSupplyDir(powerMonitorSymlink).ReadUevent()
		if err != nil {
			return nil, err
		}

		if status.CAPACITY_LEVEL.IsBelowNormal() {
			batteryLowItems = append(batteryLowItems, barItem{
				Name:     "battery_low_indicator",
				FullText: fmt.Sprintf("🪫 Battery %s: %s", status.CAPACITY_LEVEL, status.MODEL_NAME),
				Color:    "#ff0000",
			})
		}
	}

	return batteryLowItems, nil
}

var batteryLowItemMu sync.Mutex
var batteryLowItem *barItem

func getPossibleBatteryLowItem() *barItem {
	defer syncutil.LockAndUnlock(&batteryLowItemMu)()

	if batteryLowItem == nil {
		return nil
	}

	return Pointer(*batteryLowItem) // copy
}

func setBatteryLowItem(item *barItem) {
	defer syncutil.LockAndUnlock(&batteryLowItemMu)()

	batteryLowItem = item
}
