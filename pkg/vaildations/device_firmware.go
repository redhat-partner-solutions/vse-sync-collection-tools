// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
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
	return checkVersion(firmwarVersionParts[0], MinFirmwareVersion)
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
