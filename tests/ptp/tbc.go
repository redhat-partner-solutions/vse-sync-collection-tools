package ptp

import (

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/images"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/pods"
	log "github.com/sirupsen/logrus"
)

const (
	NgenPodName = "ngen-g82732-tbc"
)

var (
	NgenArgs   = []string{
		"--help",			// Constant set of arguments
	}
)
var _ = Describe("TBC", func() {
	When("specs for T-BC G.8273.2 profile are finally run", func() {
		It("should have access to custom configuration", func() {
			log.Warningf("Warning %v", ptpConfig)
			Expect(ptpConfig.Namespace).To(Equal("openshift-ptp"))
			Expect(ptpConfig.BoundaryClock.InterfaceName).To(Equal("ens2f0"))
		})
	})
	When("There is a Telecom Boundary Clock under test", func() {
		It("should pass G.8273.2 Noise Generation Requirements", func() {
			//define pod to calculate KPIs
			NgenImageName, err := images.GetNgenBaseImage()
			if err != nil {
				log.Panic("unexpected error: %s", err)
			}
			ngenPod := pods.NewBuilder(NgenPodName, ptpConfig.Namespace, NgenImageName, NgenArgs)	
			
			ngenPod, err = ngenPod.Create()
			if ngenPod.Definition == nil {
				log.Infof("uops definition not created")
			}
			if err != nil {
				log.Panic("unexpected error: %s", err)	
			}
		})
	})	
})