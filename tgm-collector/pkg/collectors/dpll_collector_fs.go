// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	collectorsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
)

type DPLLFilesystemCollector struct {
	*collectorsBase.ExecCollector
	interfaceName string
}

const (
	DPLLFilesystemCollectorName = "DPLL-Filesystem"
	DPLLInfo                    = "dpll-info-fs"
)

// polls for the dpll info then passes it to the callback
func (dpll *DPLLFilesystemCollector) poll() error {
	dpllInfo, err := devices.GetDevDPLLFilesystemInfo(dpll.GetContext(), dpll.interfaceName)

	if err != nil {
		return fmt.Errorf("failed to fetch %s %w", DPLLInfo, err)
	}
	err = dpll.Callback.Call(&dpllInfo, DPLLInfo)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (dpll *DPLLFilesystemCollector) Poll(resultsChan chan collectorsBase.PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()
	errorsToReturn := make([]error, 0)
	err := dpll.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- collectorsBase.PollResult{
		CollectorName: DPLLFilesystemCollectorName,
		Errors:        errorsToReturn,
	}
}

// Returns a new DPLLFilesystemCollector from the CollectionConstuctor Factory
func NewDPLLFilesystemCollector(constructor *collectorsBase.CollectionConstructor) (collectorsBase.Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &DPLLFilesystemCollector{}, fmt.Errorf("failed to create DPLLFilesystemCollector: %w", err)
	}
	ptpInterface, err := getPTPInterfaceName(constructor)
	if err != nil {
		return &DPLLFilesystemCollector{}, err
	}
	err = devices.BuildFilesystemDPLLInfoFetcher(ptpInterface)
	if err != nil {
		return &DPLLFilesystemCollector{}, fmt.Errorf("failed to build fetcher for DPLLInfo %w", err)
	}

	collector := DPLLFilesystemCollector{
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
