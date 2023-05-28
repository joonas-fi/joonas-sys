package sysfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/function61/gokit/os/osutil"
)

type PowerSupplyItem struct {
	CAPACITY_LEVEL sysfsClassPowerCapacityLevel
	MODEL_NAME     string
}

// `id` is the entry under directory '/sys/class/power_supply'
// https://i3wm.org/docs/i3status.html#_battery
func ReadPowerSupplyCapacity(id string) (*PowerSupplyItem, error) {
	dir := PowerSupplyDir(filepath.Join("/sys/class/power_supply", id))
	return dir.ReadUevent()
}

// points to "/sys/class/power_supply/<id>"
type PowerSupplyDir string

func (p PowerSupplyDir) ReadUevent() (*PowerSupplyItem, error) {
	/*
	   	example file:

	   POWER_SUPPLY_NAME=hidpp_battery_0
	   POWER_SUPPLY_TYPE=Battery
	   POWER_SUPPLY_ONLINE=1
	   POWER_SUPPLY_STATUS=Discharging
	   POWER_SUPPLY_SCOPE=Device
	   POWER_SUPPLY_MODEL_NAME=MX Ergo Multi-Device Trackball
	   POWER_SUPPLY_MANUFACTURER=Logitech
	   POWER_SUPPLY_SERIAL_NUMBER=406f-80-58-20-27
	   POWER_SUPPLY_CAPACITY_LEVEL=Full
	*/
	powerSupplyUevent, err := os.ReadFile(filepath.Join(string(p), "uevent"))
	if err != nil {
		return nil, err
	}

	capacityStr := ""
	model := ""

	for _, line := range strings.Split(string(powerSupplyUevent), "\n") {
		if line == "" {
			continue
		}

		key, valueRaw := osutil.ParseEnv(line)
		if key == "" {
			return nil, fmt.Errorf("readPowerSupplies: failed parsing line '%s'", line)
		}

		switch key {
		case "POWER_SUPPLY_CAPACITY_LEVEL":
			capacityStr = valueRaw
		case "POWER_SUPPLY_MODEL_NAME":
			model = valueRaw
		}
	}

	/*
		if model != findModelName {
			return nil, fmt.Errorf("not expected model name ('%s'), got: '%s'", findModelName, model)
		}
	*/

	if capacityStr == "" {
		return nil, errors.New("capacity not reported")
	}

	capacity, err := parseSysfsClassPowerCapacityLevel(capacityStr)
	if err != nil {
		return nil, err
	}

	return &PowerSupplyItem{
		CAPACITY_LEVEL: capacity,
		MODEL_NAME:     model,
	}, nil
}

// "Coarse representation of battery capacity." - https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-class-power
type sysfsClassPowerCapacityLevel string

// Unknown < Critical < Low < Normal < High < Full
// ________________________  <-- below normal
func (s sysfsClassPowerCapacityLevel) IsBelowNormal() bool {
	switch s {
	case sysfsClassPowerCapacityLevelUnknown, sysfsClassPowerCapacityLevelCritical, sysfsClassPowerCapacityLevelLow:
		return true
	default:
		return false
	}
}

const (
	sysfsClassPowerCapacityLevelUnknown  sysfsClassPowerCapacityLevel = "Unknown"
	sysfsClassPowerCapacityLevelCritical sysfsClassPowerCapacityLevel = "Critical"
	sysfsClassPowerCapacityLevelLow      sysfsClassPowerCapacityLevel = "Low"
	sysfsClassPowerCapacityLevelNormal   sysfsClassPowerCapacityLevel = "Normal"
	sysfsClassPowerCapacityLevelHigh     sysfsClassPowerCapacityLevel = "High"
	sysfsClassPowerCapacityLevelFull     sysfsClassPowerCapacityLevel = "Full"
)

func parseSysfsClassPowerCapacityLevel(raw string) (sysfsClassPowerCapacityLevel, error) {
	cast := sysfsClassPowerCapacityLevel(raw)
	switch cast {
	case sysfsClassPowerCapacityLevelUnknown, sysfsClassPowerCapacityLevelCritical, sysfsClassPowerCapacityLevelLow, sysfsClassPowerCapacityLevelNormal, sysfsClassPowerCapacityLevelHigh, sysfsClassPowerCapacityLevelFull:
		return cast, nil
	default:
		return sysfsClassPowerCapacityLevelUnknown, fmt.Errorf("parseSysfsClassPowerCapacityLevel: unsupported value '%s'", raw)
	}
}
