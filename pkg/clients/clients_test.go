// SPDX-License-Identifier: GPL-2.0-or-later

package clients_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

const (
	kubeconfigPath     string = "test_files/kubeconfig"
	notAKubeconfigPath string = ""
)

var (
	emptyKubeconfigList []string
)

var _ = Describe("Client", func() {
	BeforeEach(func() {
		clients.ClearClientSet()
	})

	When("A clientset is requested with no kubeconfig", func() {
		It("should panic", func() {
			var clientset *clients.Clientset
			Expect(func() { clientset = clients.GetClientset(notAKubeconfigPath) }).To(Panic())
			Expect(clientset).To(BeNil())
			Expect(func() { clientset = clients.GetClientset(emptyKubeconfigList...) }).To(Panic())
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

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clients Suite")
}
