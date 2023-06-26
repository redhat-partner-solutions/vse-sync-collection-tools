// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"
	"sync/atomic"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type DPLLCollector struct {
	callback      callbacks.Callback
	ctx           clients.ContainerContext
	interfaceName string
	count         uint32
	running       bool
	pollRate      float64
}

const (
	DPLLCollectorName = "DPLL"

	DPLLInfo = "dpll-info"
	All      = "all"
)

func (dpll *DPLLCollector) GetPollRate() float64 {
	return dpll.pollRate
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
func (dpll *DPLLCollector) poll() []error {
	dpllInfo, err := devices.GetDevDPLLInfo(dpll.ctx, dpll.interfaceName)

	if err != nil {
		return []error{fmt.Errorf("failed to fetch dpllInfo %w", err)}
	}
	err = dpll.callback.Call(&dpllInfo, DPLLInfo)
	if err != nil {
		return []error{fmt.Errorf("callback failed %w", err)}
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (dpll *DPLLCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
		atomic.AddUint32(&dpll.count, 1)
	}()
	errorsToReturn := dpll.poll()
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

func (dpll *DPLLCollector) GetPollCount() int {
	return int(atomic.LoadUint32(&dpll.count))
}

// Returns a new DPLLCollector from the CollectionConstuctor Factory
func NewDPLLCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := clients.NewContainerContext(constructor.Clientset, PTPNamespace, PodNamePrefix, PTPContainer)
	if err != nil {
		return &DPLLCollector{}, fmt.Errorf("could not create container context %w", err)
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
		pollRate:      constructor.PollRate,
	}

	return &collector, nil
}

func init() {
	registry.register(DPLLCollectorName, NewDPLLCollector)
}
