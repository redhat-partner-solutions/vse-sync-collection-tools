// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ptp

import (
	"math"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck // ginkgo and gomega are dot imports by convention.
	. "github.com/onsi/gomega"    //nolint:stylecheck // ginkgo and gomega are dot imports by convention.

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
)

const (
	VendorIntel = "0x8086"
	DeviceE810  = "0x1593"
)

var _ = Describe("TGM", func() {
	When("specs for T-GM G.8275.1 profile are finally run", func() {
		It("should have access to custom configuration", func() {
			log.Warningf("Warning %v", ptpConfig)
			Expect(ptpConfig.Namespace).To(Equal("openshift-ptp"))
		})
		It("can extract a positive tty timeout from custom configuration", func() {
			log.Warningf("Warning %v", ptpConfig)
			Expect(ptpConfig.TtyTimeout).To(BeNumerically(">", 0))
		})
		It("can extract a positive number of dpll reads from custom configuration", func() {
			log.Warningf("Warning %v", ptpConfig)
			Expect(ptpConfig.DpllReads).To(BeNumerically(">", 0))
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

			gnssParts := strings.Split(gnss, ",")
			Expect(len(gnssParts)).To(BeNumerically(">", 7), "Failed to parse GNSS string: %s", gnss)

			// TODO use ublox CLI to parse.
			// (http://aprs.gids.nl/nmea/#rmc) These two are bad: "$GNRMC,,V,,,,,,,,,,N,V*37", "$GNGGA,,,,,,0,00,99.99,,,,,,*56"
			By("validating TTY GNSS GNRMC GPS/Transit data and GNGGA Positioning System Fix Data")
			if strings.Contains(gnssParts[0], "GNRMC") {
				Expect(gnssParts[2]).To(Not(Equal("V")))
			} else if strings.Contains(gnssParts[0], "GNGGA") {
				Expect(gnssParts[6]).To(Not(Equal("0")))
			}
		})
		It("DPLL should receive a valid and stable 1PPS signal coming from GNSS", MustPassRepeatedly(ptpConfig.DpllReads), func() {

			ptpContext := clients.NewContainerContext(ptpConfig.Namespace, ptpConfig.PodName, ptpConfig.Container)

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
			dpllOffset /= 100
			Expect(math.Abs(dpllOffset)).To(BeNumerically("<=", 30), "1PPS Phase OUT of Sync")
		})
	})
})
