package main

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	kubeconfig string = "default"
)

var _ = Describe("Tests in a Plugin", func() {
	When("There is a test in a Go plugin", func() {
		It("should be run isolated from other plugins", func() {
			fmt.Println("Running first test case from a plugin with no BeforeEach")
			Expect(kubeconfig).To(Equal("default"))
			Expect(true).To(BeTrue())
		})
		It("should be run without using BeforeEach in another plugin", func() {
			fmt.Println("Running second test case from a plugin with no BeforeEach")
			Expect(kubeconfig).To(Equal("default"))
			Expect(true).To(BeTrue())
		})
	})
})
