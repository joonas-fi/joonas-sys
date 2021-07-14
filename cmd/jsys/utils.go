package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/function61/gokit/app/retry"
	"github.com/function61/gokit/os/osutil"
)

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
