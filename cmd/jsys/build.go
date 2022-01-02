package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/function61/gokit/os/osutil"
	"github.com/joonas-fi/joonas-sys/pkg/logtee"
	"github.com/spf13/cobra"
)

const (
	treeLocation = "/mnt/j-os-inmem-staging"
)

type Step struct {
	ScriptName   string // "426 - LibreOffice.sh"
	FriendlyName string // ScriptName -> "LibreOffice"
	logLines     []string
}

func buildEntrypoint() *cobra.Command {
	keep := false
	rm := false
	verbose := false
	fancyUI := true

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Builds the system tree (so it can be flashed somewhere)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func() error {
				started := time.Now()
				defer func() {
					duration := time.Since(started)
					if duration > 1*time.Second { // fast failures are not interesting
						fmt.Printf("finished in %s\n", duration)
					}
				}()

				return build(
					osutil.CancelOnInterruptOrTerminate(nil),
					keep,
					rm,
					verbose,
					fancyUI)
			}())
		},
	}

	cmd.Flags().BoolVarP(&keep, "keep", "", keep, "Keep current tree (if one exists)")
	cmd.Flags().BoolVarP(&rm, "rm", "", rm, "Remove current tree (if one exists)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", verbose, "Verbose logging output")
	cmd.Flags().BoolVarP(&fancyUI, "fancy-ui", "", fancyUI, "Use fancy UI to show progress")

	return cmd
}

func build(ctx context.Context, keep bool, rm bool, verbose bool, fancyUI bool) error {
	if err := requireRoot(); err != nil {
		return err
	}

	if keep && rm {
		return errors.New("illegal to combine --keep & --rm")
	}

	currentTreeExists, err := osutil.Exists(treeLocation)
	if err != nil {
		return err
	}

	createRamdisk := func() error {
		// create mount point
		if err := os.MkdirAll(treeLocation, 0777); err != nil {
			return fmt.Errorf("error making mount point: %w", err)
		}

		if err := syscall.Mount("", treeLocation, "tmpfs", 0, "rw,size=20G"); err != nil {
			return fmt.Errorf("failed mounting RAM disk for %s: %w", treeLocation, err)
		}

		return nil
	}

	if currentTreeExists {
		switch {
		case keep:
			// no-op
		case rm:
			if err := removeDirectoryChildren(treeLocation); err != nil {
				return err
			}

			if err := createRamdisk(); err != nil {
				return err
			}
		default:
			return errors.New("current systree exists. cannot continue")
		}
	} else {
		if err := createRamdisk(); err != nil {
			return err
		}
	}

	mounted, err := isMounted(treeLocation)
	if err != nil {
		return err
	}

	if !mounted {
		return fmt.Errorf("%s is not a mountpoint. would write files to disk", treeLocation)
	}

	steps, err := loadAndValidateSteps()
	if err != nil {
		return err
	}

	if err := buildSysBuilderImage(ctx, verbose); err != nil {
		return fmt.Errorf("buildSysBuilderImage: %w", err)
	}

	nextStep := make(chan Void) // should be unbuffered to avoid race conditions

	uiCtx, cancel := context.WithCancel(ctx)
	defer func() {
		time.Sleep(7000 * time.Millisecond) // give time for UI to be tore down. FIXME
		cancel()
		time.Sleep(100 * time.Millisecond) // give time for UI to be tore down. FIXME
	}()

	appendLogLine := make(chan string)

	go func() {
		if fancyUI {
			if err := displayFancyUI(uiCtx, nextStep, steps, appendLogLine); err != nil {
				panic(err)
			}
		} else {
			for {
				select {
				case line := <-appendLogLine:
					fmt.Println(line)
				case <-uiCtx.Done():
					return
				case <-nextStep:
					fmt.Println("----------------------")
				}
			}
		}
	}()

	workdir, err := os.Getwd()
	if err != nil {
		return err
	}

	badLinesFixed := 0

	lineCompleted := func(line string) {
		if strings.ContainsAny(line, "\r\n") {
			line = strings.ReplaceAll(line, "\r", "")
			line = strings.ReplaceAll(line, "\n", "")

			badLinesFixed++
		}

		appendLogLine <- line
	}

	for _, step := range steps {
		step := step // pin

		if err := runIfNotAlreadyCompleted(step, func() error {
			cmd := exec.CommandContext(ctx,
				"docker",
				"run",
				"--rm",
				"--volume", fmt.Sprintf("%s:%s:slave", treeLocation, treeLocation),
				"--volume", fmt.Sprintf("%s:/repo", workdir), // shouldn't use ADD in Dockerfile, because we have secrets.env
				"--privileged",
				"--shm-size=1024M", // if default 64M, Nvidia driver installation (via DKMS) fails due to compiler segfault (I guess by null pointer dereference by not checking SHM alloc success?)
				"joonas-sys-builder",
				step.ScriptName,
			)

			tee := logtee.NewLineSplitterTee(ioutil.Discard, lineCompleted)

			cmd.Stdout = tee
			cmd.Stderr = tee

			if err := cmd.Run(); err != nil {
				return err
			}

			if badLinesFixed > 0 {
				lineCompleted(fmt.Sprintf("%d bad lines fixed", badLinesFixed))
				badLinesFixed = 0
			}

			return nil
		}); err != nil {
			// echo -e "Build failed. For interactive debugging:\n    $ docker run --rm -it -v \"${treeLocation}:${treeLocation}:slave\" -v \"\$(pwd):/repo\" --privileged joonas-sys-builder"
			return fmt.Errorf("step '%s' failed: %w", step.FriendlyName, err)
		}

		nextStep <- Void{}
	}

	return nil
}

