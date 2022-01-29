package statusbar

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/encoding/jsonfile"
)

type barItem struct {
	Name     string `json:"name"`
	Instance string `json:"instance"`
	Markup   string `json:"markup"`
	FullText string `json:"full_text"`
	Color    string `json:"color"` // example: "#FF0000"
}

func augmentI3Status(ctx context.Context, augment func(items []barItem) ([]barItem, error)) error {
	i3status := exec.Command("i3status")
	stdout, err := i3status.StdoutPipe()
	if err != nil {
		return err
	}

	augmentOneLine := func(line string, prefix string) error {
		items := []barItem{}

		lineJson := strings.TrimPrefix(line, prefix)
		if err := jsonfile.UnmarshalDisallowUnknownFields(strings.NewReader(lineJson), &items); err != nil {
			return err
		}

		augmentedItems, err := augment(items)
		if err != nil {
			return err
		}

		augmentedJSON, err := json.Marshal(augmentedItems) // re-assemble
		if err != nil {
			return err
		}

		fmt.Println(prefix + string(augmentedJSON))

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
				if err := augmentOneLine(line, ","); err != nil {
					panic(err)
				}
			case strings.HasPrefix(line, "[{"): // first item line. "[" not enough to discriminate
				if err := augmentOneLine(line, ""); err != nil {
					panic(err)
				}
			case line == `{"version":1}`:
				// add that we support click events
				fmt.Println(`{ "version": 1, "click_events": true }`)
			default:
				fmt.Println(line) // passthrough as -is
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

	return IgnoreErrorIfCanceled(ctx, i3status.Run())
}
