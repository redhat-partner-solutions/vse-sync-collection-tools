package ptp

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TGM", func() {
	When("the specs are finally run", func() {
		It("shoud have access to its custom configuration", func() {
			Expect(ptpConfig.MaxOffset).To(Equal(30))
			Expect(ptpConfig.OtherValue).To(Equal("foo"))
		})
	})
})
