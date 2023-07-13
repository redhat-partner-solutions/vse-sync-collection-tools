// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
	"golang.org/x/mod/semver"
)

const deviceFirmwareID = "Card firmware is valid"

var (
	MinFirmwareVersion = "4.20"
)

type DeviceFirmware struct {
	Version string `json:"version"`
}

func (dev *DeviceFirmware) Verify() error {
	firmwarVersionParts := strings.Split(dev.Version, " ")
	actualVersion := fmt.Sprintf("v%s", firmwarVersionParts[0])
	if semver.Compare(actualVersion, fmt.Sprintf("v%s", MinFirmwareVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid firmware version: %s < %s",
				firmwarVersionParts[0],
				MinFirmwareVersion,
			),
		)
	}
	return nil
}

func (dev *DeviceFirmware) GetID() string {
	return deviceFirmwareID
}

func (dev *DeviceFirmware) GetData() any { //nolint:ireturn // data will very for each validation
	return dev
}

func NewDeviceFirmware(ptpDevInfo *devices.PTPDeviceInfo) *DeviceFirmware {
	return &DeviceFirmware{
		Version: ptpDevInfo.FirmwareVersion,
	}
}
