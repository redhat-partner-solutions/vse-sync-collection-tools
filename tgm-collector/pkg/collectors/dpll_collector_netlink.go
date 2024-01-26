// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/clients"
	collectorsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
)

type DPLLNetlinkCollector struct {
	*collectorsBase.ExecCollector
	interfaceName string
	clockID       int64
}

const (
	DPLLNetlinkCollectorName = "DPLL-Netlink"
	DPLLNetlinkInfo          = "dpll-info-nl"
)

// Start sets up the collector so it is ready to be polled
func (dpll *DPLLNetlinkCollector) Start() error {
	err := dpll.ExecCollector.Start()
	if err != nil {
		return fmt.Errorf("failed to start dpll netlink collector: %w", err)
	}

	ctx, ok := dpll.GetContext().(*clients.ContainerCreationExecContext)
	if !ok {
		return fmt.Errorf("dpll netlink collector has an incorrect context type")
	}

	err = ctx.CreatePodAndWait()
	if err != nil {
		return fmt.Errorf("dpll netlink collector failed to start pod: %w", err)
	}
	log.Debug("dpll.interfaceName: ", dpll.interfaceName)
	log.Debug("dpll.ctx: ", ctx)
	clockIDStuct, err := devices.GetClockID(ctx, dpll.interfaceName)
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
	dpllInfo, err := devices.GetDevDPLLNetlinkInfo(dpll.GetContext(), dpll.clockID)

	if err != nil {
		return fmt.Errorf("failed to fetch %s %w", DPLLNetlinkInfo, err)
	}
	err = dpll.Callback.Call(&dpllInfo, DPLLNetlinkInfo)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (dpll *DPLLNetlinkCollector) Poll(resultsChan chan collectorsBase.PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()
	errorsToReturn := make([]error, 0)
	err := dpll.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- collectorsBase.PollResult{
		CollectorName: DPLLNetlinkCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (dpll *DPLLNetlinkCollector) CleanUp() error {
	err := dpll.ExecCollector.CleanUp()
	if err != nil {
		return fmt.Errorf("failed to cleanly stop dpll collector: %w", err)
	}

	ctx, ok := dpll.GetContext().(*clients.ContainerCreationExecContext)
	if !ok {
		return fmt.Errorf("dpll netlink collector has an incorrect context type")
	}

	err = ctx.DeletePodAndWait()
	if err != nil {
		return fmt.Errorf("dpll netlink collector failed to clean up: %w", err)
	}
	return nil
}

// Returns a new DPLLNetlinkCollector from the CollectionConstuctor Factory
func NewDPLLNetlinkCollector(constructor *collectorsBase.CollectionConstructor) (collectorsBase.Collector, error) {
	ctx, err := contexts.GetNetlinkContext(constructor.Clientset)
	if err != nil {
		return &DPLLNetlinkCollector{}, fmt.Errorf("failed to create DPLLNetlinkCollector: %w", err)
	}
	ptpInterface, err := getPTPInterfaceName(constructor)
	if err != nil {
		return &DPLLNetlinkCollector{}, err
	}

	collector := DPLLNetlinkCollector{
		ExecCollector: collectorsBase.NewExecCollector(
			constructor.PollInterval,
			false,
			constructor.Callback,
			ctx,
		),
		interfaceName: ptpInterface,
	}

	return &collector, nil
}
