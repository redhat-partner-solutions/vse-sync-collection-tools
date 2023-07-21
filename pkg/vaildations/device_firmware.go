// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

const deviceFirmwareID = "Card firmware is valid"

var (
	MinFirmwareVersion = "4.20"
)

func NewDeviceFirmware(ptpDevInfo *devices.PTPDeviceInfo) *VersionCheck {
	parts := strings.Split(ptpDevInfo.FirmwareVersion, " ")
	return &VersionCheck{
		id:           deviceFirmwareID,
		Version:      ptpDevInfo.FirmwareVersion,
		checkVersion: parts[0],
		minVersion:   MinFirmwareVersion,
	}
}
