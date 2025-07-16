// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"
	"time"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type Collector interface {
	Start() error                                // Setups any internal state required for collection to happen
	Poll(chan PollResult, *utils.WaitGroupCount) // Poll for collectables
	CleanUp() error                              // Stops the collector and cleans up any internal state. It should result in a state that can be started again
	GetPollInterval() time.Duration              // Returns the collectors polling interval
	IsAnnouncer() bool
}

// A union of all values required to be passed into all constructions
type CollectionConstructor struct {
	Callback               callbacks.Callback
	Clientset              *clients.Clientset
	ErroredPolls           chan PollResult
	PTPInterface           string
	PTPNodeName            string
	LogsOutputFile         string
	TempDir                string
	PollInterval           int
	DevInfoAnnouceInterval int
	IncludeLogTimestamps   bool
	KeepDebugFiles         bool
	UnmanagedDebugPod      bool
	ClockType              string
}

func NewCollectionConstructor(
	kubeConfig string,
	useAnalyserJSON bool,
	outputFile string,
	ptpInterface string,
	ptpNodeName string,
	logsOutputFile string,
	tempDir string,
	pollInterval int,
	devInfoAnnouceInterval int,
	includeLogTimestamps bool,
	keepDebugFiles bool,
	unmanagedDebugPod bool,
	clockType string,
) (*CollectionConstructor, error) {
	clientset, err := clients.GetClientset(kubeConfig)
	if err != nil {
		return &CollectionConstructor{}, fmt.Errorf("failed to create constructor values: %w", err)
	}

	outputFormat := callbacks.Raw
	if useAnalyserJSON {
		outputFormat = callbacks.AnalyserJSON
	}

	callback, err := callbacks.SetupCallback(outputFile, outputFormat)
	if err != nil {
		return &CollectionConstructor{}, fmt.Errorf("failed to create constructor values: %w", err)
	}

	return &CollectionConstructor{
		Callback:               callback,
		Clientset:              clientset,
		PTPInterface:           ptpInterface,
		PTPNodeName:            ptpNodeName,
		LogsOutputFile:         logsOutputFile,
		TempDir:                tempDir,
		PollInterval:           pollInterval,
		DevInfoAnnouceInterval: devInfoAnnouceInterval,
		IncludeLogTimestamps:   includeLogTimestamps,
		KeepDebugFiles:         keepDebugFiles,
		UnmanagedDebugPod:      unmanagedDebugPod,
		ClockType:              clockType,
	}, nil
}

type PollResult struct {
	CollectorName string
	Errors        []error
}

type baseCollector struct {
	callback     callbacks.Callback
	poller       func() (callbacks.OutputType, error)
	name         string
	callbackTag  string
	pollInterval time.Duration
	isAnnouncer  bool
	running      bool
}

func (base *baseCollector) GetPollInterval() time.Duration {
	return base.pollInterval
}

func (base *baseCollector) IsAnnouncer() bool {
	return base.isAnnouncer
}

func (base *baseCollector) Start() error {
	base.running = true
	if base.poller == nil {
		utils.IfErrorExitOrPanic(fmt.Errorf("poller not set for collector %s", base.name))
	}
	return nil
}

func (base *baseCollector) CleanUp() error {
	base.running = false
	return nil
}

func (base *baseCollector) poll() error {
	result, err := base.poller()
	if err != nil {
		return fmt.Errorf("failed to fetch  %s %w", base.callbackTag, err)
	}
	err = base.callback.Call(result, gpsNavKey)
	if err != nil {
		return fmt.Errorf("callback failed %w", err)
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (base *baseCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer wg.Done()
	errorsToReturn := make([]error, 0)
	err := base.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- PollResult{
		CollectorName: base.name,
		Errors:        errorsToReturn,
	}
}

func newBaseCollector(
	pollInterval int,
	isAnnouncer bool,
	callback callbacks.Callback,
	name string,
	callbackTag string,
) *baseCollector {
	return &baseCollector{
		name:         name,
		callbackTag:  callbackTag,
		callback:     callback,
		isAnnouncer:  isAnnouncer,
		running:      false,
		pollInterval: time.Duration(pollInterval) * time.Second,
	}
}
