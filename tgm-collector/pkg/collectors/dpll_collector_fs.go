// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/utils"
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

// Start sets up the collector so it is ready to be polled
func (dpll *DPLLFilesystemCollector) Start() error {
	dpll.running = true
	return nil
}

// polls for the dpll info then passes it to the callback
func (dpll *DPLLFilesystemCollector) poll() error {
	dpllInfo, err := devices.GetDevDPLLFilesystemInfo(dpll.ctx, dpll.interfaceName)

	if err != nil {
		return fmt.Errorf("failed to fetch %s %w", DPLLInfo, err)
	}
	err = dpll.callback.Call(&dpllInfo, DPLLInfo)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (dpll *DPLLFilesystemCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()
	errorsToReturn := make([]error, 0)
	err := dpll.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- PollResult{
		CollectorName: DPLLFilesystemCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (dpll *DPLLFilesystemCollector) CleanUp() error {
	dpll.running = false
	return nil
}

// Returns a new DPLLFilesystemCollector from the CollectionConstuctor Factory
func NewDPLLFilesystemCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &DPLLFilesystemCollector{}, fmt.Errorf("failed to create DPLLFilesystemCollector: %w", err)
	}
	err = devices.BuildFilesystemDPLLInfoFetcher(constructor.PTPInterface)
	if err != nil {
		return &DPLLFilesystemCollector{}, fmt.Errorf("failed to build fetcher for DPLLInfo %w", err)
	}

	collector := DPLLFilesystemCollector{
		baseCollector: newBaseCollector(
			constructor.PollInterval,
			false,
			constructor.Callback,
		),
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
	}
	return &collector, nil
}
