// SPDX-License-Identifier: GPL-2.0-or-later

package devices_test

import (
	"bufio"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/tools/remotecommand"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/testutils"
)

var _ = Describe("GetGPSNav", func() {
	var clientset *clients.Clientset
	var response map[string][]byte
	BeforeEach(func() { //nolint:dupl // this is test setup code
		clientset = testutils.GetMockedClientSet(testPod)
		response = make(map[string][]byte)
		responder := func(method string, url *url.URL, options remotecommand.StreamOptions) ([]byte, []byte, error) {
			reader := bufio.NewReader(options.Stdin)
			cmd := ""
			keepReading := true
			for keepReading {
				line, prefix, _ := reader.ReadLine()
				keepReading = prefix
				cmd += string(line)
			}
			return response[cmd], []byte(""), nil
		}
		clients.NewSPDYExecutor = testutils.NewFakeNewSPDYExecutor(responder, nil)
	})

	When("called GetGPSNav", func() {
		It("should return a valid GPSNav", func() {
			expectedInput := "echo '<GPS>';ubxtool -t -p NAV-STATUS -p NAV-CLOCK -P 29.20;echo '</GPS>';"

			expectedOutput := strings.Join([]string{
				"<GPS>",
				"1686916187.0584",
				"UBX-NAV-STATUS:",
				"  iTOW 474605000 gpsFix 3 flags 0xdd fixStat 0x0 flags2 0x8",
				"  ttff 25030, msss 4294967295",
				"",
				"1686916187.0586",
				"UBX-NAV-CLOCK:",
				"  iTOW 474605000 clkB 61594 clkD 56 tAcc 5 fAcc 164",
				"",
				"</GPS>",
			}, "\n")
			response[expectedInput] = []byte(expectedOutput)

			expectedTimestampStatus, err := time.Parse(time.RFC3339Nano, "2023-06-16T11:49:47.0584Z")
			Expect(err).NotTo(HaveOccurred())
			expectedTimestampClock, err := time.Parse(time.RFC3339Nano, "2023-06-16T11:49:47.0586Z")
			Expect(err).NotTo(HaveOccurred())
			local, err := time.LoadLocation("Local")
			Expect(err).NotTo(HaveOccurred())
			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())

			gpsInfo, err := devices.GetGPSNav(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(gpsInfo.TimestampStatus).To(Equal(expectedTimestampStatus.In(local).Format(time.RFC3339Nano)))
			Expect(gpsInfo.TimestampClock).To(Equal(expectedTimestampClock.In(local).Format(time.RFC3339Nano)))
			Expect(gpsInfo.GPSFix).To(Equal("3"))
			Expect(gpsInfo.TimeAcc).To(Equal("5"))
			Expect(gpsInfo.FreqAcc).To(Equal("164"))
		})
	})
})
