package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/function61/gokit/app/retry"
	"github.com/function61/gokit/os/osutil"
)

// TODO: move to gokit/os/osutil
func copyFile(sourcePath string, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close() // double close intentional

	destination, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destination.Close() // double close intentional

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	if err := destination.Close(); err != nil {
		return err
	}

	return source.Close()
}

func waitForFileAvailable(ctx context.Context, file string) error {
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
	}, retry.DefaultBackoff(), func(err error) { log.Println(err.Error()) })
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

func requireRoot() error {
	if os.Getuid() != 0 {
		return errors.New("need root")
	}

	return nil
}
