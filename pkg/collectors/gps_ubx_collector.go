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

var (
	GPSCollectorName = "GNSS"
	gpsNavKey        = "gpsNav"
	GPSContainer     = "gpsd"
)

type GPSCollector struct {
	callback      callbacks.Callback
	ctx           clients.ContainerContext
	interfaceName string
	running       bool
	count         uint32
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
		atomic.AddUint32(&gps.count, 1)
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

func (gps *GPSCollector) GetPollCount() int {
	return int(atomic.LoadUint32(&gps.count))
}

// Returns a new GPSCollector based on values in the CollectionConstructor
func NewGPSCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := clients.NewContainerContext(constructor.Clientset, PTPNamespace, PodNamePrefix, GPSContainer)
	if err != nil {
		return &GPSCollector{}, fmt.Errorf("could not create container context %w", err)
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
	RegisterCollector(GPSCollectorName, NewGPSCollector)
}
