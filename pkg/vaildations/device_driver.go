// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
)

const deviceDriverVersionID = "Card driver is valid"

var (
	MinDriverVersion = " 5.14.0"
)

type DeviceDriver struct {
	Version string `json:"version"`
}

func (dev *DeviceDriver) Verify() error {
	return checkVersion(dev.Version, MinDriverVersion)
}

func (dev *DeviceDriver) GetID() string {
	return deviceDriverVersionID
}

func (dev *DeviceDriver) GetData() any { //nolint:ireturn // data will very for each validation
	return dev
}

func NewDeviceDriver(ptpDevInfo *devices.PTPDeviceInfo) *DeviceDriver {
	return &DeviceDriver{
		Version: ptpDevInfo.DriverVersion,
	}
}
