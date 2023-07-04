// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const deviceDetailsID = "Card is valid NIC"

var (
	VendorIntel = "0x8086"
	DeviceE810  = "0x1593"
)

type DeviceDetails struct {
	VendorID         string `json:"vendorId"`
	ExpectedVendorID string `json:"expectedVendorId"`
	DeviceID         string `json:"deviceId"`
	ExpectedDeviceID string `json:"expectedDeviceId"`
}

func (dev *DeviceDetails) Verify() error {
	if dev.VendorID != dev.ExpectedVendorID || dev.DeviceID != dev.ExpectedDeviceID {
		return utils.NewInvalidEnvError(fmt.Errorf("NIC device is not based on E810"))
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
		VendorID:         ptpDevInfo.VendorID,
		ExpectedVendorID: VendorIntel,
		DeviceID:         ptpDevInfo.DeviceID,
		ExpectedDeviceID: DeviceE810,
	}
}
