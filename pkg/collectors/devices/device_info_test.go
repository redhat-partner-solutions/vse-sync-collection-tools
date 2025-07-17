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

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/constants"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/testutils"
)

const ethtoolOutput = `<ethtoolOut>
driver: ice
version: %s
firmware-version: %s
expansion-rom-version:
bus-info: 0000:86:00.0
supports-statistics: yes
supports-test: yes
supports-eeprom-access: yes
supports-register-dump: yes
supports-priv-flags: yes
</ethtoolOut>
`

var testPod = &v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "TestPod-8292",
		Namespace:   "TestNamespace",
		Annotations: map[string]string{},
	},
	Spec: v1.PodSpec{
		NodeName: "TestNodeName",
	},
}

var _ = Describe("NewContainerContext", func() {
	type Response struct {
		err    error
		stdout string
		stderr string
	}

	var clientset *clients.Clientset
	var response map[string]Response
	BeforeEach(func() { //nolint:dupl // this is test setup code
		clientset = testutils.GetMockedClientSet(testPod)
		response = make(map[string]Response)
		responder := func(method string, url *url.URL, options remotecommand.StreamOptions) ([]byte, []byte, error) {
			reader := bufio.NewReader(options.Stdin)
			cmd := ""
			keepReading := true
			for keepReading {
				line, prefix, _ := reader.ReadLine()
				keepReading = prefix
				cmd += string(line)
			}
			resp, ok := response[cmd]
			if !ok {
				return []byte(resp.stdout), []byte(resp.stderr), fmt.Errorf("Response not found")
			}
			return []byte(resp.stdout), []byte(resp.stderr), resp.err
		}
		clients.NewSPDYExecutor = testutils.NewFakeNewSPDYExecutor(responder, nil)
		devices.ClearDevFetcher()
	})

	When("called GetPTPDeviceInfo", func() {
		It("should return a valid DeviceInfo for GM clock type", func() {
			vendor := "0x8086"
			devID := "0x1593"
			gnssDev := "gnss0"
			firmwareVersion := "4.20 0x8001778b 1.3346.0"
			driverVersion := "1.11.20.7"

			response["ls /sys/class/net/aFakeInterface/device/gnss/"] = Response{stdout: gnssDev}

			expectedInput := "echo '<date>';date +%s.%N;echo '</date>';"
			expectedInput += "echo '<gnss>';ls /sys/class/net/aFakeInterface/device/gnss/;echo '</gnss>';"
			expectedInput += "echo '<devID>';cat /sys/class/net/aFakeInterface/device/device;echo '</devID>';"
			expectedInput += "echo '<vendorID>';cat /sys/class/net/aFakeInterface/device/vendor;echo '</vendorID>';"
			expectedInput += "echo '<ethtoolOut>';ethtool -i aFakeInterface;echo '</ethtoolOut>';"

			expectedOutput := "<date>\n1686916187.0584\n</date>\n"
			expectedOutput += fmt.Sprintf("<gnss>\n%s\n</gnss>\n", gnssDev)
			expectedOutput += fmt.Sprintf("<devID>\n%s\n</devID>\n", devID)
			expectedOutput += fmt.Sprintf("<vendorID>\n%s\n</vendorID>\n", vendor)
			expectedOutput += fmt.Sprintf(ethtoolOutput, driverVersion, firmwareVersion)

			response[expectedInput] = Response{stdout: expectedOutput}

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer", "TestNodeName")
			Expect(err).NotTo(HaveOccurred())
			info, err := devices.GetPTPDeviceInfo("aFakeInterface", ctx, constants.ClockTypeGM)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Timestamp).To(Equal("2023-06-16T11:49:47.0584Z"))
			Expect(info.DeviceID).To(Equal(devID))
			Expect(info.VendorID).To(Equal(vendor))
			Expect(info.GNSSDev).To(Equal("/dev/" + gnssDev))
			Expect(info.FirmwareVersion).To(Equal(firmwareVersion))
			Expect(info.DriverVersion).To(Equal(driverVersion))

		})
	})

	When("called GetPTPDeviceInfo with BC clock type", func() {
		It("should return a valid DeviceInfo without GNSS for BC", func() {
			vendor := "0x8086"
			devID := "0x1593"
			firmwareVersion := "4.20 0x8001778b 1.3346.0"
			driverVersion := "1.11.20.7"

			expectedInput := "echo '<date>';date +%s.%N;echo '</date>';"
			expectedInput += "echo '<devID>';cat /sys/class/net/aFakeInterface/device/device;echo '</devID>';"
			expectedInput += "echo '<vendorID>';cat /sys/class/net/aFakeInterface/device/vendor;echo '</vendorID>';"
			expectedInput += "echo '<ethtoolOut>';ethtool -i aFakeInterface;echo '</ethtoolOut>';"

			expectedOutput := "<date>\n1686916187.0584\n</date>\n"
			expectedOutput += fmt.Sprintf("<devID>\n%s\n</devID>\n", devID)
			expectedOutput += fmt.Sprintf("<vendorID>\n%s\n</vendorID>\n", vendor)
			expectedOutput += fmt.Sprintf(ethtoolOutput, driverVersion, firmwareVersion)

			response[expectedInput] = Response{stdout: expectedOutput}

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer", "TestNodeName")
			Expect(err).NotTo(HaveOccurred())
			info, err := devices.GetPTPDeviceInfo("aFakeInterface", ctx, constants.ClockTypeBC)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Timestamp).To(Equal("2023-06-16T11:49:47.0584Z"))
			Expect(info.DeviceID).To(Equal(devID))
			Expect(info.VendorID).To(Equal(vendor))
			Expect(info.GNSSDev).To(Equal(""))
			Expect(info.FirmwareVersion).To(Equal(firmwareVersion))
			Expect(info.DriverVersion).To(Equal(driverVersion))

		})
	})

	When("called GetPTPDeviceInfo with GM but GNSS fails", func() {
		It("should return a valid DeviceInfo with empty GNSS when GNSS command fails", func() {
			vendor := "0x8086"
			devID := "0x1593"
			firmwareVersion := "4.20 0x8001778b 1.3346.0"
			driverVersion := "1.11.20.7"

			response["ls /sys/class/net/aFakeInterface/device/gnss/"] = Response{err: fmt.Errorf("Not found")}

			expectedInput := "echo '<date>';date +%s.%N;echo '</date>';"
			expectedInput += "echo '<devID>';cat /sys/class/net/aFakeInterface/device/device;echo '</devID>';"
			expectedInput += "echo '<vendorID>';cat /sys/class/net/aFakeInterface/device/vendor;echo '</vendorID>';"
			expectedInput += "echo '<ethtoolOut>';ethtool -i aFakeInterface;echo '</ethtoolOut>';"

			expectedOutput := "<date>\n1686916187.0584\n</date>\n"
			expectedOutput += fmt.Sprintf("<devID>\n%s\n</devID>\n", devID)
			expectedOutput += fmt.Sprintf("<vendorID>\n%s\n</vendorID>\n", vendor)
			expectedOutput += fmt.Sprintf(ethtoolOutput, driverVersion, firmwareVersion)

			response[expectedInput] = Response{stdout: expectedOutput}

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer", "TestNodeName")
			Expect(err).NotTo(HaveOccurred())
			info, err := devices.GetPTPDeviceInfo("aFakeInterface", ctx, constants.ClockTypeGM)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Timestamp).To(Equal("2023-06-16T11:49:47.0584Z"))
			Expect(info.DeviceID).To(Equal(devID))
			Expect(info.VendorID).To(Equal(vendor))
			Expect(info.GNSSDev).To(Equal(""))
			Expect(info.FirmwareVersion).To(Equal(firmwareVersion))
			Expect(info.DriverVersion).To(Equal(driverVersion))

		})
	})
})

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devices Suite")
}
