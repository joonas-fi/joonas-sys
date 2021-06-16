package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

var (
	systemRegistry = []systemSpec{
		{
			label: "system_a",

			warnIfNotRunningIn: "system_b",

			systemDevice: "/dev/disk/by-label/system_a",

			espDevice: "/dev/disk/by-label/ESP-USB-DT", // b/c current partition too smol
		},
		{
			label: "system_b",

			warnIfNotRunningIn: "system_a",

			systemDevice: "/dev/disk/by-label/system_b",

			espDevice: "/dev/disk/by-label/ESP-USB-DT", // b/c current partition too smol
		},
		{
			label: "in-ram",

			warnIfNotRunningIn: "", // no warnings

			systemDevice:                    "/dev/shm/joonas-os-ram-image",
			systemDeviceCanCreateIfNotFound: true, // only because RAM-backed disk (not messing with any persistent media)

			espDevice:                    "/dev/disk/by-label/ESP-VM", // RAM-backed disk for a VM
			espDeviceCanCreateIfNotFound: true,                        // only because RAM-backed disk (not messing with any persistent media)
		},
	}
)

type systemSpec struct {
	label string

	warnIfNotRunningIn string

	systemDevice                    string
	systemDeviceCanCreateIfNotFound bool

	espDevice                    string
	espDeviceCanCreateIfNotFound bool
}

func (s systemSpec) lieAboutLabelIfVirtualMachine() string {
	if s.label == "in-ram" {
		return "system_a"
	} else {
		return s.label
	}
}

func (s systemSpec) diffPath() string {
	return fmt.Sprintf("/persist/apps/SYSTEM_nobackup/%s-diff", s.label)
}

func (s systemSpec) espDeviceLabel() (string, error) {
	diskByLabelPrefix := "/dev/disk/by-label/"

	if strings.HasPrefix(s.espDevice, diskByLabelPrefix) {
		return strings.TrimPrefix(s.espDevice, diskByLabelPrefix), nil
	}

	return "", fmt.Errorf(
		"ESP device does not start with '"+diskByLabelPrefix+"', cannot deduce label for %s",
		s.espDevice)
}

var inlineSystemSpecRe = regexp.MustCompile(`^([^,]+),sysdev=([^,]+),espdev=([^,]+)$`)

func getSystemNoEditCheck(label string) (systemSpec, error) {
	if matches := inlineSystemSpecRe.FindStringSubmatch(label); matches != nil {
		return systemSpec{
			label:        matches[1],
			systemDevice: matches[2],
			espDevice:    matches[3],
		}, nil
	}

	for _, system := range systemRegistry {
		if system.label == label {
			return system, nil
		}
	}

	return systemSpec{}, fmt.Errorf("system not found: %s", label)
}

func getSystemNotCurrent(label string) (systemSpec, error) {
	runningSysId, err := readRunningSystemId()
	if err != nil {
		return systemSpec{}, err
	}

	if label == runningSysId {
		return systemSpec{}, errors.New("unsafe: trying to edit the running system you're running on right now.")
	}

	return getSystemNoEditCheck(label)
}

func currentSysIdEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "current-sys-id",
		Short: "Prints currently running system ID",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func() error {
				sysId, err := readRunningSystemId()
				if err != nil {
					return err
				}

				fmt.Println(sysId)

				return nil
			}())
		},
	}
}
