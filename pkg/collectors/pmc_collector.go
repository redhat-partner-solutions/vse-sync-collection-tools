// SPDX-License-Identifier: GPL-2.0-or-later

package collectors //nolint:dupl // new collector

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	PMCCollectorName = "PMC"
	PMCInfo          = "pmc-info"
)

type PMCCollector struct {
	callback     callbacks.Callback
	ctx          clients.ContainerContext
	running      bool
	pollInterval int
}

func (pmc *PMCCollector) GetPollInterval() int {
	return pmc.pollInterval
}

func (pmc *PMCCollector) IsAnnouncer() bool {
	return false
}

// Start sets up the collector so it is ready to be polled
func (pmc *PMCCollector) Start() error {
	pmc.running = true
	return nil
}

func (pmc *PMCCollector) poll() error {
	gmSetting, err := devices.GetPMC(pmc.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch  %s %w", PMCInfo, err)
	}
	err = pmc.callback.Call(&gmSetting, PMCInfo)
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

// Returns a new PMCCollector based on values in the CollectionConstructor
func NewPMCCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &PMCCollector{}, fmt.Errorf("failed to create PMCCollector: %w", err)
	}

	collector := PMCCollector{
		ctx:          ctx,
		running:      false,
		callback:     constructor.Callback,
		pollInterval: constructor.PollInterval,
	}

	return &collector, nil
}

func init() {
	RegisterCollector(PMCCollectorName, NewPMCCollector, false, true)
}
