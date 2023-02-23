package env_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/env"
)

var _ = Describe("Cluster", func() {
	When("Loading  a cluster definition from file", func() {
		It("should load without error", func() {
			_, err := env.LoadClusterDefFromFile("test_files/cluster_load.yaml")
			Expect(err).To(BeNil())
		})
		It("should load data from the file", func() {
			clusterDef, _ := env.LoadClusterDefFromFile("test_files/cluster_load.yaml")
			Expect(clusterDef).ToNot(BeNil())
			Expect(clusterDef.KubeconfigPath).To(Equal("~/kubeconfig"))
			Expect(clusterDef.ClusterName).To(Equal("Albatross"))
			Expect(clusterDef.Nodes[0].Hostname).To(Equal("albatross"))
		})
	})
})

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Suite")
}
