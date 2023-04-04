// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// configSectionRegistry maps a `sectionKey` to an instance pointer that references
// the target for that config section.
var configSectionRegistry = make(map[string]*interface{})

// customConfig exists to implement a custom UnmarshalYAML.
// The custom UnmarshalYAML allows dynamically selecting the struct instance into which that config section is unmarshalled.
// Register an instance to be used by calling `config.RegisterCustomConfigSection`.
type customConfig struct{}

// This custom UnmarshalYAML will find the instance that has been registered
// against the YAML key and populate it with the data from the corresponding
// value.
func (cc *customConfig) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("CustomConfig must contain YAML Map, has %v", value.Kind)
	}
	for item := 0; item < len(value.Content); item += 2 {
		var customConfigKey string
		if err := value.Content[item].Decode(&customConfigKey); err != nil {
			return fmt.Errorf("could not decode custom config key: %w", err)
		}
		log.Infof("Found individual test configuration section: %s", customConfigKey)
		log.Debugf("Raw value of %s is: %v", customConfigKey, value.Content[item+1])

		// Get the object into which to decode from the section registry
		testConfigStruct, err := getRegisteredInstancePtr(customConfigKey)
		if err != nil {
			log.Warnf("Unknown custom config section found, skipping: %s", customConfigKey)
			continue
		}
		log.Debugf("Retrieved [%T]%+v", testConfigStruct, testConfigStruct)

		if err := value.Content[item+1].Decode(testConfigStruct); err != nil {
			return fmt.Errorf("could not load custom config section: %w", err)
		}
	}
	return nil
}

// RegisterCustomConfig will register the `target` instance pointer to later be
// populated with data from a section of the configuration file with key
// `sectionKey`.
// The key MUST be found in the configuration file in a position that maps to a
// `CustomConfig` struct in the top-level `ConfigFile` struct.
func RegisterCustomConfig(sectionKey string, target interface{}) error {
	log.Infof("Registering section '%s'", sectionKey)
	if _, exists := configSectionRegistry[sectionKey]; exists {
		return fmt.Errorf("cannot register %s, already registered", sectionKey)
	}
	configSectionRegistry[sectionKey] = &target

	log.Debugf("Registered section '%s' for [%T]%v", sectionKey, target, target)
	log.Debugf("Current state of registry %v", configSectionRegistry)
	return nil
}

// getRegisteredInstancePtr returns the pointer to the `target` struct that was
// registered for the `sectionKey`
func getRegisteredInstancePtr(sectionKey string) (interface{}, error) {
	log.Debugf("Retrieving section '%s'", sectionKey)
	if _, exists := configSectionRegistry[sectionKey]; !exists {
		return nil, fmt.Errorf("cannot find: %s", sectionKey)
	}
	log.Debugf(
		"Retrieved section '%s', got [%T]%v",
		sectionKey,
		configSectionRegistry[sectionKey],
		configSectionRegistry[sectionKey],
	)
	return *configSectionRegistry[sectionKey], nil
}
