package tests

import (
	"plugin"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

// Entrypoint for test suite showing how specs can be loaded from plugins at runtime.
func TestRun(t *testing.T) {

	// load test plugins. This can later be done from configuration file.
	test_plugins := []string{"ptp.so", "rvr.so", "quq.so"}

	for _, test_plugin := range test_plugins {

		log.Infof("Attempting to load plugin: %s", test_plugin)

		// opening the plugin registers the specs with Ginkgo
		// providing the specs are registered at the top level (var _ = Describe(...))
		plug, err := plugin.Open(test_plugin)
		if err != nil {
			log.Fatalf("Fail opening plugin!")
		}

		// A function in a plugin can be called if needed, for example to pass in
		// configuration.
		suite, err := plug.Lookup("Configure")
		if err == nil {
			suite.(func(string, int))("ThisIsTheTestKubeconfigValue", 42)
		} else {
			log.Infof("No 'Configure' method for plugin '%s', skipping.", test_plugin)
		}
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests from main app")
}
