// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"errors"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/utils"
)

const (
	expectedAntStatus        = 2
	gnssAntStatusID          = TGMSyncEnvPath + "/gnss/antenna-connected/wpc/"
	gnssAntStatusDescription = "GNSS Module is connected to an antenna"
)

type GNSSAntStatus struct {
	Blocks []*devices.GPSAntennaDetails `json:"blocks"`
}

func (gnssAnt *GNSSAntStatus) Verify() error {
	for _, block := range gnssAnt.Blocks {
		if block.Status == expectedAntStatus {
			return nil
		}
	}
	return utils.NewInvalidEnvError(errors.New("no GNSS antenna connected"))
}

func (gnssAnt *GNSSAntStatus) GetID() string {
	return gnssAntStatusID
}

func (gnssAnt *GNSSAntStatus) GetDescription() string {
	return gnssAntStatusDescription
}

func (gnssAnt *GNSSAntStatus) GetData() any { //nolint:ireturn // data will vary for each validation
	return gnssAnt
}

func (gnssAnt *GNSSAntStatus) GetOrder() int {
	return gnssConnectedToAntOrdering
}

func NewGNSSAntStatus(gpsSatus *devices.GPSDetails) *GNSSAntStatus {
	return &GNSSAntStatus{Blocks: gpsSatus.AntennaDetails}
}
