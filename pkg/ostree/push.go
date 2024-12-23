package ostree

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/os/osutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/spf13/cobra"
)

const (
	distrepoPath = "/tmp/distrepo"
)

func pushEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push to joonas-sys remote",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(ctx context.Context, _ []string) error {
			return push(ctx)
		}),
	}
}

func push(ctx context.Context) error {
	if err := os.MkdirAll(distrepoPath, 0766); err != nil {
		return err
	}

	if err := pushInit(ctx); err != nil {
		return err
	}

	slog.Info("pulling locally to archive format", "dest", distrepoPath)

	if err := pushPullLocal(ctx); err != nil {
		return err
	}

	slog.Info("pushing to remote repo", "eta_guesstimate", "30m")

	pushStarted := time.Now()

	if err := pushPush(ctx); err != nil {
		return err
	}

	slog.Info("succeeded", "push_duration", time.Since(pushStarted))

	return nil
}

func pushInit(ctx context.Context) error {
	isInited, err := osutil.Exists(filepath.Join(distrepoPath, "objects"))
	if err != nil {
		return err
	}

	if !isInited {
		slog.Info("doing init")

		// `--mode=archive` is important in order for the files to be representable in S3-like systems (symlinks are as regular files etc.).
		// (this also explains why the distribution repo and the actual local repo are different.)
		if output, err := exec.CommandContext(ctx, "ostree", "init", "--mode=archive", "--repo="+distrepoPath, "--collection-id=fi.joonas.os").CombinedOutput(); err != nil {
			return fmt.Errorf("push: %w: %s", err, string(output))
		}
	}

	return nil
}

func pushPullLocal(ctx context.Context) error {
	pullLocal := exec.CommandContext(ctx, "ostree", "pull-local", "--repo="+distrepoPath, filelocations.Sysroot.App(common.AppOSRepo), ostreeBranchNameX8664)
	pullLocal.Stdout = os.Stdout
	pullLocal.Stderr = os.Stderr

	if err := pullLocal.Run(); err != nil {
		return fmt.Errorf("push: pull-local. %w", err)
	}

	return nil
}

func pushPush(ctx context.Context) error {
	// TODO: rclone config: https://github.com/joonas-fi/joonas-sys/issues/55

	sync := exec.CommandContext(ctx, "rclone", "copy", "--progress", ".", "joonas-os:/fi-joonas-os/ostree/")
	sync.Dir = distrepoPath
	sync.Stdout = os.Stdout
	sync.Stderr = os.Stderr

	if err := sync.Run(); err != nil {
		return fmt.Errorf("push: rclone: %w", err)
	}

	return nil
}
