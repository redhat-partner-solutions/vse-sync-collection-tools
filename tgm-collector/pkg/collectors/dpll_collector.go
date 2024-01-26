// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

const (
	DPLLCollectorName = "DPLL"
)

// Returns a new DPLLCollector from the CollectionConstuctor Factory
func NewDPLLCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &DPLLNetlinkCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}
	dpllFSExists, err := devices.IsDPLLFileSystemPresent(ctx, constructor.PTPInterface)
	log.Debug("DPLL FS exists: ", dpllFSExists)
	if dpllFSExists && err == nil {
		return NewDPLLFilesystemCollector(constructor)
	} else {
		return NewDPLLNetlinkCollector(constructor)
	}
}

func init() {
	RegisterCollector(DPLLCollectorName, NewDPLLCollector, optional)
}
