// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type Collector interface {
	Start(key string) error                      // Setups any internal state required for collection to happen
	Poll(chan PollResult, *utils.WaitGroupCount) // Poll for collectables
	CleanUp(key string) error                    // Stops the collector and cleans up any internal state. It should result in a state that can be started again
	GetPollCount() int                           // Returns the number of completed poll
	GetPollRate() float64                        // Returns the collectors polling rate
}

// A union of all values required to be passed into allconstructions
type CollectionConstructor struct {
	Callback     callbacks.Callback
	Clientset    *clients.Clientset
	PTPInterface string
	Msg          string
	PollRate     float64
}

type PollResult struct {
	CollectorName string
	Errors        []error
}
