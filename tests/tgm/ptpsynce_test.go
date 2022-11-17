package ptpsynce_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	_ "github.com/redhat-partner-solutions/vse-sync-testsuite/tests/ptpsynce/featurecalnex"
)

func TestPtpsynce(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ptpsynce Suite")
}

var _ = Describe("run Ptpsynce", func() {
	message := "LOCKED"
	It("Test Spec 1: Is the card detected?", func() {
		//put the code of the test spec 1
	})
	It("Test Spec 2: Do we have a GPS reference clock?", func() {
		//code for that test spec 2
    })
	It("Test Spec 3: Is ptp grandmaster clock locked?", func() {
		Expect(len(message)).To(Equal(6)) //testing an assertion provided by gomega
		//code for that
	})
})
