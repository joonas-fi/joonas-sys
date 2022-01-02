package main

// Warns about low disk space

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/esiqveland/notify"
	"github.com/function61/gokit/os/osutil"
	"github.com/pkg/xattr"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const (
	rulesDir = "/persist/apps/SYSTEM_nobackup/lowdiskspace-check-rules"
)

func lowDiskSpaceCheckerEntrypoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lowdiskspace-checker [msg]",
		Short: "Shows notification if low on disk space",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(lowDiskSpaceChecker())
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "rules-print",
		Short: "Shows the low disk space checker rules",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func() error {
				rules, err := loadRules()
				if err != nil {
					return err
				}

				for _, rule := range rules {
					fmt.Printf("%s: %s (%d bytes)\n", rule.label, rule.mountpoint, rule.lowThreshold)
				}

				return nil
			}())
		},
	})

	cmd.AddCommand(setThresholdEntrypoint())
	cmd.AddCommand(createRuleEntrypoint())

	cmd.AddCommand(&cobra.Command{
		Use:   "systemd-units",
		Short: "Write systemd units to automatically run this checker periodically",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(writeSystemdUnits())
		},
	})

	return cmd
}

type rule struct {
	mountpoint   string
	label        string
	lowThreshold int64 // bytes
}

func lowDiskSpaceChecker() error {
	rules, err := loadRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		stat := unix.Statfs_t{}

		if err := unix.Statfs(rule.mountpoint, &stat); err != nil {
			return fmt.Errorf("%s: %w", rule.mountpoint, err)
		}

		// (difference between available and free: availableBlocks = freeBlocks - reservedBlocks)
		availableBytes := int64(stat.Bavail) * stat.Bsize

		if availableBytes < rule.lowThreshold {
			if err := func() error {
				dbusConn, err := getDbusConn()
				if err != nil {
					return err
				}
				defer dbusConn.Close()

				if _, err := notify.SendNotification(dbusConn, notify.Notification{
					AppName: "lowdiskspace-checker",
					Summary: "Low disk space",
					Body:    fmt.Sprintf("Low disk space on %s; available = %d MB\n", rule.mountpoint, availableBytes/1024/1024),
				}); err != nil {
					return fmt.Errorf("failed to send notification: %w", err)
				}

				return nil
			}(); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadRules() ([]rule, error) {
	items, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil, err
	}

	rules := []rule{}

	for _, item := range items {
		path := ruleFile(item.Name())

		// symlink might sound attractive instead of a regular file, but it could cause unnecessary
		// burden for file scan tools, and you can't set xattrs for symlinks:
		// https://bugs.launchpad.net/ubuntu/+source/linux/+bug/919896
		target, err := xattr.Get(path, "user.target")
		if err != nil {
			return nil, fmt.Errorf("user.target xattr: %w", err)
		}

		thresholdStr, err := xattr.Get(path, "user.threshold")
		if err != nil {
			return nil, fmt.Errorf("user.threshold xattr: %w", err)
		}

		threshold, err := strconv.ParseInt(string(thresholdStr), 10, 64)
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule{
			mountpoint:   string(target),
			label:        item.Name(),
			lowThreshold: threshold,
		})
	}

	return rules, nil
}

func createRuleEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "rule-create [rule-name] [target]",
		Short: "Create low disk space rule",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(createRule(args[0], args[1]))
		},
	}
}

func setThresholdEntrypoint() *cobra.Command {
	mb := int64(0)
	gb := int64(0)

	cmd := &cobra.Command{
		Use:   "rule-set-threshold [rule-name]",
		Short: "Set low disk space rule's threshold",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func() error {
				switch {
				case mb != 0 && gb != 0:
					return errors.New("cannot define both --mb and --gb")
				case mb != 0:
					return setThreshold(args[0], gb*1024*1024)
				case gb != 0:
					return setThreshold(args[0], gb*1024*1024*1024)
				default:
					return errors.New("define either --mb or --gb")
				}
			}())
		},
	}

	cmd.Flags().Int64VarP(&gb, "gb", "", 0, "Gigabytes")
	cmd.Flags().Int64VarP(&mb, "mb", "", 0, "Megabytes")

	return cmd
}

func createRule(label string, target string) error {
	if err := touch(ruleFile(label)); err != nil {
		return err
	}

	if err := xattr.Set(ruleFile(label), "user.target", []byte(target)); err != nil {
		return err
	}

	if err := setThreshold(label, 0); err != nil {
		return err
	}

	return nil
}

func setThreshold(label string, threshold int64) error {
	return xattr.Set(ruleFile(label), "user.threshold", []byte(strconv.FormatInt(threshold, 10)))
}

// TODO: gokit/systemdinstaller to support timers and oneshot services?
func writeSystemdUnits() error {
	service := `[Unit]
Description=Low disk space checker
OnFailure=failure-notification@%n

[Service]
Type=oneshot
ExecStart=/usr/bin/jsys lowdiskspace-checker
`

	timer := `[Unit]
Description=Run Low disk space checker

[Timer]
OnBootSec=5m
OnUnitActiveSec=5m

[Install]
WantedBy=timers.target
`

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	systemdUserUnitPath := func(filename string) string {
		return filepath.Join(configDir, "systemd/user", filename)
	}

	writeFileButOnlyIfNotExists := func(path string, content string) error {
		exists, err := osutil.Exists(path)
		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("file already exists: %s", path)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}

		return nil
	}

	if err := writeFileButOnlyIfNotExists(systemdUserUnitPath("lowdiskspace-checker.service"), service); err != nil {
		return err
	}

	if err := writeFileButOnlyIfNotExists(systemdUserUnitPath("lowdiskspace-checker.timer"), timer); err != nil {
		return err
	}

	return nil
}

func ruleFile(name string) string {
	return filepath.Join(rulesDir, name)
}

// just makes an empty file
// TODO: only if one doesn't already exist?
func touch(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return file.Close()
}
