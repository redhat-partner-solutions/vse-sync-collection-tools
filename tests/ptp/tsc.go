package ptp

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TSC", func() {
	When("The package is imported", func() {
		It("Should cause the test to be run", func() {
			Expect(true).To(BeTrue())
		})
	})
})
