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
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/testutils"
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

			expectedInput := "echo '<date>';date +%s.%N;echo '</date>';"
			expectedInput += "echo '<gnss>';ls /sys/class/net/aFakeInterface/device/gnss/;echo '</gnss>';"
			expectedInput += "echo '<devID>';cat /sys/class/net/aFakeInterface/device/device;echo '</devID>';"
			expectedInput += "echo '<vendorID>';cat /sys/class/net/aFakeInterface/device/vendor;echo '</vendorID>';"

			expectedOutput := "<date>\n1686916187.0584\n</date>\n"
			expectedOutput += fmt.Sprintf("<gnss>\n%s\n</gnss>\n", gnssDev)
			expectedOutput += fmt.Sprintf("<devID>\n%s\n</devID>\n", devID)
			expectedOutput += fmt.Sprintf("<vendorID>\n%s\n</vendorID>\n", vendor)

			response[expectedInput] = []byte(expectedOutput)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			info, err := devices.GetPTPDeviceInfo("aFakeInterface", ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Timestamp).To(Equal("2023-06-16T11:49:47.0584Z"))
			Expect(info.DeviceID).To(Equal(devID))
			Expect(info.VendorID).To(Equal(vendor))
			Expect(info.GNSSDev).To(Equal("/dev/" + gnssDev))
		})
	})
})

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devices Suite")
}
