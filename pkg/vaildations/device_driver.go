// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

const deviceDriverVersionID = "Card driver is valid"

var (
	MinDriverVersion           = "5.14.0"
	outOfTreeIceDriverSegments = 4
)

func NewDeviceDriver(ptpDevInfo *devices.PTPDeviceInfo) *VersionWithErrorCheck {
	var err error
	ver := fmt.Sprintf("v%s", ptpDevInfo.DriverVersion)
	if !semver.IsValid(ver) {
		if strings.Count(ptpDevInfo.DriverVersion, ".") == outOfTreeIceDriverSegments {
			err = fmt.Errorf(
				"unable to parse version (%s), likely an out of tree driver",
				ptpDevInfo.DriverVersion,
			)
		}
	}

	return &VersionWithErrorCheck{
		VersionCheck: VersionCheck{
			id:           deviceDriverVersionID,
			Version:      ptpDevInfo.DriverVersion,
			checkVersion: ptpDevInfo.DriverVersion,
			minVersion:   MinDriverVersion,
		},
		Error: err,
	}
}
