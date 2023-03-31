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
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type configFile struct {
	Value       string         `yaml:"value"` // Framework config here e.g. reports, global timeouts, run descriptor, etc.
	TestConfigs []customConfig `yaml:"suite_configs"`
}

var Config = configFile{}

// LoadConfigFromFile is the entrypoint for the config package; it will load a configuration
// file, populate the top-level config, and any custom config sections that have
// been registered prior to the config being loaded.
func LoadConfigFromFile(filePath string) error {
	log.Infof("Loading config from file: %s", filePath)

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not open config file: %w", err)
	}

	return yaml.Unmarshal(contents, &Config) //nolint:wrapcheck // Allowing error propagation in this instance
}
