package statusbar

// status bar widgets. wraps i3status and adds new widgets on top.
//
// Will probably at some later date replace i3status completely since it's easier to customize widgets here.

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "statuswidgets",
		Short: "Extends i3status with custom widgets",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			osutil.ExitIfError(logic(
				osutil.CancelOnInterruptOrTerminate(nil)))
		},
	}
}

// we augment i3status with additional elements
func logic(ctx context.Context) error {
	latestNetworkItem := &atomic.Value{} // *barItem

	// used as key to netlink.LinkByIndex(...)
	internetFacingLinkIdxAtomic := &atomic.Value{} // int
	if err := storeDefaultLinkIndex(internetFacingLinkIdxAtomic); err != nil {
		return err
	}

	tasks := taskrunner.New(ctx, logex.Discard)

	// i3 sends click events via our stdin
	tasks.Start("clickevents", func(ctx context.Context) error {
		// NOTE: os.Stdin reads are blocking (.Close() will not help), so this task never exits even
		//       if context is cancelled, unless the party doing the cancellation also closes stdin
		//       (I think i3 luckily does that)
		//
		// https://benjamincongdon.me/blog/2020/04/23/Cancelable-Reads-in-Go/

		stdinLines := bufio.NewScanner(os.Stdin)
		for stdinLines.Scan() {
			if stdinLines.Text() == "[" { // start of endless JSON object stream
				continue // just ignore it
			}

			// each line is an ,{ ... } (expect first payload line)
			// ",{ ... }" => "{ ... }"
			object := strings.TrimPrefix(stdinLines.Text(), ",")

			click := clickEvent{}
			if err := jsonfile.UnmarshalDisallowUnknownFields(strings.NewReader(object), &click); err != nil {
				return err
			}

			switch click.Name {
			case "inetbw":
				if err := startInteractiveShellCommandInDialog("nethogs", "nethogs"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			case "tztime":
				if err := startInteractiveShellCommandInDialog("cal", "cal; read"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			case "disk_info":
				if err := startInteractiveShellCommandInDialog("df", "df -h; read"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			case "cpu_usage":
				if err := startInteractiveShellCommandInDialog("htop", "htop"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			default:
				log.Printf("unmapped click '%s'", click.Name)
			}
		}
		log.Println("unblocked from scan")
		if err := stdinLines.Err(); err != nil {
			return err
		}

		return nil
	})

	tasks.Start("i3status-augment", func(ctx context.Context) error {
		return augmentI3Status(ctx, func(items []barItem) ([]barItem, error) {
			prepend := []barItem{}

			// this is where we append our augmented modules
			if item := latestNetworkItem.Load(); item != nil {
				prepend = append(prepend, *item.(*barItem))
			}

			return append(prepend, items...), nil
		})
	})

	// each time the routing table changes is is a natural time to refresh the link list and see which
	// one of them is the internet-facing one.
	tasks.Start("routesubscriber", func(ctx context.Context) error {
		routesUpdated := make(chan netlink.RouteUpdate, 1)
		stopSubscription := make(chan struct{})
		defer close(stopSubscription)
		if err := netlink.RouteSubscribe(routesUpdated, stopSubscription); err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-routesUpdated:
				log.Println("routes updated")

				// update the internet facing link index atomic, so it will be picked up by
				// networkPoller() during next poll
				if err := storeDefaultLinkIndex(internetFacingLinkIdxAtomic); err != nil {
					return err
				}
			}
		}
	})

	tasks.Start("networkpoll", func(ctx context.Context) error {
		return networkPoller(ctx, internetFacingLinkIdxAtomic, latestNetworkItem)
	})

	return tasks.Wait()
}

func networkPoller(ctx context.Context, internetFacingLinkIdxAtomic *atomic.Value, latestNetworkItem *atomic.Value) error {
	refreshRate := 5 * time.Second

	var statsPrev *netlink.LinkStatistics

	pollOnce := func() error {
		internetFacingLink, err := netlink.LinkByIndex(internetFacingLinkIdxAtomic.Load().(int))
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

		latestNetworkItem.Store(&barItem{
			Name: "inetbw",
			// Instance: internetFacingLinkName,
			Instance: "",
			Markup:   "none",
			FullText: largestOfRxOrTx,
		})

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
		return 0, os.ErrNotExist
	}

	return defaultRoutes[0].LinkIndex, nil
}

// wrapper for storing defaultLinkIndex in an atomic value
func storeDefaultLinkIndex(to *atomic.Value) error {
	defaultLinkIndex, err := getDefaultLinkIndex()
	if err != nil {
		return err
	}

	to.Store(defaultLinkIndex)

	return nil
}

// interactive so .bashrc gets sourced (so we support aliases, functions etc.)
func startInteractiveShellCommandInDialog(description string, shellCommand string) error {
	// window class with magic prefix so window manager can match a rule to transform it into a dialog
	// (and not automatically-tiled window for instance, if using a tiling window manager)
	windowClass := fmt.Sprintf("%s_9dc82_dialog", description)

	if err := exec.Command("i3-sensible-terminal", "--class", windowClass, "--command", "/bin/bash", "-i", "-c", shellCommand).Start(); err != nil {
		return fmt.Errorf("startInteractiveShellCommandInDialog(%s): %w", description, err)
	}

	return nil
}

type clickEvent struct {
	Name       string   `json:"name"`
	Instance   string   `json:"instance"`
	Button     int      `json:"button"`
	Modifiers  []string `json:"modifiers"`
	X          int      `json:"x"`
	Y          int      `json:"y"`
	Relative_x int      `json:"relative_x"`
	Relative_y int      `json:"relative_y"`
	Output_x   int      `json:"output_x"`
	Output_y   int      `json:"output_y"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
}
