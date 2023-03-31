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

package collectors_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocpv1 "github.com/openshift/api/config/v1"
	ocpfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors"
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
			_, err := ocp.ConfigV1().ClusterOperators().Create(context.TODO(), &ocpv1.ClusterOperator{
				TypeMeta:   metav1.TypeMeta{Kind: "ClusterOperator", APIVersion: "config.openshift.io/v1"},
				ObjectMeta: metav1.ObjectMeta{Name: "openshift-apiserver"},
				Spec:       ocpv1.ClusterOperatorSpec{},
				Status: ocpv1.ClusterOperatorStatus{Versions: []ocpv1.OperandVersion{
					{Name: "notThisVersion", Version: "notThisVersion"},
					{Name: collectors.ApiServerClusterOperator, Version: expectedVersion},
				}},
			}, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return the cluster version", func() {
			clusterVersion, err := collectors.GetClusterVersion(ocp.ConfigV1())

			Expect(err).ToNot(HaveOccurred())
			Expect(clusterVersion).ToNot(Equal(collectors.UnknownClusterVersion))
			Expect(clusterVersion).To(Equal(expectedVersion))
		})
	})
	When("No cluster version exists", func() {
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
