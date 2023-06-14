// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
)

type PTPCollector struct {
	callback      callbacks.Callback
	running       map[string]bool
	DataTypes     [2]string
	ctx           clients.ContainerContext
	interfaceName string
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
func (ptpDev *PTPCollector) fetchLine(key string) (line []byte, err error) { //nolint:funlen // allow slightly long function
	switch key {
	case DeviceInfo:
		ptpDevInfo, fetchError := devices.GetPTPDeviceInfo(ptpDev.interfaceName, ptpDev.ctx)
		if fetchError != nil {
			return []byte{}, fmt.Errorf("failed to fetch ptpDevInfo %w", fetchError)
		}
		line, err = json.Marshal(ptpDevInfo)
	case DPLLInfo:
		dpllInfo, fetchError := devices.GetDevDPLLInfo(ptpDev.ctx, ptpDev.interfaceName)
		if fetchError != nil {
			return []byte{}, fmt.Errorf("failed to fetch dpllInfo %w", fetchError)
		}
		line, err = json.Marshal(dpllInfo)
	default:
		return []byte{}, ptpDev.getNotCollectableError(key)
	}
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshall line(%v) in PTP collector: %w", key, err)
	}
	return line, nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (ptpDev *PTPCollector) Poll(resultsChan chan PollResult) {
	errorsToReturn := make([]error, 0)

	for key, isRunning := range ptpDev.running {
		if isRunning {
			line, err := ptpDev.fetchLine(key)
			// TODO: handle (better)
			if err != nil {
				errorsToReturn = append(errorsToReturn, err)
			} else {
				err = ptpDev.callback.Call(fmt.Sprintf("%T", ptpDev), key, string(line))
				if err != nil {
					errorsToReturn = append(errorsToReturn, err)
				}
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

// Returns a new PTPCollector from the CollectionConstuctor Factory
// It will set the lastPoll one polling time in the past such that the initial
// request to ShouldPoll should return True
func (constuctor *CollectionConstuctor) NewPTPCollector() (*PTPCollector, error) {
	ctx, err := clients.NewContainerContext(constuctor.Clientset, PTPNamespace, PodNamePrefix, PTPContainer)
	if err != nil {
		return &PTPCollector{}, fmt.Errorf("could not create container context %w", err)
	}

	running := make(map[string]bool)

	ptpDevInfo, err := devices.GetPTPDeviceInfo(constuctor.PTPInterface, ctx)
	if err != nil {
		return &PTPCollector{}, fmt.Errorf("failed to fetch initial DeviceInfo %w", err)
	}
	if ptpDevInfo.VendorID != VendorIntel || ptpDevInfo.DeviceID != DeviceE810 {
		return &PTPCollector{}, fmt.Errorf("NIC device is not based on E810")
	}

	collector := PTPCollector{
		interfaceName: constuctor.PTPInterface,
		ctx:           ctx,
		DataTypes:     ptpCollectables,
		running:       running,
		callback:      constuctor.Callback,
	}

	return &collector, nil
}
