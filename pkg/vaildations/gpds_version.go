package vaildations

// SPDX-License-Identifier: GPL-2.0-or-later

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const gpsdID = "GPSD Version is valid"

var (
	MinGSPDVersion = "3.25"
)

type GPSDVersion struct {
	Version string `json:"version"`
}

func (gpsd *GPSDVersion) Verify() error {
	parts := strings.Split(gpsd.Version, " ")
	actualVersion := fmt.Sprintf("v%s", strings.ReplaceAll(parts[1], "~", "-"))
	if !semver.IsValid(actualVersion) {
		return fmt.Errorf("could not parse version %s", actualVersion)
	}
	if semver.Compare(actualVersion, fmt.Sprintf("v%s", MinGSPDVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid firmware version: %s < %s",
				parts[1],
				MinGSPDVersion,
			),
		)
	}
	return nil
}

func (gpsd *GPSDVersion) GetID() string {
	return gpsdID
}

func (gpsd *GPSDVersion) GetData() any {
	return gpsd
}

func NewGPSDVersion(gpsdVer *devices.GPSDVersion) *GPSDVersion {
	return &GPSDVersion{
		Version: gpsdVer.Version,
	}
}
