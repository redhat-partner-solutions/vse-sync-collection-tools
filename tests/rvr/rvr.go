package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

var (
	kubeconfig string = "default"
	count int = 0
)

var _ = Describe("Tests in a third Plugin", func() {
	When("There is a test in a third Go plugin", func() {
		It("should be run isolated from other plugins", func() {
			// Test logic goes here
			log.Infof("Running first test case %d from a third plugin", count)
			Expect(kubeconfig).NotTo(BeEmpty())
			Expect(count).To(Equal(0))
		})
		It("should be run isolated from other plugins", func() {
			// Test logic goes here
			log.Infof("Running second test case %d from a third plugin", count)
			Expect(kubeconfig).NotTo(BeEmpty())
			Expect(count).To(Equal(0))
		})
	})
})

// A top-level execution for the plugin tests is not supported. It will run all
// tests that have been registered by over plugins. However, if it's not run by
// the framework it can be used to run the plugin in isolation as an executable.
func RunTests(t *testing.T) {
	log.Infof("Running tests in third plugin only")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests from Third plugin.")
}
