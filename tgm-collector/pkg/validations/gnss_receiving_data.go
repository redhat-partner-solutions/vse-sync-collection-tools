// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"errors"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
)

const (
	gnssStatusID          = TGMSyncEnvPath + "/gnss/gpsfix-valid/wpc/"
	gnssStatusDescription = "GNSS Module receiving data"
)

type GNSSNavStatus struct {
	Status *devices.GPSNavStatus `json:"status"`
}

func (gnss *GNSSNavStatus) Verify() error {
	if gnss.Status.GPSFix <= 0 {
		return utils.NewInvalidEnvError(errors.New("GNSS module is not receiving data"))
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

func (gnss *GNSSNavStatus) GetOrder() int {
	return gnssReceivingDataOrdering
}

func NewGNSSNavStatus(gpsDatails *devices.GPSDetails) *GNSSNavStatus {
	return &GNSSNavStatus{Status: &gpsDatails.NavStatus}
}
