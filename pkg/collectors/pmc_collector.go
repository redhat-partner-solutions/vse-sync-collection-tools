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

type PMCCollector struct {
	callback      callbacks.Callback
	ctx           clients.ContainerContext
	interfaceName string
	count         uint32
	running       bool
	pollRate      float64
}

const (
	PMCCollectorName = "PMC"
	PMCInfo          = "pmc-info"
)

func (pmc *PMCCollector) GetPollRate() float64 {
	return pmc.pollRate
}

func (pmc *PMCCollector) IsAnnouncer() bool {
	return false
}

// Start sets up the collector so it is ready to be polled
func (pmc *PMCCollector) Start() error {
	pmc.running = true
	return nil
}

// polls for the pmc info then passes it to the callback
func (pmc *PMCCollector) poll() error {
	pmcInfo, err := devices.GetPMCGrandMaster(pmc.ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch %s %w", PMCInfo, err)
	}
	err = pmc.callback.Call(&pmcInfo, PMCInfo)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (pmc *PMCCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
		atomic.AddUint32(&pmc.count, 1)
	}()
	errorsToReturn := make([]error, 0)
	err := pmc.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- PollResult{
		CollectorName: PMCCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (pmc *PMCCollector) CleanUp() error {
	pmc.running = false
	return nil
}

func (pmc *PMCCollector) GetPollCount() int {
	return int(atomic.LoadUint32(&pmc.count))
}

// Returns a new PMCCollector from the CollectionConstuctor Factory
func NewPMCCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := clients.NewContainerContext(constructor.Clientset, PTPNamespace, PodNamePrefix, PTPContainer)
	if err != nil {
		return &PMCCollector{}, fmt.Errorf("could not create container context %w", err)
	}
	collector := PMCCollector{
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
		running:       false,
		callback:      constructor.Callback,
		pollRate:      constructor.PollRate,
	}

	return &collector, nil
}

func init() {
	registry.register(PMCCollectorName, NewPMCCollector)
}
