// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type Collector interface {
	Start(key string) error   // Setups any internal state required for collection to happen
	ShouldPoll() bool         // Check if poll time has alapsed and if it should be polled again
	Poll() []error            // Poll for collectables
	CleanUp(key string) error // Cleans up any internal state
}

// A union of all values required to be passed into all constuctions
type CollectionConstuctor struct {
	Callback     callbacks.Callback
	Clientset    *clients.Clientset
	PTPInterface string
	Msg          string
	PollRate     float64
}
