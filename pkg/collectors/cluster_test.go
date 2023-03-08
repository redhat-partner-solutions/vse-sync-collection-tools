package collectors_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors"

	ocpv1 "github.com/openshift/api/config/v1"
	ocpfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	expectedVersion = "v0.0.0"
)

var _ = Describe("Cluster", func() {
	var ocp *ocpfake.Clientset

	BeforeEach(func() {
		ocp = ocpfake.NewSimpleClientset()
	})

	When("A cluster version exists", func() {
		BeforeEach(func() {
			ocp.ConfigV1().ClusterOperators().Create(context.TODO(), &ocpv1.ClusterOperator{
				metav1.TypeMeta{Kind: "ClusterOperator", APIVersion: "config.openshift.io/v1"},
				metav1.ObjectMeta{Name: "openshift-apiserver"},
				ocpv1.ClusterOperatorSpec{},
				ocpv1.ClusterOperatorStatus{Versions: []ocpv1.OperandVersion{
					{Name: "notThisVersion",Version: "notThisVersion"},
					{Name: collectors.ApiServerClusterOperator,Version: expectedVersion},
				}},
			}, metav1.CreateOptions{})
		})

		It("should return the cluster version", func() {
			clusterVersion, err := collectors.GetClusterVersion(ocp.ConfigV1())

			Expect(err).ToNot(HaveOccurred())
			Expect(clusterVersion).ToNot(Equal(collectors.UnknownClusterVersion))
			Expect(clusterVersion).To(Equal(expectedVersion))
		})
	})
	When("No cluster version exists", func(){
		It("should raise an error", func() {
			clusterVersion, err := collectors.GetClusterVersion(ocp.ConfigV1())

			Expect(err).To(HaveOccurred())
			Expect(clusterVersion).To(Equal(collectors.UnknownClusterVersion))
		})
	})
})

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clients Suite")
}
