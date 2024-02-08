// SPDX-License-Identifier: GPL-2.0-or-later

package clients_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/clients"
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
		It("should return an error", func() {
			var clientset *clients.Clientset
			clientset, err := clients.GetClientset(notAKubeconfigPath)
			Expect(err).To(HaveOccurred())
			Expect(clientset).To(BeNil())

			clientset, err = clients.GetClientset(emptyKubeconfigList...)
			Expect(err).To(HaveOccurred())
			Expect(clientset).To(BeNil())
		})
	})
	When("A clientset is requested using a valid kubeconfig", func() {
		It("should be returned", func() {
			clientset, err := clients.GetClientset(kubeconfigPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(clientset).NotTo(BeNil())
		})
	})
})

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clients Suite")
}
