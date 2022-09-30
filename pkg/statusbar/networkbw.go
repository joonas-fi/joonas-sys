package statusbar

// Network bandwidth bar item

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	. "github.com/function61/gokit/builtin"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func networkPoller(ctx context.Context, internetFacingLinkIdxAtomic *atomic.Value, latestNetworkItem *atomic.Value) error {
	refreshRate := 5 * time.Second

	var statsPrev *netlink.LinkStatistics

	pollOnce := func() error {
		storeInetbwBarItem := func(fullText string) {
			latestNetworkItem.Store(&barItem{
				Name: "inetbw",
				// Instance: internetFacingLinkName,
				Instance: "",
				Markup:   "none",
				FullText: fullText,
			})
		}

		internetFacingLinkIdx := internetFacingLinkIdxAtomic.Load().(int)
		if internetFacingLinkIdx == -1 { // no default route
			storeInetbwBarItem("Offline")
			return nil
		}
		internetFacingLink, err := netlink.LinkByIndex(internetFacingLinkIdx)
		if err != nil {
			return err
		}

		statsNow := *internetFacingLink.Attrs().Statistics // shorthand

		if statsPrev == nil { // first iteration
			statsPrev = &statsNow
		}

		// stats difference per e.g. five seconds
		statsDiff := subtract(statsNow, *statsPrev)

		statsPerSecond := multiply(statsDiff, 1/refreshRate.Seconds())

		largestOfRxOrTx := func() string {
			if statsPerSecond.RxBytes >= statsPerSecond.TxBytes {
				return fmt.Sprintf("⬇️ %s/s", toFixedWidthKiloBytesOrMegaBytes(int(statsPerSecond.RxBytes)))
			} else {
				return fmt.Sprintf("⬆️ %s/s", toFixedWidthKiloBytesOrMegaBytes(int(statsPerSecond.TxBytes)))
			}
		}()

		storeInetbwBarItem(largestOfRxOrTx)

		statsPrev = &statsNow

		return nil
	}

	if err := pollOnce(); err != nil { // initial
		return err
	}

	refresh := time.NewTicker(refreshRate)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-refresh.C:
			if err := pollOnce(); err != nil {
				return err
			}
		}
	}
}

func subtract(a, b netlink.LinkStatistics) netlink.LinkStatistics {
	return addMul(a, b, -1)
}

func multiply(a netlink.LinkStatistics, multiplier float64) netlink.LinkStatistics {
	return addMul(netlink.LinkStatistics{}, a, multiplier)
}

func addMul(a, b netlink.LinkStatistics, multiplier float64) netlink.LinkStatistics {
	return netlink.LinkStatistics{
		RxBytes: a.RxBytes + uint64(multiplier*float64(b.RxBytes)),
		TxBytes: a.TxBytes + uint64(multiplier*float64(b.TxBytes)),
	}
}

// focused in producing fixed-width bandwidth measurements in scale humans are used to reading
func toFixedWidthKiloBytesOrMegaBytes(size int) string {
	if size == 0 {
		return "  0 kB"
	}

	inKilobytesFloat := fmt.Sprintf("%.1f kB", float64(size)/1024)
	if len(inKilobytesFloat) <= 6 {
		return inKilobytesFloat
	}

	inKilobytes := fmt.Sprintf(" %d kB", int(math.Round(float64(size)/1024)))
	if len(inKilobytes) <= 6 {
		return inKilobytes
	}

	inMegabytesFloat := fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	if len(inMegabytesFloat) <= 6 {
		return inMegabytesFloat
	}

	return fmt.Sprintf("%3d MB", int(math.Round(float64(size)/(1024*1024))))
}

// find the link where the default route points to.
// this is usually the internet-facing link.
// index is -1 if no default route
func getDefaultLinkIndex() (int, error) {
	routes, err := netlink.RouteList(nil, unix.AF_INET)
	if err != nil {
		return 0, err
	}

	defaultRoutes := []netlink.Route{}
	for _, route := range routes {
		if route.Dst == nil { // a default route is one that doesn't have a destination subnet
			defaultRoutes = append(defaultRoutes, route)
		}
	}

	if len(defaultRoutes) < 1 {
		return -1, nil
	}

	return defaultRoutes[0].LinkIndex, nil
}

// wrapper for storing defaultLinkIndex in an atomic value
func storeDefaultLinkIndex(to *atomic.Value) error {
	defaultLinkIndex, err := getDefaultLinkIndex()
	if err != nil {
		return ErrorWrap("storeDefaultLinkIndex", err)
	}

	to.Store(defaultLinkIndex)

	return nil
}
