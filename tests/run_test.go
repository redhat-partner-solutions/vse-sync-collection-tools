// run_tests provides the entrypoint into the tests
package run_tests

import (
	"flag"
	"testing"

	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/config"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/env"

	_ "github.com/redhat-partner-solutions/vse-sync-testsuite/tests/ptp"
)

var (
	configPath string
	envPath    string
)

// This is a workaround to set the LogLevel to Debug before flags are parsed or the config is loaded. It is necessary
// because flags can only be parsed once globally which must be after ginkgo registers its flags, and because logging
// needs to happen when config sections are registered by Suite packages.
func setLogLevel() bool {
	log.SetLevel(log.DebugLevel)
	return true
}

var _ = setLogLevel()

// The workaround could be made unnecessary if the framework had control over when the test Suite packages are loaded.
// One possibility for this would be dynamically loading Suite packages as Go plugins but that method has different
// compromises and restrictions. See https://github.com/redhat-partner-solutions/vse-sync-testsuite/tree/plugin-test/tests

func init() {
	flag.StringVar(&configPath, "config", "./config.yaml", "path to the config file")
	flag.StringVar(&envPath, "env", "./env.yaml", "path to the environment file")
}

// TestRun is the entrypoint for the framework trigger execution of all
// discovered Ginkgo specs
func TestRun(t *testing.T) {
	flag.Parse() // Ginkgo will trigger the parse, and a second call is ignored, so this does nothing in most cases.
	log.Info("Starting Tests")

	if err := config.LoadConfigFromFile(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// TODO: Can we modify the ginkgo Spec tree at this point to make it more user friendly to check about skipping tests?

	clusterInfo, err := env.LoadClusterDefFromFile(envPath)
	if err != nil {
		log.Fatal(err)
	}
	clients.GetClientset(clusterInfo.KubeconfigPath)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")

	log.Info("Finished Tests")
}
