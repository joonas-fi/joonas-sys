package statusbar

// Parses bar items from i3status, so we can benefit from implementations and only add what we need
// on top ("compose").

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"syscall"

	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/encoding/jsonfile"
)

func getOutputFromI3Status(
	ctx context.Context,
	notifyOfNewItems func(items []barItem),
) error {
	i3status := exec.Command("i3status")
	stdout, err := i3status.StdoutPipe()
	if err != nil {
		return err
	}

	processOneLine := func(line string, prefix string) error {
		items := []barItem{}

		lineJson := strings.TrimPrefix(line, prefix)
		if err := jsonfile.UnmarshalDisallowUnknownFields(strings.NewReader(lineJson), &items); err != nil {
			return err
		}

		notifyOfNewItems(items)

		return nil
	}

	go func() {
		// https://i3wm.org/docs/i3bar-protocol.html
		stdoutLines := bufio.NewScanner(stdout)
		for stdoutLines.Scan() {
			line := stdoutLines.Text()

			/* First three lines are special:

			{"version":1}
			[
			[{ ... item ...

			All the following lines are:

			,[{ ... item ...
			*/
			switch {
			case strings.HasPrefix(line, ",["): // subsequent item lines (the most common case)
				if err := processOneLine(line, ","); err != nil {
					panic(err)
				}
			case strings.HasPrefix(line, "[{"): // first item line. "[" not enough to discriminate
				if err := processOneLine(line, ""); err != nil {
					panic(err)
				}
			case line == `{"version":1}` || line == "[":
				// no-op
			default: //
				panic("unrecognized line: " + line)
			}
		}

		if err := stdoutLines.Err(); err != nil {
			panic(err)
		}
	}()

	go func() {
		<-ctx.Done()

		// ask the process nicely to stop
		if err := i3status.Process.Signal(syscall.SIGINT); err != nil {
			panic(err)
		}
	}()

	// start '$ i3status' and also wait it to exit
	return IgnoreErrorIfCanceled(ctx, i3status.Run())
}
