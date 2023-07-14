// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const gnssID = "GNSS Version is valid"

var (
	MinGNSSVersion = "2.20"
)

type GNSSVersion struct {
	Version string `json:"version"`
}

func (gnss *GNSSVersion) Verify() error {
	parts := strings.Split(gnss.Version, " ")
	actualVersion := fmt.Sprintf("v%s", parts[1])
	if !semver.IsValid(actualVersion) {
		return fmt.Errorf("could not parse version %s", actualVersion)
	}
	if semver.Compare(actualVersion, fmt.Sprintf("v%s", MinGNSSVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid firmware version: %s < %s",
				parts[1],
				MinGNSSVersion,
			),
		)
	}
	return nil
}

func (gnss *GNSSVersion) GetID() string {
	return gnssID
}

func (gnss *GNSSVersion) GetData() any { //nolint:ireturn // data will very for each validation
	return gnss
}

func NewGNSS(gnss *devices.GPSNav) *GNSSVersion {
	return &GNSSVersion{
		Version: gnss.FirmwareVersion,
	}
}
