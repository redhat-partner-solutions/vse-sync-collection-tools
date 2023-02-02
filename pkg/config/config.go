package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type configFile struct {
	Value string `yaml:"value"` // Framework config here e.g. reports, global timeouts, run descriptor, etc.
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
		return err
	}

	return yaml.Unmarshal(contents, &Config)
}
