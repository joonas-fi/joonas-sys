package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/esiqveland/notify"
	"github.com/function61/gokit/os/osutil"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/cobra"
)

// Commands for backlight control (increase/decrease brightness)

const (
	// points to the backlight interface, e.g. "/sys/class/backlight/intel_backlight"
	backlightPath = "/dev/screen-backlight"

	// assumes a udev rule sets up a symlink
	keyboardBacklightDevice = "/dev/keyboard-backlight"

	desktopNotificationPreviousIDFile = "backlightctl/desktop-notification-previous-id"
)

func backlightEntrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backlight",
		Short: "Backlight management (screen/keyboard/...)",
	}

	cmd.AddCommand(backlightKeyboardEntrypoint())
	cmd.AddCommand(backlightScreenEntrypoint())

	return cmd
}

func backlightKeyboardEntrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keyboard",
		Short: "Keyboard backlight",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "cycle",
		Short: "Cycle keyboard backlight (off/medium/high)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(keyboardBacklightCycle())
		},
	})

	return cmd
}

func backlightScreenEntrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen",
		Short: "Screen backlight",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "increase",
		Short: "Increase brightness",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(backlightAdjustBy(context.Background(), 0.10))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "decrease",
		Short: "Decrease brightness",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(backlightAdjustBy(context.Background(), -0.10))
		},
	})

	return cmd
}

// cycles values between 0 and max_brightness. this is sensible only when the numbers are mapped to
// "modes", i.e. 0 => off, 1 => medium, 2 => high etc.
func keyboardBacklightCycle() error {
	brightnessPath := filepath.Join(keyboardBacklightDevice, "brightness")

	max, err := readIntFile(filepath.Join(keyboardBacklightDevice, "max_brightness"))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("keyboard backlight not found. did you forget to set up the symlink? (%s)", keyboardBacklightDevice)
		} else {
			return err
		}
	}

	current, err := readIntFile(brightnessPath)
	if err != nil {
		return err
	}

	// With max=2
	// - current=0 -> 1
	// - current=1 -> 2
	// - current=2 -> 0
	next := (current + 1) % (max + 1)

	return os.WriteFile(brightnessPath, []byte(strconv.Itoa(next)), 0611)
}

func backlightAdjustBy(ctx context.Context, incrementPercentagePoints float64) error {
	maxBrightness, err := readIntFile(filepath.Join(backlightPath, "max_brightness"))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backlight not found. did you forget to set up the symlink? (%s)", backlightPath)
		} else {
			return err
		}
	}

	brightnessFile := filepath.Join(backlightPath, "brightness")

	currentBrightness, err := readIntFile(brightnessFile)
	if err != nil {
		return err
	}

	// 0 %, 10 %, ..., 90 %, 100 %
	currentPercentageRoundedToTens := math.Round(float64(currentBrightness)/float64(maxBrightness)*10) / 10

	newPercentage := clamp(currentPercentageRoundedToTens+incrementPercentagePoints, 0.0, 1.0)

	newAbsoluteValue := int(float64(maxBrightness) * newPercentage)

	if err := os.WriteFile(brightnessFile, []byte(fmt.Sprintf("%d\n", newAbsoluteValue)), 0511); err != nil {
		return fmt.Errorf("error adjusting backlight! did you forget to set up udev rule to give permissions? %w", err)
	}

	if err := notifyNewBrightness(newPercentage); err != nil {
		return fmt.Errorf("notification: %w", err)
	}

	return nil
}

// reads a file with integer as text and ends in newline.
// e.g. "123\n"
func readIntFile(path string) (int, error) {
	intStr, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimRight(string(intStr), "\n"))
}

func clamp(num, min, max float64) float64 {
	switch {
	case num < min:
		return min
	case num > max:
		return max
	default:
		return num
	}
}

// desktop notifications boilerplate. TODO: this is duplicated in hautomo-client

// TODO: instead subscribe via udev to brightness notifications so control + notification is decoupled?
func notifyNewBrightness(newBrighness float64) error {
	dbusConn, err := getDbusConn()
	if err != nil {
		return err
	}

	// to prevent having multiple concurrent brightness notifications, keep track of
	// "notification correlation id" so we can replace any previous notification (if one is still visible)
	desktopNotificationPreviousIDPath := filepath.Join(runtimeDir(), desktopNotificationPreviousIDFile)

	return notifyWithConcurrentSuppression(desktopNotificationPreviousIDPath, func(replacesID uint32) (uint32, error) {
		return notify.SendNotification(dbusConn, notify.Notification{
			AppName:       "backlightctl",
			ReplacesID:    replacesID,
			Summary:       "🔆",
			ExpireTimeout: 2500 * time.Millisecond,
			Body:          progressBar(int(newBrighness*100), 40, progressBarDefaultTheme()),
		})
	})
}

// wraps notify.SendNotification() with suppression
func notifyWithConcurrentSuppression(
	desktopNotificationPreviousIDPath string,
	sendNotification func(replacesID uint32) (uint32, error),
) error {
	replacesID, err := func() (uint32, error) {
		cache, err := os.ReadFile(desktopNotificationPreviousIDPath)
		if err != nil {
			if os.IsNotExist(err) {
				return 0, nil // not an error - there simply isn't a previous one
			} else {
				return 0, err // some unexpected error
			}
		}

		return binary.LittleEndian.Uint32(cache), nil
	}()
	if err != nil {
		return err
	}

	newReplacesID, err := sendNotification(replacesID)
	if err != nil {
		return err
	}

	if newReplacesID != replacesID {
		if err := os.MkdirAll(filepath.Dir(desktopNotificationPreviousIDPath), 0700); err != nil {
			return fmt.Errorf("writing desktopNotificationPreviousIDPath: %w", err)
		}

		newReplacesIDBytes := [4]byte{}
		binary.LittleEndian.PutUint32(newReplacesIDBytes[:], newReplacesID)
		if err := os.WriteFile(desktopNotificationPreviousIDPath, newReplacesIDBytes[:], 0600); err != nil {
			return fmt.Errorf("writing desktopNotificationPreviousIDPath: %w", err)
		}
	}

	return nil
}

func getDbusConn() (*dbus.Conn, error) {
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

func progressBar(pct int, barLength int, theme progressBarTheme) string {
	r := make([]rune, barLength)

	ratio := float64(barLength) * float64(pct) / 100.0

	for i := 0; i < barLength; i++ {
		ch := theme.Vacant
		if float64(i+1) <= ratio {
			ch = theme.Filled
		}

		r[i] = ch
	}

	return string(r)
}

type progressBarTheme struct {
	Filled rune
	Vacant rune
}

func progressBarDefaultTheme() progressBarTheme {
	return progressBarTheme{'█', '░'}
}

// returns XDG_RUNTIME_DIR which usually is /run/user/<uid> (e.g. /run/user/1000)
func runtimeDir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir
	} else { // fallback
		return fmt.Sprintf("/run/user/%d", os.Getuid())
	}
}
