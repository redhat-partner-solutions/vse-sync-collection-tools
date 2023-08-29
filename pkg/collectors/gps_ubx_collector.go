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

var (
	GPSCollectorName = "GNSS"
	gpsNavKey        = "gpsNav"
)

type GPSCollector struct {
	callback      callbacks.Callback
	ctx           clients.ContainerContext
	interfaceName string
	running       bool
	pollInterval  int
}

func (gps *GPSCollector) GetPollInterval() int {
	return gps.pollInterval
}

func (gps *GPSCollector) IsAnnouncer() bool {
	return false
}

// Start sets up the collector so it is ready to be polled
func (gps *GPSCollector) Start() error {
	gps.running = true
	return nil
}

func (gps *GPSCollector) poll() error {
	gpsNav, err := devices.GetGPSNav(gps.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch  %s %w", gpsNavKey, err)
	}
	err = gps.callback.Call(&gpsNav, gpsNavKey)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (gps *GPSCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()

	errorsToReturn := make([]error, 0)
	err := gps.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- PollResult{
		CollectorName: GPSCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (gps *GPSCollector) CleanUp() error {
	gps.running = false
	return nil
}

// Returns a new GPSCollector based on values in the CollectionConstructor
func NewGPSCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &GPSCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}

	collector := GPSCollector{
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
		running:       false,
		callback:      constructor.Callback,
		pollInterval:  constructor.PollInterval,
	}

	return &collector, nil
}

func init() {
	RegisterCollector(GPSCollectorName, NewGPSCollector, includeByDefault)
}
