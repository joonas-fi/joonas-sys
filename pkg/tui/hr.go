package tui

// does same as https://github.com/LuRsT/hr

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/turbobob/pkg/ansicolor"
	"github.com/function61/turbobob/pkg/powerline"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func HREntrypoint() *cobra.Command {
	return &cobra.Command{
		Use: "hr",
		Run: cli.RunnerNoArgs(hr),
	}
}

func hr(_ context.Context, _ *log.Logger) error {
	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}

	// TODO: automatically detect or opt-in?
	const usePowerline = true

	if usePowerline {
		blackOnWhite := powerline.ColorPair{ansicolor.Black, ansicolor.White}
		const powerlineOverhead = 4
		powerlineStr := powerline.Generate(powerline.NewSegment(strings.Repeat(" ", width-powerlineOverhead), blackOnWhite))
		fmt.Println(powerlineStr)
	} else {
		fmt.Println(strings.Repeat("#", width))
	}

	return nil
}
