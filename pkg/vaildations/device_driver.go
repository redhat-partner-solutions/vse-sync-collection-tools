// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const deviceDriverVersionID = "Card driver is valid"

var (
	MinDriverVersion = " 5.14.0"
)

type DeviceDriver struct {
	Version string `json:"version"`
}

func (dev *DeviceDriver) Verify() error {
	if semver.Compare(fmt.Sprintf("v%s", dev.Version), fmt.Sprintf("v%s", MinDriverVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid firmware version: %s < %s",
				dev.Version,
				MinDriverVersion,
			),
		)
	}
	return nil
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
