// SPDX-License-Identifier: GPL-2.0-or-later

package devices_test

import (
	"bufio"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/tools/remotecommand"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/testutils"
)

var _ = Describe("GetPMC", func() {
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

	When("called GetPMC", func() {
		It("should return a valid GMSettings", func() {
			expectedInput := "echo '<date>';date +%s.%N;echo '</date>';"
			expectedInput += "echo '<PMC>';pmc -u -f /var/run/ptp4l.0.config  'GET GRANDMASTER_SETTINGS_NP';echo '</PMC>';"

			expectedOutput := strings.Join([]string{
				"<date>",
				"1686916187.0584",
				"</date>",
				"<PMC>",
				"sending: GET GRANDMASTER_SETTINGS_NP",
				"	507c6f.fffe.30fbe8-0 seq 0 RESPONSE MANAGEMENT GRANDMASTER_SETTINGS_NP",
				"		clockClass              248",
				"		clockAccuracy           0xfe",
				"		offsetScaledLogVariance 0xffff",
				"		currentUtcOffset        37",
				"		leap61                  0",
				"		leap59                  0",
				"		currentUtcOffsetValid   0",
				"		ptpTimescale            1",
				"		timeTraceable           0",
				"		frequencyTraceable      0",
				"		timeSource              0xa0",
				"</PMC>",
			}, "\n")
			response[expectedInput] = []byte(expectedOutput)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())

			pmcInfo, err := devices.GetPMC(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(pmcInfo.Timestamp).To(Equal("2023-06-16T11:49:47.0584Z"))
			Expect(pmcInfo.ClockAccuracy).To(Equal("0xfe"))
			Expect(pmcInfo.ClockClass).To(Equal(248))
			Expect(pmcInfo.OffsetScaledLogVariance).To(Equal("0xffff"))
			Expect(pmcInfo.CurrentUtcOffset).To(Equal(37))
			Expect(pmcInfo.Leap61).To(Equal(0))
			Expect(pmcInfo.Leap59).To(Equal(0))
			Expect(pmcInfo.CurrentUtcOffsetValid).To(Equal(0))
			Expect(pmcInfo.PtpTimescale).To(Equal(1))
			Expect(pmcInfo.TimeTraceable).To(Equal(0))
			Expect(pmcInfo.FrequencyTraceable).To(Equal(0))
			Expect(pmcInfo.TimeSource).To(Equal("0xa0"))

		})
	})
})
