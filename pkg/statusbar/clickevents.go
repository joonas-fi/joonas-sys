package statusbar

// Processes click events sent from i3bar to us (via out stdin).

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/encoding/jsonfile"
)

// calls handleClick() callback as many times as there are incoming click events
func parseClicksFromStdin(ctx context.Context, handleClick func(event clickEvent)) error {
	// cancelable reader b/c if we get canceled by e.g. sibling task, we could end up blocking
	// reading from our stdin (which click events from i3bar) and never exit, so in effect i3bar
	// content will get stuck.
	//
	// IIRC i3 properly closes our stdin so we'll exit properly, but we cannot rely on that alone
	// as we must be able to also exit if stdin is kept open. (os.Stdin.Close() has no effect)
	stdinLines := bufio.NewScanner(newCancelableReader(ctx, os.Stdin))
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

		handleClick(click)
	}
	// IgnoreErrorIfCanceled because "context canceled" error is expected if we cancel
	if err := IgnoreErrorIfCanceled(ctx, stdinLines.Err()); err != nil {
		return err
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

type cancelableReader struct {
	ctx   context.Context
	inner io.Reader // the reader we're actually reading from

	// latest read result (int, error) pair
	readBytes int
	err       chan error
}

// Use when you have potentially blocking syscalls that you need to be able to cancel
//
// https://benjamincongdon.me/blog/2020/04/23/Cancelable-Reads-in-Go/
//
// concurrent Read() calls not allowed
func newCancelableReader(ctx context.Context, inner io.Reader) io.Reader {
	return &cancelableReader{
		ctx:   ctx,
		inner: inner,
		err:   make(chan error, 1), // buffered as not to leak a goroutine on cancel'd read syscall eventual return
	}
}

var _ io.Reader = (*cancelableReader)(nil)

// concurrent Read() calls not allowed
// do not Read() after the first canceled Read() returns error
func (i *cancelableReader) Read(p []byte) (int, error) {
	select { // check for cancellation status pre-read
	case <-i.ctx.Done():
		return 0, i.ctx.Err()
	default:
	}

	// start a goroutine, so if we end up blocking on a read() syscall, we can still interrupt the Read()
	go func() {
		var err error
		i.readBytes, err = i.inner.Read(p)
		i.err <- err
	}()

	select {
	case <-i.ctx.Done():
		return 0, i.ctx.Err()
	case err := <-i.err:
		return i.readBytes, err
	}
}
