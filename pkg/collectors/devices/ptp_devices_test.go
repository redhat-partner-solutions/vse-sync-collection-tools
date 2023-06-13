// SPDX-License-Identifier: GPL-2.0-or-later

package devices_test

import (
	"bufio"
	"fmt"
	"net/url"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/testutils"
)

var testPod = &v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "TestPod-8292",
		Namespace:   "TestNamespace",
		Annotations: map[string]string{},
	},
}

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

	When("called GetPTPDeviceInfo", func() {
		It("should return a valid PTPDeviceInfo", func() {
			vendor := "0x8086"
			devID := "0x1593"
			gnssDev := "gnss0"

			expectedInput := "echo '<date>';date --iso-8601=ns;echo '</date>';"
			expectedInput += "echo '<gnss>';ls /sys/class/net/aFakeInterface/device/gnss/;echo '</gnss>';"
			expectedInput += "echo '<devID>';cat /sys/class/net/aFakeInterface/device/device;echo '</devID>';"
			expectedInput += "echo '<vendorID>';cat /sys/class/net/aFakeInterface/device/vendor;echo '</vendorID>';"

			expectedOutput := "<date>\n1234\n</date>\n"
			expectedOutput += fmt.Sprintf("<gnss>\n%s\n</gnss>\n", gnssDev)
			expectedOutput += fmt.Sprintf("<devID>\n%s\n</devID>\n", devID)
			expectedOutput += fmt.Sprintf("<vendorID>\n%s\n</vendorID>\n", vendor)

			response[expectedInput] = []byte(expectedOutput)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			info, err := devices.GetPTPDeviceInfo("aFakeInterface", ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.DeviceID).To(Equal(devID))
			Expect(info.VendorID).To(Equal(vendor))
			Expect(info.GNSSDev).To(Equal("/dev/" + gnssDev))
		})
	})
	When("called GetDevDPLLInfo", func() {
		It("should return a valid PTPDeviceInfo", func() {
			eecState := "2"
			pssState := "10"
			offset := "-34"

			expectedInput := "echo '<date>';date --iso-8601=ns;echo '</date>';"
			expectedInput += "echo '<dpll_0_state>';cat /sys/class/net/aFakeInterface/device/dpll_0_state;echo '</dpll_0_state>';"
			expectedInput += "echo '<dpll_1_state>';cat /sys/class/net/aFakeInterface/device/dpll_1_state;echo '</dpll_1_state>';"
			expectedInput += "echo '<dpll_1_offset>';cat /sys/class/net/aFakeInterface/device/dpll_1_offset;echo '</dpll_1_offset>';"

			expectedOutput := "<date>\n1234\n</date>\n"
			expectedOutput += fmt.Sprintf("<dpll_0_state>\n%s\n</dpll_0_state>\n", eecState)
			expectedOutput += fmt.Sprintf("<dpll_1_state>\n%s\n</dpll_1_state>\n", pssState)
			expectedOutput += fmt.Sprintf("<dpll_1_offset>\n%s\n</dpll_1_offset>\n", offset)

			response[expectedInput] = []byte(expectedOutput)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			info, err := devices.GetDevDPLLInfo(ctx, "aFakeInterface")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.EECState).To(Equal(eecState))
			Expect(info.PPSState).To(Equal(pssState))
			Expect(info.PPSOffset).To(Equal(offset))
		})
	})
	When("called ReadGNSSDev", func() {
		It("should return a valid GNSSLines", func() {
			line := "definitely a line from a log"
			nLines := 1
			timeout := 10
			devInfo := devices.PTPDeviceInfo{
				GNSSDev: "/dev/gnss0",
			}

			expectedInput := "echo '<date>';date --iso-8601=ns;echo '</date>';"
			expectedInput += fmt.Sprintf(
				"echo '<lines>';timeout %d head -n %d %s;echo '</lines>';",
				timeout,
				nLines,
				devInfo.GNSSDev,
			)

			expectedOutput := "<date>\n1234\n</date>\n"
			expectedOutput += fmt.Sprintf("<lines>\n%s\n</lines>\n", line)

			response[expectedInput] = []byte(expectedOutput)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			GNSSLines, err := devices.ReadGNSSDev(ctx, devInfo, nLines, timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(GNSSLines.Dev).To(Equal(devInfo.GNSSDev))
			Expect(GNSSLines.Lines).To(Equal(line))
		})
	})
})

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devices Suite")
}
