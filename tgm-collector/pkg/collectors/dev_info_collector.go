// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	collectorsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
	validationsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/validations"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/validations"
)

type DevInfoCollector struct {
	*collectorsBase.ExecCollector
	devInfo       *devices.PTPDeviceInfo
	quit          chan os.Signal
	erroredPolls  chan collectorsBase.PollResult
	requiresFetch chan bool
	interfaceName string
	wg            sync.WaitGroup
}

const (
	DevInfoCollectorName = "DevInfo"
	DeviceInfo           = "device-info"
)

// Start sets up the collector so it is ready to be polled
func (ptpDev *DevInfoCollector) Start() error {
	err := ptpDev.ExecCollector.Start()
	if err != nil {
		return fmt.Errorf("failed to start dev info collector: %w", err)
	}
	go ptpDev.monitorErroredPolls()
	return nil
}

// monitorErrored Polls will process errors placed on
// the erredPolls and if required will populate requiresFetch
//
//	if it receives an error:
//		if requiresFetch is empty a value will be inserted.
//		if requiresFetch is not empty it will process errors in an attempt to clear any backlog.
//			We do not want a backlog because if erroredPolls becomes full will block the main
//			loop in runner.Run
func (ptpDev *DevInfoCollector) monitorErroredPolls() {
	ptpDev.wg.Add(1)
	defer ptpDev.wg.Done()
	for {
		select {
		case <-ptpDev.quit:
			return
		case <-ptpDev.erroredPolls:
			// Without this there is a potental deadlock as blocking erroredPolls
			// would eventually cause pollResults to block with would stop the collectors
			// and the consumer of requiresFetch  is a collector.
			if len(ptpDev.requiresFetch) == 0 {
				ptpDev.requiresFetch <- true
			}
		default:
			time.Sleep(time.Microsecond)
		}
	}
}

// polls for the device info, stores it then passes it to the callback
func (ptpDev *DevInfoCollector) poll() error {
	var devInfo *devices.PTPDeviceInfo
	select {
	case <-ptpDev.requiresFetch:
		fetchedDevInfo, err := devices.GetPTPDeviceInfo(ptpDev.interfaceName, ptpDev.GetContext())
		if err != nil {
			return fmt.Errorf("failed to fetch %s %w", DeviceInfo, err)
		}
		ptpDev.devInfo = &fetchedDevInfo
		devInfo = &fetchedDevInfo
	default:
		devInfo = ptpDev.devInfo
	}

	err := ptpDev.Callback.Call(devInfo, DeviceInfo)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (ptpDev *DevInfoCollector) Poll(resultsChan chan collectorsBase.PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()
	errorsToReturn := make([]error, 0)
	err := ptpDev.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- collectorsBase.PollResult{
		CollectorName: DPLLCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (ptpDev *DevInfoCollector) CleanUp() error {
	err := ptpDev.ExecCollector.CleanUp()
	if err != nil {
		return fmt.Errorf("failed to cleanly stop ptp dev info collector: %w", err)
	}
	ptpDev.quit <- os.Kill
	ptpDev.wg.Wait()
	return nil
}

func verify(ptpDevInfo *devices.PTPDeviceInfo, constructor *collectorsBase.CollectionConstructor) error {
	checkErrors := make([]error, 0)
	checks := []validationsBase.Validation{
		validations.NewDeviceDetails(ptpDevInfo),
		validations.NewDeviceDriver(ptpDevInfo),
		validations.NewDeviceFirmware(ptpDevInfo),
	}

	for _, check := range checks {
		err := check.Verify()
		if err != nil {
			var invalidEnv *utils.InvalidEnvError
			if errors.As(err, &invalidEnv) {
				checkErrors = append(checkErrors, err)
			} else {
				log.Warningf("failed to verify %s: %s", check.GetDescription(), err.Error())
			}
		}
	}

	if len(checkErrors) > 0 {
		callbackErr := constructor.Callback.Call(ptpDevInfo, DeviceInfo)
		if callbackErr != nil {
			checkErrors = append(checkErrors, fmt.Errorf("callback failed %w", callbackErr))
		}
		//nolint:wrapcheck // this returns a wrapped error
		return utils.MakeCompositeInvalidEnvError(checkErrors)
	}

	return nil
}

// Returns a new DevInfoCollector from the CollectionConstuctor Factory
func NewDevInfoCollector(constructor *collectorsBase.CollectionConstructor) (collectorsBase.Collector, error) {
	// Build DPPInfoFetcher ahead of time call to GetPTPDeviceInfo will build the other
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &DevInfoCollector{}, fmt.Errorf("failed to create DevInfoCollector: %w", err)
	}

	ptpInterface, err := getPTPInterfaceName(constructor)
	if err != nil {
		return &DevInfoCollector{}, err
	}

	err = devices.BuildPTPDeviceInfo(ptpInterface)
	if err != nil {
		return &DevInfoCollector{}, fmt.Errorf("failed to build fetcher for PTPDeviceInfo %w", err)
	}

	ptpDevInfo, err := devices.GetPTPDeviceInfo(ptpInterface, ctx)
	if err != nil {
		return &DevInfoCollector{}, fmt.Errorf("failed to fetch initial DeviceInfo %w", err)
	}

	err = verify(&ptpDevInfo, constructor)
	if err != nil {
		return &DevInfoCollector{}, err
	}

	collector := DevInfoCollector{
		ExecCollector: collectorsBase.NewExecCollector(
			constructor.DevInfoAnnouceInterval,
			true,
			constructor.Callback,
			ctx,
		),
		interfaceName: ptpInterface,
		devInfo:       &ptpDevInfo,
		quit:          make(chan os.Signal),
		erroredPolls:  constructor.ErroredPolls,
		requiresFetch: make(chan bool, 1),
	}

	return &collector, nil
}

func init() {
	collectorsBase.RegisterCollector(DevInfoCollectorName, NewDevInfoCollector, collectorsBase.Required)
}
