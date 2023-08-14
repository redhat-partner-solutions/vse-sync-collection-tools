// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"errors"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	gnssStatusID          = TGMIdBaseURI + "/status/gnss/gpsfix-valid/wpc/"
	gnssStatusDescription = "GNSS Module reciving data"
)

type GNSSNavStatus struct {
	Status *devices.GPSNavStatus `json:"status"`
}

func (gnss *GNSSNavStatus) Verify() error {
	if gnss.Status.GPSFix <= 0 {
		return utils.NewInvalidEnvError(errors.New("GNSS module is not reciving data"))
	}
	return nil
}

func (gnss *GNSSNavStatus) GetID() string {
	return gnssStatusID
}

func (gnss *GNSSNavStatus) GetDescription() string {
	return gnssStatusDescription
}

func (gnss *GNSSNavStatus) GetData() any { //nolint:ireturn // data will vary for each validation
	return gnss
}

func NewGNSSNavStatus(gpsDatails *devices.GPSDetails) *GNSSNavStatus {
	return &GNSSNavStatus{Status: &gpsDatails.NavStatus}
}
