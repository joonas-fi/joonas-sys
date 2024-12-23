package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/function61/gokit/app/cli"
	"github.com/spf13/cobra"
)

func versionReportEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "versions",
		Short: "Prints software versions that were installed",
		Args:  cobra.NoArgs,
		Run: cli.WrapRun(func(_ context.Context, _ []string) error {
			return versionReport()
		}),
	}
}

// #versioncommand: clipcatd --version
// can't use ^ because we're processing the whole file as a string
var versionCommandRe = regexp.MustCompile(`\n#versioncommand: (.+)`)

func versionReport() error {
	steps, err := loadAndValidateSteps()
	if err != nil {
		return err
	}

	// TODO: base OS echo "$(. /etc/os-release; echo \"$VERSION\")"

	for _, step := range steps {
		script, err := ioutil.ReadFile(filepath.Join(installStepsDir, step.ScriptName))
		if err != nil {
			return err
		}

		matches := versionCommandRe.FindStringSubmatch(string(script))
		if matches == nil { // versioncommand is optional
			continue // skip
		}

		// use shell to run it
		versionOutputRaw, err := exec.Command("sh", "-c", matches[1]).Output()
		if err != nil {
			return fmt.Errorf("%s: %s: %w: %s", step.FriendlyName, matches[1], err, versionOutputRaw)
		}

		// many commands output multiple lines, where only the first is significant.
		// we do the equivalent of "| head -1" here to reduce repetition in version commands.
		versionOutputFirstLine := strings.TrimRight(strings.Split(string(versionOutputRaw), "\n")[0], "\r")

		fmt.Printf("%s -> %s\n", step.FriendlyName, versionOutputFirstLine)
	}

	return nil
}
