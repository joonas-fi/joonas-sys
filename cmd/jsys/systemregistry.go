package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	systemRegistry = []systemSpec{
		{
			label: "system_a",

			warnIfNotRunningIn: "system_b",

			systemDevice: "/dev/disk/by-label/system_a",

			espDevice: "_AUTODETECT_", // or hardcode something like /dev/disk/by-label/ESP
		},
		{
			label: "system_b",

			warnIfNotRunningIn: "system_a",

			systemDevice: "/dev/disk/by-label/system_b",

			espDevice: "_AUTODETECT_", // or hardcode something like /dev/disk/by-label/ESP
		},
		{
			label:       "in-ram",
			labelActual: "system_a", // if testing in a VM, we internally refer to it as system_a

			warnIfNotRunningIn: "", // no warnings

			systemDevice:                    "/dev/shm/joonas-os-ram-image",
			systemDeviceCanCreateIfNotFound: true, // only because RAM-backed disk (not messing with any persistent media)

			espDevice:                    "/dev/disk/by-label/ESP-VM", // RAM-backed disk for a VM
			espDeviceCanCreateIfNotFound: true,                        // only because RAM-backed disk (not messing with any persistent media)
		},
	}
)

type systemSpec struct {
	label       string
	labelActual string

	warnIfNotRunningIn string

	systemDevice                    string
	systemDeviceCanCreateIfNotFound bool

	espDevice                    string
	espDeviceCanCreateIfNotFound bool
}

func (s systemSpec) lieAboutLabelIfVirtualMachine() string {
	if s.labelActual != "" {
		return s.labelActual
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
	sys, err := func() (systemSpec, error) {
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
	}()
	if err != nil {
		return sys, err
	}

	if sys.espDevice == "_AUTODETECT_" {
		labels, err := os.ReadDir("/dev/disk/by-label")
		if err != nil {
			return sys, err
		}

		candidates := []string{}

		for _, label := range labels {
			if strings.HasPrefix(label.Name(), "ESP") && label.Name() != "ESP-VM" {
				candidates = append(candidates, "/dev/disk/by-label/"+label.Name())
			}
		}

		candidate, err := func() (string, error) {
			if len(candidates) == 1 {
				return candidates[0], nil
			} else {
				return "", fmt.Errorf("autodetect used but got %d candidates", len(candidates))
			}
		}()
		if err != nil {
			return sys, err
		}

		sys.espDevice = candidate
	}

	return sys, nil
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
