// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type Collector interface {
	Start() error                                // Setups any internal state required for collection to happen
	Poll(chan PollResult, *utils.WaitGroupCount) // Poll for collectables
	CleanUp() error                              // Stops the collector and cleans up any internal state. It should result in a state that can be started again
	GetPollInterval() int                        // Returns the collectors polling interval
	IsAnnouncer() bool
}

// A union of all values required to be passed into all constructions
type CollectionConstructor struct {
	Callback               callbacks.Callback
	Clientset              *clients.Clientset
	ErroredPolls           chan PollResult
	PTPInterface           string
	Msg                    string
	LogsOutputFile         string
	TempDir                string
	PollInterval           int
	DevInfoAnnouceInterval int
	IncludeLogTimestamps   bool
	KeepDebugFiles         bool
}

type PollResult struct {
	CollectorName string
	Errors        []error
}

type baseCollector struct {
	callback     callbacks.Callback
	isAnnouncer  bool
	running      bool
	pollInterval int
}

func (base *baseCollector) GetPollInterval() int {
	return base.pollInterval
}

func (base *baseCollector) IsAnnouncer() bool {
	return base.isAnnouncer
}

func (base *baseCollector) Start() error {
	base.running = true
	return nil
}

func (base *baseCollector) CleanUp() error {
	base.running = false
	return nil
}

func newBaseCollector(
	pollInterval int,
	isAnnouncer bool,
	callback callbacks.Callback,
) *baseCollector {
	return &baseCollector{
		callback:     callback,
		isAnnouncer:  isAnnouncer,
		running:      false,
		pollInterval: pollInterval,
	}
}
