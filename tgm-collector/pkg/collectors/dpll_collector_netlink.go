// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/utils"
)

type DPLLNetlinkCollector struct {
	*baseCollector
	ctx           *clients.ContainerCreationExecContext
	interfaceName string
	clockID       int64
}

const (
	DPLLNetlinkCollectorName = "DPLL-Netlink"
	DPLLNetlinkInfo          = "dpll-info-nl"
)

// Start sets up the collector so it is ready to be polled
func (dpll *DPLLNetlinkCollector) Start() error {
	dpll.running = true
	err := dpll.ctx.CreatePodAndWait()
	if err != nil {
		return fmt.Errorf("dpll netlink collector failed to start pod: %w", err)
	}
	log.Debug("dpll.interfaceName: ", dpll.interfaceName)
	log.Debug("dpll.ctx: ", dpll.ctx)
	clockIDStuct, err := devices.GetClockID(dpll.ctx, dpll.interfaceName)
	if err != nil {
		return fmt.Errorf("dpll netlink collector failed to find clock id: %w", err)
	}
	log.Debug("clockIDStuct.ClockID: ", clockIDStuct.ClockID)
	err = devices.BuildDPLLNetlinkInfoFetcher(clockIDStuct.ClockID)
	if err != nil {
		return fmt.Errorf("failed to build fetcher for DPLLNetlinkInfo %w", err)
	}
	dpll.clockID = clockIDStuct.ClockID
	return nil
}

// polls for the dpll info then passes it to the callback
func (dpll *DPLLNetlinkCollector) poll() error {
	dpllInfo, err := devices.GetDevDPLLNetlinkInfo(dpll.ctx, dpll.clockID)

	if err != nil {
		return fmt.Errorf("failed to fetch %s %w", DPLLNetlinkInfo, err)
	}
	err = dpll.callback.Call(&dpllInfo, DPLLNetlinkInfo)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (dpll *DPLLNetlinkCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()
	errorsToReturn := make([]error, 0)
	err := dpll.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- PollResult{
		CollectorName: DPLLNetlinkCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (dpll *DPLLNetlinkCollector) CleanUp() error {
	dpll.running = false
	err := dpll.ctx.DeletePodAndWait()
	if err != nil {
		return fmt.Errorf("dpll netlink collector failed to clean up: %w", err)
	}
	return nil
}

// Returns a new DPLLNetlinkCollector from the CollectionConstuctor Factory
func NewDPLLNetlinkCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetNetlinkContext(constructor.Clientset)
	if err != nil {
		return &DPLLNetlinkCollector{}, fmt.Errorf("failed to create DPLLNetlinkCollector: %w", err)
	}

	collector := DPLLNetlinkCollector{
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
