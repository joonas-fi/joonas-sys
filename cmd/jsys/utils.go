package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/function61/gokit/app/retry"
	"github.com/function61/gokit/os/osutil"
	"github.com/manifoldco/promptui"
)

func waitForFileAvailable(ctx context.Context, file string) error {
	errorLogger := retry.ErrorsToLogger(slog.With("op", "waitForFileAvailable"))

	return retry.Retry(ctx, func(ctx context.Context) error {
		exists, err := osutil.Exists(file)
		if err != nil {
			return err
		}

		if exists {
			return nil
		} else {
			return fmt.Errorf("not yet available: %s", file)
		}
	}, retry.DefaultBackoff(), retry.IgnoreErrorsWithin(3*time.Second, errorLogger))
}

// shell equivalent: "$ rm -f". os.RemoveAll() is very close to we want, but it can be dangerous
// (it removes children)
func removeIfExists(file string) error {
	exists, err := osutil.Exists(file)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	return os.Remove(file)
}

// shell equivalent: "$ touch"
func createEmptyFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	return file.Close()
}

func removeDirectoryChildren(directory string) error {
	dentries, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, dentry := range dentries {
		if err := os.RemoveAll(filepath.Join(directory, dentry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// hides the boilerplate of promptui.Select{}.Run()
func promptUISelect(label string, items []string) (int, string, error) {
	return (&promptui.Select{
		Label:    label,
		Size:     20, // increased from ridiculous minimum
		Items:    items,
		Stdout:   &bellSkipper{}, // don't do audible bell on item highlight
		HideHelp: true,           // hide "Use the arrow keys to navigate: ↓ ↑ → ←"
	}).Run()
}

// bellSkipper implements an io.WriteCloser that skips the terminal bell
// character (ASCII code 7), and writes the rest to os.Stderr. It is used to
// replace readline.Stdout, that is the package used by promptui to display the
// prompts.
//
// This is a workaround for the bell issue documented in
// https://github.com/manifoldco/promptui/issues/49.
type bellSkipper struct{}

// Write implements an io.WriterCloser over os.Stderr, but it skips the terminal
// bell character.
func (bs *bellSkipper) Write(b []byte) (int, error) {
	const charBell = 7 // c.f. readline.CharBell
	if len(b) == 1 && b[0] == charBell {
		return 0, nil
	}
	return os.Stderr.Write(b)
}

// Close implements an io.WriterCloser over os.Stderr.
func (bs *bellSkipper) Close() error {
	return os.Stderr.Close()
}

func createAndTruncateFile(path string, size int64) error {
	exists, err := osutil.Exists(path)
	if err != nil {
		return fmt.Errorf("createAndTruncateFile: %w", err)
	}

	if !exists { // file needs to exist before we can call truncate
		if err := createEmptyFile(path); err != nil {
			return err
		}
	}

	return os.Truncate(path, size)
}
