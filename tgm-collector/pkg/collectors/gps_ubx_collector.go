// SPDX-License-Identifier: GPL-2.0-or-later

package collectors //nolint:dupl // new collector

import (
	"fmt"

	collectorsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
)

var (
	GPSCollectorName = "GNSS"
	gpsNavKey        = "gpsNav"
)

type GPSCollector struct {
	*collectorsBase.ExecCollector
	interfaceName string
}

func (gps *GPSCollector) poll() error {
	gpsNav, err := devices.GetGPSNav(gps.GetContext())
	if err != nil {
		return fmt.Errorf("failed to fetch  %s %w", gpsNavKey, err)
	}
	err = gps.Callback.Call(&gpsNav, gpsNavKey)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (gps *GPSCollector) Poll(resultsChan chan collectorsBase.PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()

	errorsToReturn := make([]error, 0)
	err := gps.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- collectorsBase.PollResult{
		CollectorName: GPSCollectorName,
		Errors:        errorsToReturn,
	}
}

// Returns a new GPSCollector based on values in the CollectionConstructor
func NewGPSCollector(constructor *collectorsBase.CollectionConstructor) (collectorsBase.Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &GPSCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}
	ptpInterface, err := getPTPInterfaceName(constructor)
	if err != nil {
		return &GPSCollector{}, err
	}
	collector := GPSCollector{
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

func init() {
	collectorsBase.RegisterCollector(GPSCollectorName, NewGPSCollector, collectorsBase.Optional)
}
