// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type DPLLNetlinkCollector struct {
	*baseCollector
	ctx               *clients.ContainerCreationExecContext
	interfaceName     string
	params            devices.NetlinkParameters
	unmanagedDebugPod bool
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
	netlinkParams, err := devices.GetNetlinkParameters(dpll.ctx, dpll.interfaceName)
	if err != nil {
		return fmt.Errorf("dpll netlink collector failed to find clock id: %w", err)
	}
	log.Debug("clockIDStuct.ClockID: ", netlinkParams.ClockID)
	err = devices.BuildDPLLNetlinkDeviceFetcher(netlinkParams)
	if err != nil {
		return fmt.Errorf("failed to build fetcher for DPLLNetlinkInfo %w", err)
	}
	dpll.params = netlinkParams
	return nil
}

// polls for the dpll info then passes it to the callback
func dpllNetlinkPoller(dpll *DPLLNetlinkCollector) func() (callbacks.OutputType, error) {
	return func() (callbacks.OutputType, error) {
		return devices.GetDevDPLLNetlinkInfo(dpll.ctx, dpll.params) //nolint:wrapcheck //no point wrapping this
	}
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (dpll *DPLLNetlinkCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer wg.Done()
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
	ctx, err := contexts.GetNetlinkContext(
		constructor.Clientset,
		constructor.PTPNodeName,
		constructor.UnmanagedDebugPod,
	)
	if err != nil {
		return &DPLLNetlinkCollector{}, fmt.Errorf("failed to create DPLLNetlinkCollector: %w", err)
	}

	collector := &DPLLNetlinkCollector{
		baseCollector: newBaseCollector(
			constructor.PollInterval,
			false,
			constructor.Callback,
			DPLLNetlinkCollectorName,
			DPLLNetlinkInfo,
		),
		interfaceName:     constructor.PTPInterface,
		ctx:               ctx,
		unmanagedDebugPod: constructor.UnmanagedDebugPod,
	}
	collector.poller = dpllNetlinkPoller(collector)
	err = collector.Start()
	if err != nil {
		collector.CleanUp()
	}
	return collector, err
}
