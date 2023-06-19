// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type PTPCollector struct {
	callback      callbacks.Callback
	running       map[string]bool
	devInfo       *devices.PTPDeviceInfo
	ctx           clients.ContainerContext
	DataTypes     [2]string
	interfaceName string
	count         uint32
}

const (
	PTPCollectorName = "PTP"

	VendorIntel = "0x8086"
	DeviceE810  = "0x1593"

	DeviceInfo = "device-info"
	DPLLInfo   = "dpll-info"
	All        = "all"

	PTPNamespace  = "openshift-ptp"
	PodNamePrefix = "linuxptp-daemon-"
	PTPContainer  = "linuxptp-daemon-container"
)

var ptpCollectables = [2]string{
	DeviceInfo,
	DPLLInfo,
}

func (ptpDev *PTPCollector) getNotCollectableError(key string) error {
	return fmt.Errorf("key %s is not a colletable of %T", key, ptpDev)
}

func (ptpDev *PTPCollector) getErrorIfNotCollectable(key string) error {
	for _, dataType := range ptpDev.DataTypes[:] {
		if dataType == key {
			return nil
		}
	}
	return ptpDev.getNotCollectableError(key)
}

// Start will add the key to the running pieces of data
// to be collects when polled
func (ptpDev *PTPCollector) Start(key string) error {
	switch key {
	case All:
		for _, dataType := range ptpDev.DataTypes[:] {
			log.Debugf("starting: %s", dataType)
			ptpDev.running[dataType] = true
		}
	default:
		err := ptpDev.getErrorIfNotCollectable(key)
		if err != nil {
			return err
		}
		ptpDev.running[key] = true
	}
	return nil
}

// fetchLine will call the requested key's function
// store the result of that function into the collectors data
// and returns a json encoded version of that data
func (ptpDev *PTPCollector) fetchLine(key string) (err error) { //nolint:funlen // allow slightly long function
	switch key {
	case DeviceInfo:
		err = ptpDev.callback.Call(ptpDev.devInfo, DeviceInfo)
		if err != nil {
			return fmt.Errorf("callback failed %w", err)
		}
	case DPLLInfo:
		dpllInfo, fetchError := devices.GetDevDPLLInfo(ptpDev.ctx, ptpDev.interfaceName)
		if fetchError != nil {
			return fmt.Errorf("failed to fetch dpllInfo %w", fetchError)
		}
		err = ptpDev.callback.Call(&dpllInfo, DPLLInfo)
		if err != nil {
			return fmt.Errorf("callback failed %w", err)
		}
	default:
		return ptpDev.getNotCollectableError(key)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (ptpDev *PTPCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
		atomic.AddUint32(&ptpDev.count, 1)
	}()
	errorsToReturn := make([]error, 0)

	for key, isRunning := range ptpDev.running {
		if isRunning {
			err := ptpDev.fetchLine(key)
			if err != nil {
				errorsToReturn = append(errorsToReturn, err)
			}
		}
	}
	resultsChan <- PollResult{
		CollectorName: PTPCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (ptpDev *PTPCollector) CleanUp(key string) error {
	switch key {
	case All:
		ptpDev.running = make(map[string]bool)
	default:
		err := ptpDev.getErrorIfNotCollectable(key)
		if err != nil {
			return err
		}
		delete(ptpDev.running, key)
	}
	return nil
}

func (ptpDev *PTPCollector) GetPollCount() int {
	return int(atomic.LoadUint32(&ptpDev.count))
}

// Returns a new PTPCollector from the CollectionConstructor Factory
// It will set the lastPoll one polling time in the past such that the initial
// request to ShouldPoll should return True
func (constructor *CollectionConstructor) NewPTPCollector() (*PTPCollector, error) {
	ctx, err := clients.NewContainerContext(constructor.Clientset, PTPNamespace, PodNamePrefix, PTPContainer)
	if err != nil {
		return &PTPCollector{}, fmt.Errorf("could not create container context %w", err)
	}

	// Build DPPInfoFetcher ahead of time call to GetPTPDeviceInfo will build the other
	err = devices.BuildPTPDeviceInfo(constructor.PTPInterface)
	if err != nil {
		return &PTPCollector{}, fmt.Errorf("failed to build fetcher for PTPDeviceInfo %w", err)
	}

	err = devices.BuildDPLLInfoFetcher(constructor.PTPInterface)
	if err != nil {
		return &PTPCollector{}, fmt.Errorf("failed to build fetcher for DPLLInfo %w", err)
	}

	ptpDevInfo, err := devices.GetPTPDeviceInfo(constructor.PTPInterface, ctx)
	if err != nil {
		return &PTPCollector{}, fmt.Errorf("failed to fetch initial DeviceInfo %w", err)
	}
	if ptpDevInfo.VendorID != VendorIntel || ptpDevInfo.DeviceID != DeviceE810 {
		return &PTPCollector{}, fmt.Errorf("NIC device is not based on E810")
	}

	collector := PTPCollector{
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
		DataTypes:     ptpCollectables,
		running:       make(map[string]bool),
		callback:      constructor.Callback,
		devInfo:       &ptpDevInfo,
	}

	return &collector, nil
}
