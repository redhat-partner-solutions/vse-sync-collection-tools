// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type ExpectedDeviceDetails struct {
	VendorID string
	DeviceID string
}

// func GetComparison(ptpDevInfo *devices.PTPDeviceInfo) map[string][2]string {
// 	comparison := make(map[string][2]string)
// 	comparison["VendorID"] = [2]string{ExpectedEnv.VendorID, ptpDevInfo.VendorID}
// 	comparison["DeviceID"] = [2]string{ExpectedEnv.DeviceID, ptpDevInfo.DeviceID}
// 	return comparison
// }

var (
	VendorIntel = "0x8086"
	DeviceE810  = "0x1593"
	ExpectedDev *ExpectedDeviceDetails
)

func init() {
	ExpectedDev = &ExpectedDeviceDetails{
		DeviceID: DeviceE810,
		VendorID: VendorIntel,
	}
}

func VerifyDeviceInfo(ptpDevInfo *devices.PTPDeviceInfo) error {
	if ptpDevInfo.VendorID != ExpectedDev.VendorID || ptpDevInfo.DeviceID != ExpectedDev.DeviceID {
		return utils.NewInvalidEnvError(fmt.Errorf("NIC device is not based on E810"))
	}
	return nil
}
