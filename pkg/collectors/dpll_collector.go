// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type DPLLCollector struct {
	callback      callbacks.Callback
	ctx           clients.ContainerContext
	interfaceName string
	running       bool
	pollInterval  int
}

const (
	DPLLCollectorName = "DPLL"
	DPLLInfo          = "dpll-info"
)

func (dpll *DPLLCollector) GetPollInterval() int {
	return dpll.pollInterval
}

func (dpll *DPLLCollector) IsAnnouncer() bool {
	return false
}

// Start sets up the collector so it is ready to be polled
func (dpll *DPLLCollector) Start() error {
	dpll.running = true
	return nil
}

// polls for the dpll info then passes it to the callback
func (dpll *DPLLCollector) poll() error {
	dpllInfo, err := devices.GetDevDPLLInfo(dpll.ctx, dpll.interfaceName)

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
func (dpll *DPLLCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()
	errorsToReturn := make([]error, 0)
	err := dpll.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- PollResult{
		CollectorName: DPLLCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (dpll *DPLLCollector) CleanUp() error {
	dpll.running = false
	return nil
}

// Returns a new DPLLCollector from the CollectionConstuctor Factory
func NewDPLLCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &DPLLCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}
	err = devices.BuildDPLLInfoFetcher(constructor.PTPInterface)
	if err != nil {
		return &DPLLCollector{}, fmt.Errorf("failed to build fetcher for DPLLInfo %w", err)
	}

	collector := DPLLCollector{
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
		running:       false,
		callback:      constructor.Callback,
		pollInterval:  constructor.PollInterval,
	}

	return &collector, nil
}

func init() {
	RegisterCollector(DPLLCollectorName, NewDPLLCollector, false, true)
}
