package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/godbus/dbus/v5"
)

const (
	BuildTreeLocation = "/mnt/j-os-inmem-staging"

	AppSYSTEM     = "SYSTEM"
	AppOSCheckout = "OS-checkout"
	AppOSDiff     = "OS-diff"
	AppOSDiffWork = "OS-diff-work"
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

func GetDbusConn() (*dbus.Conn, error) {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, err
	}

	if err = conn.Auth(nil); err != nil {
		return nil, err
	}

	// "This method must be called after authentication, but before sending any other messages to the bus."
	if err = conn.Hello(); err != nil {
		return nil, err
	}

	return conn, nil
}
