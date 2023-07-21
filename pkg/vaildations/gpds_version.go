package vaildations

// SPDX-License-Identifier: GPL-2.0-or-later

import (
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

const gpsdID = "GPSD Version is valid"

var (
	MinGSPDVersion = "3.25"
)

func NewGPSDVersion(gpsdVer *devices.GPSDVersion) *VersionCheck {
	parts := strings.Split(gpsdVer.Version, " ")
	return &VersionCheck{
		id:           gpsdID,
		Version:      gpsdVer.Version,
		checkVersion: strings.ReplaceAll(parts[1], "~", "-"),
		minVersion:   MinGSPDVersion,
	}
}
