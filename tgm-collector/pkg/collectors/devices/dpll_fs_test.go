// SPDX-License-Identifier: GPL-2.0-or-later

package devices_test

import (
	"bufio"
	"fmt"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/tools/remotecommand"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/testutils"
)

var _ = Describe("NewContainerContext", func() {
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
	When("called GetDevDPLLInfo", func() {
		It("should return a valid DevDPLLInfo", func() {
			eecState := "2"
			pssState := "10"
			offset := float64(-34)

			expectedInput := "echo '<date>';date +%s.%N;echo '</date>';"
			expectedInput += "echo '<dpll_0_state>';cat /sys/class/net/aFakeInterface/device/dpll_0_state;echo '</dpll_0_state>';"
			expectedInput += "echo '<dpll_1_state>';cat /sys/class/net/aFakeInterface/device/dpll_1_state;echo '</dpll_1_state>';"
			expectedInput += "echo '<dpll_1_offset>';cat /sys/class/net/aFakeInterface/device/dpll_1_offset;echo '</dpll_1_offset>';"

			expectedOutput := "<date>\n1686916187.0584\n</date>\n"
			expectedOutput += fmt.Sprintf("<dpll_0_state>\n%s\n</dpll_0_state>\n", eecState)
			expectedOutput += fmt.Sprintf("<dpll_1_state>\n%s\n</dpll_1_state>\n", pssState)
			expectedOutput += fmt.Sprintf("<dpll_1_offset>\n%f\n</dpll_1_offset>\n", offset)

			response[expectedInput] = []byte(expectedOutput)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			info, err := devices.GetDevDPLLFilesystemInfo(ctx, "aFakeInterface")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Timestamp).To(Equal("2023-06-16T11:49:47.0584Z"))
			Expect(info.EECState).To(Equal(eecState))
			Expect(info.PPSState).To(Equal(pssState))
			Expect(info.PPSOffset).To(Equal(offset))
		})
	})
})
