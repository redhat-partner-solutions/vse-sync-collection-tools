package vaildations

// SPDX-License-Identifier: GPL-2.0-or-later

import (
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

const (
	gpsdID         = "GPSD Version is valid"
	MinGSPDVersion = "3.25"
)

func NewGPSDVersion(gpsdVer *devices.GPSVersions) *VersionCheck {
	parts := strings.Split(gpsdVer.GPSDVersion, " ")
	return &VersionCheck{
		id:           gpsdID,
		Version:      gpsdVer.GPSDVersion,
		checkVersion: strings.ReplaceAll(parts[0], "~", "-"),
		minVersion:   MinGSPDVersion,
	}
}
