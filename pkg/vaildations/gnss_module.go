// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	expectedModuleName  = "ZED-F9T"
	gnssModuleIsCorrect = "GNSS module is valid"
)

type GNSSModule struct {
	Module string `json:"module"`
}

func (gnssModule *GNSSModule) Verify() error {
	if gnssModule.Module != expectedModuleName {
		return utils.NewInvalidEnvError(
			fmt.Errorf("reported gnss module is not %s", expectedModuleName),
		)
	}
	return nil
}

func (gnssModule *GNSSModule) GetID() string {
	return gnssModuleIsCorrect
}

func (gnssModule *GNSSModule) GetData() any { //nolint:ireturn // data will vary for each validation
	return gnssModule
}

func NewGNSSModule(gpsdVer *devices.GPSVersions) *GNSSModule {
	return &GNSSModule{Module: gpsdVer.Module}
}