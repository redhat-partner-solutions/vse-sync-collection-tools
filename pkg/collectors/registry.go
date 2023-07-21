// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"
)

type collectonBuilderFunc func(*CollectionConstructor) (Collector, error)

type CollectorRegistry struct {
	registry map[string]collectonBuilderFunc
	required []string
	optional []string
}

var registry *CollectorRegistry

func GetRegistry() *CollectorRegistry {
	return registry
}

func (reg *CollectorRegistry) register(
	collectorName string,
	builderFunc collectonBuilderFunc,
	reqiuired bool,
) {
	reg.registry[collectorName] = builderFunc
	if reqiuired {
		reg.required = append(reg.required, collectorName)
	} else {
		reg.optional = append(reg.optional, collectorName)
	}
}

func (reg *CollectorRegistry) GetBuilderFunc(collectorName string) (collectonBuilderFunc, error) {
	builderFunc, ok := reg.registry[collectorName]
	if !ok {
		return nil, fmt.Errorf("not index in registry for collector named %s", collectorName)
	}
	return builderFunc, nil
}

func (reg *CollectorRegistry) GetRequiredNames() []string {
	return reg.required
}

func (reg *CollectorRegistry) GetOptionalNames() []string {
	return reg.optional
}

func RegisterCollector(collectorName string, builderFunc collectonBuilderFunc, required bool) {
	if registry == nil {
		registry = &CollectorRegistry{
			registry: make(map[string]collectonBuilderFunc, 0),
			required: make([]string, 0),
			optional: make([]string, 0),
		}
	}
	registry.register(collectorName, builderFunc, required)
}
