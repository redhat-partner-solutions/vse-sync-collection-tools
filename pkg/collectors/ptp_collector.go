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

type PTPCollector struct {
	lastPoll        time.Time
	callback        callbacks.Callback
	data            map[string]interface{}
	running         map[string]bool
	DataTypes       [3]string
	ctx             clients.ContainerContext
	interfaceName   string
	inversePollRate float64
}

const (
	VendorIntel = "0x8086"
	DeviceE810  = "0x1593"

	DeviceInfo = "device-info"
	DPLLInfo   = "dpll-info"
	GNSSDev    = "gnss-dev"
	All        = "all"

	PTPNamespace  = "openshift-ptp"
	PodNamePrefix = "linuxptp-daemon-"
	PTPContainer  = "linuxptp-daemon-container"
)

var collectables = [3]string{
	DeviceInfo,
	DPLLInfo,
	GNSSDev,
}

func (ptpDev *PTPCollector) getNotCollectableError(key string) error {
	return fmt.Errorf("key %s is not a colletable of %T", key, ptpDev)
}

func (ptpDev *PTPCollector) getErrorIfNotCollectable(key string) error {
	if _, ok := ptpDev.data[key]; !ok {
		return ptpDev.getNotCollectableError(key)
	} else {
		return nil
	}
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

// ShouldPoll checks if enough time has passed since the last poll
func (ptpDev *PTPCollector) ShouldPoll() bool {
	log.Debugf("since: %v", time.Since(ptpDev.lastPoll).Seconds())
	log.Debugf("wait: %v", ptpDev.inversePollRate)
	return time.Since(ptpDev.lastPoll).Seconds() >= ptpDev.inversePollRate
}

// fetchLine will call the requested key's function
// store the result of that function into the collectors data
// and returns a json encoded version of that data
func (ptpDev *PTPCollector) fetchLine(key string) (line []byte, err error) {
	switch key {
	case DeviceInfo:
		ptpDevInfo := devices.GetPTPDeviceInfo(ptpDev.interfaceName, ptpDev.ctx)
		ptpDev.data[DeviceInfo] = ptpDevInfo
		line, err = json.Marshal(ptpDevInfo)
	case DPLLInfo:
		dpllInfo := devices.GetDevDPLLInfo(ptpDev.ctx, ptpDev.interfaceName)
		ptpDev.data[DPLLInfo] = dpllInfo
		line, err = json.Marshal(dpllInfo)
	case GNSSDev:
		// TODO make lines and timeout configs
		devInfo, ok := ptpDev.data[DeviceInfo].(devices.PTPDeviceInfo)
		if !ok {
			return nil, fmt.Errorf("DeviceInfo was not able to be unpacked")
		}
		gnssDevLine := devices.ReadGNSSDev(ptpDev.ctx, devInfo, 1, 1)

		ptpDev.data[GNSSDev] = gnssDevLine
		line, err = json.Marshal(gnssDevLine)
	default:
		return nil, ptpDev.getNotCollectableError(key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to marshall line(%v) in PTP collector: %w", key, err)
	}
	return line, nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (ptpDev *PTPCollector) Poll() []error {
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
	ptpDev.lastPoll = time.Now()
	return errorsToReturn
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

	data := make(map[string]interface{})
	running := make(map[string]bool)

	data[DeviceInfo] = devices.GetPTPDeviceInfo(constuctor.PTPInterface, ctx)
	data[DPLLInfo] = devices.GetDevDPLLInfo(ctx, constuctor.PTPInterface)

	ptpDevInfo, ok := data[DeviceInfo].(devices.PTPDeviceInfo)
	if !ok {
		return &PTPCollector{}, fmt.Errorf("DeviceInfo was not able to be unpacked")
	}
	if ptpDevInfo.VendorID != VendorIntel || ptpDevInfo.DeviceID != DeviceE810 {
		return &PTPCollector{}, fmt.Errorf("NIC device is not based on E810")
	}

	inversePollRate := 1.0 / constuctor.PollRate
	offset := time.Duration(float64(time.Second.Nanoseconds()) * inversePollRate)

	collector := PTPCollector{
		interfaceName:   constuctor.PTPInterface,
		ctx:             ctx,
		DataTypes:       collectables,
		data:            data,
		running:         running,
		callback:        constuctor.Callback,
		inversePollRate: inversePollRate,
		lastPoll:        time.Now().Add(-offset), // Subtract off a polling time so the first poll hits
	}

	return &collector, nil
}
