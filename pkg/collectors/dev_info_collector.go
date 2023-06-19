// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type DevInfoCollector struct {
	callback      callbacks.Callback
	devInfo       *devices.PTPDeviceInfo
	quit          chan os.Signal
	erroredPolls  chan PollResult
	requiresFetch chan bool
	ctx           clients.ContainerContext
	interfaceName string
	count         uint32
	running       bool
}

const (
	DevInfoCollectorName = "DevInfo"
	DeviceInfo           = "device-info"
	VendorIntel          = "0x8086"
	DeviceE810           = "0x1593"
)

// Start will add the key to the running pieces of data
// to be collects when polled
func (ptpDev *DevInfoCollector) Start(key string) error {
	switch key {
	case All, DeviceInfo:
		ptpDev.running = true
		go ptpDev.monitorErroredPolls()
	default:
		return fmt.Errorf("key %s is not a colletable of %T", key, ptpDev)
	}
	return nil
}

func (ptpDev *DevInfoCollector) monitorErroredPolls() {
	for {
		select {
		case <-ptpDev.quit:
			return
		case <-ptpDev.erroredPolls:
			ptpDev.requiresFetch <- true
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

// polls for the device info, stores it then passes it to the callback
func (ptpDev *DevInfoCollector) poll() []error {
	var devInfo *devices.PTPDeviceInfo
	select {
	case <-ptpDev.requiresFetch:
		fetchedDevInfo, err := devices.GetPTPDeviceInfo(ptpDev.interfaceName, ptpDev.ctx)
		if err != nil {
			return []error{fmt.Errorf("failed to fetch DeviceInfo %w", err)}
		}
		ptpDev.devInfo = &fetchedDevInfo
		devInfo = &fetchedDevInfo
	default:
		devInfo = ptpDev.devInfo
	}

	err := ptpDev.callback.Call(devInfo, DeviceInfo)
	if err != nil {
		return []error{fmt.Errorf("callback failed %w", err)}
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (ptpDev *DevInfoCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
		atomic.AddUint32(&ptpDev.count, 1)
	}()
	errorsToReturn := ptpDev.poll()
	resultsChan <- PollResult{
		CollectorName: DPLLCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (ptpDev *DevInfoCollector) CleanUp(key string) error {
	switch key {
	case All, DeviceInfo:
		ptpDev.running = false
		ptpDev.quit <- os.Kill
	default:
		return fmt.Errorf("key %s is not a colletable of %T", key, ptpDev)
	}
	return nil
}

func (ptpDev *DevInfoCollector) GetPollCount() int {
	return int(atomic.LoadUint32(&ptpDev.count))
}

// Returns a new DevInfoCollector from the CollectionConstuctor Factory
func (constructor *CollectionConstructor) NewDevInfoCollector(erroredPolls chan PollResult) (*DevInfoCollector, error) {
	ctx, err := clients.NewContainerContext(constructor.Clientset, PTPNamespace, PodNamePrefix, PTPContainer)
	if err != nil {
		return &DevInfoCollector{}, fmt.Errorf("could not create container context %w", err)
	}

	// Build DPPInfoFetcher ahead of time call to GetPTPDeviceInfo will build the other
	err = devices.BuildPTPDeviceInfo(constructor.PTPInterface)
	if err != nil {
		return &DevInfoCollector{}, fmt.Errorf("failed to build fetcher for PTPDeviceInfo %w", err)
	}

	ptpDevInfo, err := devices.GetPTPDeviceInfo(constructor.PTPInterface, ctx)
	if err != nil {
		return &DevInfoCollector{}, fmt.Errorf("failed to fetch initial DeviceInfo %w", err)
	}

	err = constructor.Callback.Call(&ptpDevInfo, DeviceInfo)
	if err != nil {
		return &DevInfoCollector{}, fmt.Errorf("callback failed %w", err)
	}

	if ptpDevInfo.VendorID != VendorIntel || ptpDevInfo.DeviceID != DeviceE810 {
		return &DevInfoCollector{}, fmt.Errorf("NIC device is not based on E810")
	}

	collector := DevInfoCollector{
		interfaceName: constructor.PTPInterface,
		ctx:           ctx,
		running:       false,
		callback:      constructor.Callback,
		devInfo:       &ptpDevInfo,
		quit:          make(chan os.Signal),
		erroredPolls:  erroredPolls,
		requiresFetch: make(chan bool),
	}

	return &collector, nil
}
