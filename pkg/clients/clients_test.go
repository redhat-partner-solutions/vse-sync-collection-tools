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

package clients_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

const (
	kubeconfigPath string = "test_files/kubeconfig"
	notAKubeconfigPath string = ""
)

var (
	emptyKubeconfigList []string
)

var _ = Describe("Client", func() {
	When("A clientset is requested with no kubeconfig", func() {
		It("should panic", func() {
			var clientset *clients.Clientset
			Expect(func() {clientset = clients.GetClientset(notAKubeconfigPath)}).To(Panic())
			Expect(clientset).To(BeNil())
			Expect(func() {clientset = clients.GetClientset(emptyKubeconfigList...)}).To(Panic())
			Expect(clientset).To(BeNil())
		})
	})
	When("A clientset is requested using a valid kubeconfig", func() {
		It("should be returned", func() {
			clientset := clients.GetClientset(kubeconfigPath)
			Expect(clientset).NotTo(BeNil())
		})
	})
})

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clients Suite")
}
