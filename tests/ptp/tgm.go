package ptp

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/retrievers/devices"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("TGM", func() {
	When("specs are finally run", func() {
		It("should have access to custom configuration", func() {
			log.Warningf("Warning %v", ptpConfig)
			Expect(ptpConfig.Namespace).To(Equal("openshift-ptp"))
			Expect(ptpConfig.InterfaceName).To(Equal("ens7f0"))
		})
	})
	When("There is a Telco Grand Master under test", func() {
		It("should receive GNSS signals", func() {
			ptpContext := clients.NewContainerContext(ptpConfig.Namespace, ptpConfig.PodName, ptpConfig.Container)
			ptpDevInfo := devices.GetPTPDeviceInfo(ptpConfig.InterfaceName, ptpContext)
			if ptpDevInfo.VendorID != "0x8086" && ptpDevInfo.DeviceID != "0x1593" {
				Skip("NIC device is not based on E810")
			}

			gnss := devices.ReadTtyGNSS(ptpContext, ptpDevInfo)

			// These two are bad: http://aprs.gids.nl/nmea/#rmc
			// $GNRMC,,V,,,,,,,,,,N,V*37
			// $GNGGA,,,,,,0,00,99.99,,,,,,*56
			s := strings.Split(gnss, ",")
			log.Debugf("s: %v", gnss)
			log.Debugf("s: %v", s)
			Expect(len(s)).To(BeNumerically(">", 7), "Failed to parse GNSS string: %s", gnss)

			By("validating TTY GNSS GNRMC GPS/Transit data and GNGGA Positioning System Fix Data")
			if strings.Contains(s[0], "GNRMC") {
				Expect(s[2]).To(Not(Equal("V")))
			} else if strings.Contains(s[0], "GNGGA") {
				Expect(s[6]).To(Not(Equal("0")))
			}

		})
	})
})
