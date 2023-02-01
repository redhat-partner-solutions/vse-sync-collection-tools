package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/config"
)

var _ = Describe("Config", func() {
	When("A config file is loaded", func() {
		It("should load correctly", func() {
			err := config.LoadConfigFromFile("test_files/test_config.yaml")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Suite")
}
