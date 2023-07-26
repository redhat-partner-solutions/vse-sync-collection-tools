// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"errors"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	hadGNSSDevices = "Has GNSS Devices"
)

type GNSDevices struct {
	Paths []string `json:"paths"`
}

func (gnssDevices *GNSDevices) Verify() error {
	if len(gnssDevices.Paths) == 0 {
		return utils.NewInvalidEnvError(errors.New("no gnss devices found"))
	}
	return nil
}

func (gnssDevices *GNSDevices) GetID() string {
	return hadGNSSDevices
}

func (gnssDevices *GNSDevices) GetData() any { //nolint:ireturn // data will vary for each validation
	return gnssDevices
}

func NewGNSDevices(gpsdVer *devices.GPSVersions) *GNSDevices {
	return &GNSDevices{Paths: gpsdVer.GNSSDevices}
}
