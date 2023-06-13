// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
)

var (
	GPSCollectorName = "GPS-UBX"
	gpsNavKey        = "gpsNav"
	ubxCollectables  = [1]string{gpsNavKey}
	GPSContainer     = "gpsd"
)

type GPSCollector struct {
	lastPoll        time.Time
	callback        callbacks.Callback
	data            devices.GPSNav
	ctx             clients.ContainerContext
	DataTypes       [1]string
	interfaceName   string
	inversePollRate float64
	running         bool
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

// ShouldPoll checks if enough time has passed since the last poll
func (gps *GPSCollector) ShouldPoll() bool {
	log.Debugf("since: %v", time.Since(gps.lastPoll).Seconds())
	log.Debugf("wait: %v", gps.inversePollRate)
	return time.Since(gps.lastPoll).Seconds() >= gps.inversePollRate
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (gps *GPSCollector) Poll() []error {
	gpsNav, err := devices.GetGPSNav(gps.ctx)
	if err != nil {
		return []error{err}
	}
	gps.data = gpsNav
	line, err := json.Marshal(gpsNav)
	if err != nil {
		return []error{err}
	} else {
		err = gps.callback.Call(fmt.Sprintf("%T", gps), gpsNavKey, string(line))
		if err != nil {
			return []error{err}
		}
	}

	gps.lastPoll = time.Now()
	return []error{}
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

// Returns a new PTPCollector from the CollectionConstuctor Factory
// It will set the lastPoll one polling time in the past such that the initial
// request to ShouldPoll should return True
func (constuctor *CollectionConstuctor) NewGPSCollector() (*GPSCollector, error) {
	ctx, err := clients.NewContainerContext(constuctor.Clientset, PTPNamespace, PodNamePrefix, GPSContainer)
	if err != nil {
		return &GPSCollector{}, fmt.Errorf("could not create container context %w", err)
	}

	inversePollRate := 1.0 / constuctor.PollRate
	offset := time.Duration(float64(time.Second.Nanoseconds()) * inversePollRate)

	collector := GPSCollector{
		interfaceName:   constuctor.PTPInterface,
		ctx:             ctx,
		DataTypes:       ubxCollectables,
		data:            devices.GPSNav{},
		running:         false,
		callback:        constuctor.Callback,
		inversePollRate: inversePollRate,
		lastPoll:        time.Now().Add(-offset), // Subtract off a polling time so the first poll hits
	}

	return &collector, nil
}
