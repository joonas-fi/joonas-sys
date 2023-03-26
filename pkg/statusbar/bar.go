package statusbar

// status bar widgets. wraps i3status and adds new widgets on top.
//
// Will probably at some later date replace i3status completely since it's easier to customize widgets here.

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync/atomic"

	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
)

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "statuswidgets",
		Short: "Extends i3status with custom widgets",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			rootLogger := logex.StandardLogger()

			osutil.ExitIfError(logic(
				osutil.CancelOnInterruptOrTerminate(rootLogger),
				rootLogger))
		},
	}
}

// we augment i3status with additional elements
func logic(ctx context.Context, logger *log.Logger) error {
	latestNetworkItem := &atomic.Value{} // *barItem

	// used as key to netlink.LinkByIndex(...)
	internetFacingLinkIdxAtomic := &atomic.Value{} // int
	if err := storeDefaultLinkIndex(internetFacingLinkIdxAtomic); err != nil {
		return err
	}

	tasks := taskrunner.New(ctx, logger)

	// i3 sends click events via our stdin
	tasks.Start("clickevents", func(ctx context.Context) error {
		return parseClicksFromStdin(ctx, func(click clickEvent) {
			switch click.Name {
			case "inetbw":
				if err := startInteractiveShellCommandInDialog("nethogs", "nethogs"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			case "tztime":
				if err := startInteractiveShellCommandInDialog("cal", "jsys cal --interactive --weeknumbers"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			case "disk_info":
				if err := startInteractiveShellCommandInDialog("lfs", "lfs; read"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			case "cpu_usage":
				if err := startInteractiveShellCommandInDialog("htop", "htop"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			case "memory":
				if err := startInteractiveShellCommandInDialog("free", "free -m; read"); err != nil {
					log.Printf("%v", err) // has enough error context
				}
			default:
				log.Printf("unmapped click '%s'", click.Name)
			}
		})
	})

	requestRefreshCh := make(chan Void, 1)

	requestRefresh := func() {
		select {
		case requestRefreshCh <- Void{}:
		default:
		}
	}

	itemsFromI3Status := make(chan []barItem, 1)

	// in the future i3status integration will be optional
	if true {
		tasks.Start("i3status", func(ctx context.Context) error {
			return getOutputFromI3Status(ctx, func(items []barItem) { // we'll get notified of new items
				itemsFromI3Status <- items
			})
		})
	}

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

	tasks.Start("micmonitor", func(ctx context.Context) error {
		return micMonitorTask(ctx, requestRefresh)
	})

	latestItems := []barItem{}

	sender := newI3barProtocolSenderSendHeaders()

	doRefresh := func() {
		prepend := []barItem{}

		if item := getPossibleMicRecordingItem(); item != nil {
			prepend = append(prepend, *item)
		}

		// this is where we append our augmented modules
		if item := latestNetworkItem.Load(); item != nil {
			prepend = append(prepend, *item.(*barItem))
		}

		items := append(prepend, latestItems...)

		sender.writeBarItems(items)
	}

	for {
		select {
		case err := <-tasks.Done():
			return err
		case items := <-itemsFromI3Status:
			latestItems = items
			doRefresh()
		case <-requestRefreshCh:
			doRefresh()
		}
	}
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
