// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

const (
	deviceDriverVersionID          = TGMEnvVerPath + "/ice-driver/"
	deviceDriverVersionDescription = "Card driver is valid"
)

var (
	minDriverVersion           = "1.11.0"
	minInTreeDriverVersion     = "5.14.0-0"
	outOfTreeIceDriverSegments = 3
)

func NewDeviceDriver(ptpDevInfo *devices.PTPDeviceInfo) *VersionWithErrorCheck {
	var err error
	checkVer := ptpDevInfo.DriverVersion
	if checkVer[len(checkVer)-1] == '.' {
		checkVer = checkVer[:len(checkVer)-1]
	}
	ver := fmt.Sprintf("v%s", strings.ReplaceAll(checkVer, "_", "-"))
	if semver.IsValid(ver) {
		if semver.Compare(ver, fmt.Sprintf("v%s", minInTreeDriverVersion)) < 0 {
			err = fmt.Errorf(
				"found device driver version %s. This is below minimum version %s so likely an out of tree driver",
				ptpDevInfo.DriverVersion, minInTreeDriverVersion,
			)
		}
	} else {
		if strings.Count(ptpDevInfo.DriverVersion, ".") == outOfTreeIceDriverSegments {
			err = fmt.Errorf(
				"unable to parse device driver version (%s), likely an out of tree driver",
				ptpDevInfo.DriverVersion,
			)
		}
	}

	return &VersionWithErrorCheck{
		VersionCheck: VersionCheck{
			id:           deviceDriverVersionID,
			Version:      ptpDevInfo.DriverVersion,
			checkVersion: checkVer,
			MinVersion:   minDriverVersion,
			description:  deviceDriverVersionDescription,
			order:        deviceDriverVersionOrdering,
		},
		Error: err,
	}
}
