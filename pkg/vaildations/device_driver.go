// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
)

const deviceDriverVersionID = "Card driver is valid"

var (
	MinDriverVersion = "4.20"
)

type DeviceDriver struct {
	Version string `json:"version"`
}

func (dev *DeviceDriver) Verify() error {
	// firmwarVersionParts := strings.Split(dev.Version, " ")
	// actualVersion := fmt.Sprintf("v%s", firmwarVersionParts[0])
	// if semver.Compare(actualVersion, fmt.Sprintf("v%s", MinDriverVersion)) < 0 {
	// 	return utils.NewInvalidEnvError(
	// 		fmt.Errorf(
	// 			"invalid firmware version: %s < %s",
	// 			firmwarVersionParts[0],
	// 			MinDriverVersion,
	// 		),
	// 	)
	// }
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
