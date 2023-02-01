// run_tests provides the entrypoint into the tests
package run_tests

import (
	"testing"

	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/redhat-partner-solutions/vse-sync-testsuite/tests/ptp"
)

// TestRun is the entrypoint for the framework and will run all tests that have
// been registered with ginkgo.
func TestRun(t *testing.T) {
	log.Info("Starting Tests")
	
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")

	log.Info("Finished Tests")
}
