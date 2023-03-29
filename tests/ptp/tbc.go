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
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck // ginkgo and gomega are dot imports by convention.
	. "github.com/onsi/gomega"    //nolint:stylecheck // ginkgo and gomega are dot imports by convention.

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/helperptp"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/images"
)

const (
	captureTIME = 5

	ptpLogsProcessed = "/tmp/data.csv"
	ptpLogs          = "/tmp/data.log"
)

var _ = Describe("TBC", func() {
	When("specs for T-BC G.8273.2 profile are finally run", func() {
		It("should have access to custom configuration", func() {
			log.Warningf("Warning %v", ptpConfig)
			Expect(ptpConfig.Namespace).To(Equal("openshift-ptp"))
		})
	})
	When("There is a Telecom Boundary Clock under test", func() {
		It("should pass G.8273.2 Noise Generation Requirements", func() {
			// if/when I have enough logs
			ptpContext := clients.NewContainerContext(ptpConfig.Namespace, ptpConfig.PodName, ptpConfig.Container)
			ptpDevInfo := devices.GetPTPDeviceInfo(ptpConfig.InterfaceName, ptpContext)
			if ptpDevInfo.VendorID != VendorIntel || ptpDevInfo.DeviceID != DeviceE810 {
				Skip("NIC device is not based on E810")
			}
			err := devices.GetPtpDeviceLogsToFile(ptpContext, time.Second*captureTIME, ptpLogs)
			Expect(err).NotTo(HaveOccurred())

			err = helperptp.ParsePtpLogs(ptpLogs, ptpLogsProcessed)
			Expect(err).NotTo(HaveOccurred())

			// define pod to calculate KPIs
			NgenImageName, err := images.GetNgenBaseImage()
			Expect(err).NotTo(HaveOccurred())
			Expect(NgenImageName).NotTo(BeNil())

			// calculate KPIs using system command for now
			out, err := exec.Command("podman", "run", "-it", "--rm", "-v", "/tmp:/tmp", NgenImageName, "-i", ptpLogsProcessed).Output() //nolint:stylecheck // TODO in next contribution
			Expect(err).NotTo(HaveOccurred())
			log.Info(string(out))
		})
	})
})
