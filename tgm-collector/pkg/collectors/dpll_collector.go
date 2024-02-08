// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	collectorsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
)

const (
	DPLLCollectorName = "DPLL"
)

// Returns a new DPLLCollector from the CollectionConstuctor Factory
func NewDPLLCollector(constructor *collectorsBase.CollectionConstructor) (collectorsBase.Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &DPLLNetlinkCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}
	ptpInterface, err := getPTPInterfaceName(constructor)
	if err != nil {
		return &DPLLNetlinkCollector{}, err
	}
	dpllFSExists, err := devices.IsDPLLFileSystemPresent(ctx, ptpInterface)
	log.Debug("DPLL FS exists: ", dpllFSExists)
	if dpllFSExists && err == nil {
		return NewDPLLFilesystemCollector(constructor)
	} else {
		return NewDPLLNetlinkCollector(constructor)
	}
}

func init() {
	collectorsBase.RegisterCollector(DPLLCollectorName, NewDPLLCollector, collectorsBase.Optional)
}
