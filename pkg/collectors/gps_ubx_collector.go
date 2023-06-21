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
	GPSCollectorName = "GPS-UBX"
	gpsNavKey        = "gpsNav"
	ubxCollectables  = [1]string{gpsNavKey}
	GPSContainer     = "gpsd"
)

type GPSCollector struct {
	callback      callbacks.Callback
	ctx           clients.ContainerContext
	DataTypes     [1]string
	interfaceName string
	running       bool
	count         uint32
	pollRate      float64
}

func (gps *GPSCollector) GetPollRate() float64 {
	return gps.pollRate
}

// Start will add the key to the running pieces of data
// to be collects when polled
func (gps *GPSCollector) Start(key string) error {
	switch key {
	case All, gpsNavKey:
		gps.running = true
	default:
		return fmt.Errorf("key %s is not a colletable of %T", key, gps)
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

	gpsNav, err := devices.GetGPSNav(gps.ctx)
	if err != nil {
		resultsChan <- PollResult{
			CollectorName: GPSCollectorName,
			Errors:        []error{err},
		}
		return
	}
	err = gps.callback.Call(&gpsNav, gpsNavKey)

	if err != nil {
		resultsChan <- PollResult{
			CollectorName: GPSCollectorName,
			Errors:        []error{err},
		}
		return
	}

	resultsChan <- PollResult{
		CollectorName: GPSCollectorName,
		Errors:        []error{},
	}
}

// CleanUp stops a running collector
func (gps *GPSCollector) CleanUp(key string) error {
	switch key {
	case All, gpsNavKey:
		gps.running = false
	default:
		return fmt.Errorf("key %s is not a colletable of %T", key, gps)
	}
	return nil
}

func (gps *GPSCollector) GetPollCount() int {
	return int(atomic.LoadUint32(&gps.count))
}

// Returns a new PTPCollector from the CollectionConstructor Factory
// It will set the lastPoll one polling time in the past such that the initial
// request to ShouldPoll should return True
func (constructor *CollectionConstructor) NewGPSCollector() (*GPSCollector, error) {
	ctx, err := clients.NewContainerContext(constructor.Clientset, PTPNamespace, PodNamePrefix, GPSContainer)
	if err != nil {
		return &GPSCollector{}, fmt.Errorf("could not create container context %w", err)
	}

	collector := GPSCollector{
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
		DataTypes:     ubxCollectables,
		running:       false,
		callback:      constructor.Callback,
		pollRate:      constructor.PollRate,
	}

	return &collector, nil
}