func runIfNotAlreadyCompleted(step *Step, run func() error) error {
	// inside systree: "/tmp/.joonas-os-install/<step>.flag-completed"
	completedFlag := filepath.Join(treeLocation, "tmp", ".joonas-os-install", fmt.Sprintf("%s.flag-completed", step.ScriptName))

	if err := os.MkdirAll(filepath.Dir(completedFlag), 0777); err != nil {
		return err
	}

	completed, err := osutil.Exists(completedFlag)
	if err != nil {
		return err
	}

	if completed {
		return nil
	}

	if err := run(); err != nil {
		return err
	}

	// mark completed
	if err := createEmptyFile(completedFlag); err != nil {
		return err
	}

	return nil
}

func buildSysBuilderImage(ctx context.Context, verbose bool) error {
	dockerBuild := exec.CommandContext(ctx,
		"docker",
		"build",
		"-t", "joonas-sys-builder",
		".")

	if verbose {
		dockerBuild.Stdout = os.Stdout
		dockerBuild.Stderr = os.Stderr
	}

	return dockerBuild.Run()
}

const (
	installStepsDir = "install-steps"
)

// "426 - LibreOffice.sh", capture group for the friendly name
var validStepNameRe = regexp.MustCompile(`^\d{3} - (.+)\.sh$`)

func loadAndValidateSteps() ([]*Step, error) {
	steps, err := ioutil.ReadDir(installStepsDir)
	if err != nil {
		return nil, err
	}

	// sort by name, because execution order is really important
	sort.Slice(steps, func(i, j int) bool { return steps[i].Name() < steps[j].Name() })

	validSteps := []*Step{}

	for _, stepEntry := range steps {
		if stepEntry.IsDir() {
			return nil, errors.New(installStepsDir + "/ must not contain directories")
		}

		if stepEntry.Name() == "common.sh" { // only exception to the rules
			continue // skip
		}

		if !isExecutable(stepEntry.Mode()) { // would break build
			return nil, fmt.Errorf("not executable: %s", stepEntry.Name())
		}

		matches := validStepNameRe.FindStringSubmatch(stepEntry.Name())
		if matches == nil {
			return nil, fmt.Errorf("invalid step name: %s", stepEntry.Name())
		}

		stepScriptContent, err := ioutil.ReadFile(filepath.Join(installStepsDir, stepEntry.Name()))
		if err != nil {
			return nil, err
		}

		// forgetting this would be catastrophic (allows to continue with errors)
		if !strings.Contains(string(stepScriptContent), "source common.sh") {
			return nil, fmt.Errorf("invalid script: %s", stepEntry.Name())
		}

		validSteps = append(validSteps, &Step{
			ScriptName:   stepEntry.Name(),
			FriendlyName: matches[1],
			logLines:     []string{},
		})
	}

	return validSteps, nil
}

// true if any of owner/group/other execute bit is up
func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
