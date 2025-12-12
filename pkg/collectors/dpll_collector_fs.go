// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

type DPLLFilesystemCollector struct {
	*baseCollector

	ctx           clients.ExecContext
	interfaceName string
}

const (
	DPLLFilesystemCollectorName = "DPLL-Filesystem"
	DPLLInfo                    = "dpll-info-fs"
)

// polls for the dpll info then passes it to the callback
func dpllFSPoller(dpll *DPLLFilesystemCollector) func() (callbacks.OutputType, error) {
	return func() (callbacks.OutputType, error) {
		return devices.GetDevDPLLFilesystemInfo(dpll.ctx, dpll.interfaceName) //nolint:wrapcheck //no point wrapping this
	}
}

// Returns a new DPLLFilesystemCollector from the CollectionConstuctor Factory
func NewDPLLFilesystemCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset, constructor.PTPNodeName)
	if err != nil {
		return &DPLLFilesystemCollector{}, fmt.Errorf("failed to create DPLLFilesystemCollector: %w", err)
	}

	err = devices.BuildFilesystemDPLLInfoFetcher(constructor.PTPInterface)
	if err != nil {
		return &DPLLFilesystemCollector{}, fmt.Errorf("failed to build fetcher for DPLLInfo %w", err)
	}

	collector := &DPLLFilesystemCollector{
		baseCollector: newBaseCollector(
			constructor.PollInterval,
			false,
			constructor.Callback,
			DPLLFilesystemCollectorName,
			DPLLInfo,
		),
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
	}
	collector.poller = dpllFSPoller(collector)

	return collector, nil
}
