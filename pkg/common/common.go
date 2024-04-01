package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
)

const (
	BuildTreeLocation = "/mnt/j-os-inmem-staging"

	AppSYSTEM     = "SYSTEM"
	AppOSCheckout = "OS-checkout"
	AppOSDiff     = "OS-diff"
)

// "sysid=v1" => "v1"
var runningSystemIdFromKernelCommandLineRe = regexp.MustCompile("sysid=([^ ]+)")

func ReadRunningSystemId() (string, error) {
	withErr := func(err error) (string, error) { return "", fmt.Errorf("ReadRunningSystemId: %w", err) }

	kernelCommandLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return withErr(err)
	}

	matches := runningSystemIdFromKernelCommandLineRe.FindStringSubmatch(string(kernelCommandLine))
	if matches == nil {
		return withErr(errors.New("failed to parse"))
	}

	return matches[1], nil
}
