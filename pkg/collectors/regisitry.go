// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"
)

type collectonBuilderFunc func(*CollectionConstructor) (Collector, error)

type CollectorRegistry struct {
	registry map[string]collectonBuilderFunc
}

var registry *CollectorRegistry

func GetRegistry() *CollectorRegistry {
	return registry
}

func (reg *CollectorRegistry) register(
	collectorName string,
	builderFunc collectonBuilderFunc,
) {
	reg.registry[collectorName] = builderFunc
}

func (reg *CollectorRegistry) GetBuilderFunc(collectorName string) (collectonBuilderFunc, error) {
	builderFunc, ok := reg.registry[collectorName]
	if !ok {
		return nil, fmt.Errorf("not index in registry for collector named %s", collectorName)
	}
	return builderFunc, nil
}

func init() {
	registry = &CollectorRegistry{
		registry: make(map[string]collectonBuilderFunc, 0),
	}
}
