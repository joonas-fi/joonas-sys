// message grouping, deduplication
package notificationdedup

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/gokit/net/netutil"
	"github.com/function61/gokit/sync/syncutil"
	"github.com/gofrs/flock"
	"github.com/spf13/cobra"
)

type displayedMessage struct {
	displayedAt time.Time
	summary     string
	lines       []string
	id          string // needed to replace existing message
}

type dedupState struct {
	mu                sync.Mutex
	displayedMessages map[string]displayedMessage
}

type dedupMsg struct {
	Summary string `json:"summary"`
	Text    string `json:"text"`
}

func handleClient(conn net.Conn, state *dedupState) error {
	msg := dedupMsg{}
	if err := jsonfile.UnmarshalDisallowUnknownFields(conn, &msg); err != nil {
		return err
	}

	defer syncutil.LockAndUnlock(&state.mu)()

	for _, displayedMessage := range state.displayedMessages {
		if displayedMessage.summary != msg.Summary {
			continue
		}

		// matching message found => update it

		return nil
	}

	// matching message not found => show it as totally new one
	slog.Info("show as new", "msg", msg)

	_ = conn.Close()

	return nil
}

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use: "notifydedup",
		// Short: "...",
		Args: cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			state := &dedupState{
				displayedMessages: map[string]displayedMessage{},
			}

			// test:
			//   $ echo '{"summary":"device add","text":"/usb/stuff"}' | nc -N -U /run/notifydedup/server.sock
			return netutil.ListenUnixAllowEveryone(ctx, "/run/notifydedup/server.sock", func(listener net.Listener) error {
				for {
					conn, err := listener.Accept()
					if err != nil {
						return err
					}

					go func() {
						if err := handleClient(conn, state); err != nil {
							slog.Error("handleClient", "err", err)
						}
					}()
				}
			})
		}),
	}
}

type dedupFile struct {
	Counter int `json:"counter"`
}

const (
	dedupFilePath     = "/run/notifydedup/counter.json"
	dedupFileLockPath = "/run/notifydedup/counter.json.lock"
)

func WorkerEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use: "notifydedup-worker",
		// Short: "...",
		Args: cobra.NoArgs,
		Run: cli.WrapRun(func(_ context.Context, _ []string) error {
			// https://gavv.github.io/articles/file-locks/
			lock := flock.New(dedupFileLockPath)
			if err := lock.Lock(); err != nil {
				return err
			}
			defer func() {
				if err := lock.Unlock(); err != nil {
					panic(err)
				}
			}()

			dedup := dedupFile{}
			if err := jsonfile.ReadDisallowUnknownFields(dedupFilePath, &dedup); err != nil {
				return err
			}

			dedup.Counter++

			return jsonfile.Write(dedupFilePath, dedup)
		}),
	}
}
