package ptp

import (
	"strings"
	"strconv"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/retrievers/devices"
	log "github.com/sirupsen/logrus"
)

const (
	VendorIntel = "0x8086"
	DeviceE810 = "0x1593"
	DPLLReads = 2
)

var _ = Describe("TGM", func() {
	When("specs for T-GM G.8275.1 profile are finally run", func() {
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
			if ptpDevInfo.VendorID != VendorIntel || ptpDevInfo.DeviceID != DeviceE810 {
				Skip("NIC device is not based on E810")
			}

			gnss := devices.ReadTtyGNSS(ptpContext, ptpDevInfo, 1, ptpConfig.TtyTimeout)

			s := strings.Split(gnss, ",")
			Expect(len(s)).To(BeNumerically(">", 7), "Failed to parse GNSS string: %s", gnss)

			// TODO use ublox CLI to parse.
			// (http://aprs.gids.nl/nmea/#rmc) These two are bad: "$GNRMC,,V,,,,,,,,,,N,V*37", "$GNGGA,,,,,,0,00,99.99,,,,,,*56"
			By("validating TTY GNSS GNRMC GPS/Transit data and GNGGA Positioning System Fix Data")
			if strings.Contains(s[0], "GNRMC") {
				Expect(s[2]).To(Not(Equal("V")))
			} else if strings.Contains(s[0], "GNGGA") {
				Expect(s[6]).To(Not(Equal("0")))
			}
		})
		It("DPLL should receive a valid and stable 1PPS signal coming from GNSS",  MustPassRepeatedly(DPLLReads), func() {

			ptpContext := clients.NewContainerContext(ptpConfig.Namespace, ptpConfig.PodName, ptpConfig.Container)
			
			// retrieve DPLL info
			dpll := devices.GetDevDPLLInfo(ptpContext, ptpConfig.InterfaceName)

			// TODO reveal actual DPLL status and provide recommendations 
			// 0: Invalid. Check your card, firmware, and drivers.
  			// 1: Freerun: The 1PPS DPLL is not defined by any incoming reference PPS signal. If this stays for a long time. Check your incoming 1PPS signal!
  			// 2: Locked to 1PPS but holdover not acquired yet
  			// 3: Normal operation. Locked 1PPS and holdover acquired. Things are good!
			By("validating PPS DPLL is in normal Operation")
			Expect(dpll.State).To(Equal("3"), "PPS DPLL state NOT in normal operation")			
			
			// The DPLL Offset value should be bounded by abs (-30,+30)ns.
			// The value is in the order of 10s of picoseconds, so it needs to be divided by 100 to get ns.
			By("validating DPLL phase offset is in sync")
			dpllOffset, err := strconv.ParseFloat(dpll.Offset, 64)
			Expect(err).NotTo(HaveOccurred())
			dpllOffset = dpllOffset / 100
			Expect(math.Abs(dpllOffset)).To(BeNumerically("<=", 30), "1PPS Phase OUT of Sync")	
		})
	})
})
