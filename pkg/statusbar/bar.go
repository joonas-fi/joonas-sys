package statusbar

// status bar widgets. wraps i3status and adds new widgets on top.
//
// Will probably at some later date replace i3status completely since it's easier to customize widgets here.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/function61/gokit/app/cli"
	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
)

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "statuswidgets",
		Short: "Extends i3status with custom widgets",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			err := IgnoreErrorIfCanceled(ctx, logic(ctx))
			if err != nil {
				errWrite := os.WriteFile("/tmp/jsys-statusbar-last-fail-reason.txt", []byte(err.Error()), 0600)
				if errWrite != nil {
					slog.Error("write of fail reason", "errWrite", errWrite)
				}
			}
			return err
		}),
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

	tasks := taskrunner.New(ctx, slog.Default())

	// i3 sends click events via our stdin
	tasks.Start("clickevents", func(ctx context.Context) error {
		return parseClicksFromStdin(ctx, func(click clickEvent) {
			switch click.Name {
			case "inetbw":
				if err := startInteractiveShellCommandInDialog("nethogs", "nethogs"); err != nil {
					slog.Error("", "err", err) // has enough error context
				}
			case "tztime":
				if err := startInteractiveShellCommandInDialog("cal", "jsys cal --interactive --weeknumbers"); err != nil {
					slog.Error("", "err", err) // has enough error context
				}
			case "disk_info":
				if err := startInteractiveShellCommandInDialog("lfs", "lfs; read"); err != nil {
					slog.Error("", "err", err) // has enough error context
				}
			case "cpu_usage":
				if err := startInteractiveShellCommandInDialog("htop", "htop"); err != nil {
					slog.Error("", "err", err) // has enough error context
				}
			case "memory":
				if err := startInteractiveShellCommandInDialog("free", "free --mega --human; read"); err != nil {
					slog.Error("", "err", err) // has enough error context
				}
			default:
				slog.Warn("unmapped click", "name", click.Name)
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
	routeSubscriber := func(ctx context.Context) error {
		// since this task can crash (see `retryForever()` callsite) we should get the truth at the start
		// because the start might be a consequence of a crash that could have updated the truth
		if err := storeDefaultLinkIndex(internetFacingLinkIdxAtomic); err != nil {
			return err
		}

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
			case _, ok := <-routesUpdated:
				if !ok {
					return errors.New("routesUpdated channel closed, WTF") // this actually happens
				}

				slog.Info("routes updated")

				// update the internet facing link index atomic, so it will be picked up by
				// networkPoller() during next poll
				if err := storeDefaultLinkIndex(internetFacingLinkIdxAtomic); err != nil {
					return err
				}
			}
		}
	}

	tasks.Start("routesubscriber", func(ctx context.Context) error {
		// the dumbass `netlink.RouteSubscribe` can actually close the channel which will result in
		// having to re-subscribe. retrying this task forever was the most pragmatic fix.
		return retryForever(ctx, routeSubscriber)
	})

	tasks.Start("networkpoll", func(ctx context.Context) error {
		return networkPoller(ctx, internetFacingLinkIdxAtomic, latestNetworkItem)
	})

	/*
		tasks.Start("micmonitor", func(ctx context.Context) error {
			return micMonitorTask(ctx, requestRefresh)
		})
	*/

	tasks.Start("powermonitor", func(ctx context.Context) error {
		return powerMonitor(ctx, requestRefresh)
	})

	latestItems := []barItem{}

	sender := newI3barProtocolSenderSendHeaders()

	doRefresh := func() {
		prepend := []barItem{}

		if item := getPossibleMicRecordingItem(); item != nil {
			prepend = append(prepend, *item)
		}

		if item := getPossibleBatteryLowItem(); item != nil {
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

func retryForever(ctx context.Context, task func(context.Context) error) error {
	for {
		err := task(ctx)
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done(): // asked to stop, so the stop of `task` was expected
			return nil
		default:
			slog.Error("retryForever task errored; retrying", "err", err)

			time.Sleep(1 * time.Second) // don't hammer at full speed
		}
	}
}
