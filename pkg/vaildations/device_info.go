// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const deviceDetailsID = "Card is valid NIC"

var (
	VendorIntel        = "0x8086"
	DeviceE810         = "0x1593"
	MinFirmwareVersion = "4.20"
)

type DeviceDetails struct {
	VendorID        string `json:"vendorId"`
	DeviceID        string `json:"deviceId"`
	FirmwareVersion string
}

func (dev *DeviceDetails) Verify() error {
	if dev.VendorID != VendorIntel || dev.DeviceID != DeviceE810 {
		return utils.NewInvalidEnvError(fmt.Errorf("NIC device is not based on E810"))
	}

	firmwarVersionParts := strings.Split(dev.FirmwareVersion, " ")
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

func (dev *DeviceDetails) GetID() string {
	return deviceDetailsID
}

func (dev *DeviceDetails) GetData() any { //nolint:ireturn // data will very for each validation
	return dev
}

func NewDeviceDetails(ptpDevInfo *devices.PTPDeviceInfo) *DeviceDetails {
	return &DeviceDetails{
		VendorID:        ptpDevInfo.VendorID,
		DeviceID:        ptpDevInfo.DeviceID,
		FirmwareVersion: ptpDevInfo.FirmwareVersion,
	}
}
