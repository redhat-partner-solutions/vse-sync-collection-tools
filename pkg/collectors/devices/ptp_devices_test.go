// SPDX-License-Identifier: GPL-2.0-or-later

package devices_test

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	BeforeEach(func() {
		clientset = testutils.GetMockedClientSet(testPod)
		response = make(map[string][]byte)
		responder := func(method string, url *url.URL) ([]byte, []byte, error) {
			cmd := url.Query()["command"]
			return response[strings.Join(cmd, " ")], []byte(""), nil
		}
		clients.NewSPDYExecutor = testutils.NewFakeNewSPDYExecutor(responder, nil)
	})

	When("called GetPTPDeviceInfo", func() {
		It("should return a valid PTPDeviceInfo", func() {
			vendor := "0x8086"
			devID := "0x1593"
			gnssDev := "gnss0"

			response["ls /sys/class/net/aFakeInterface/device/gnss/"] = []byte(gnssDev)
			response["cat /sys/class/net/aFakeInterface/device/device"] = []byte(devID)
			response["cat /sys/class/net/aFakeInterface/device/vendor"] = []byte(vendor)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			info := devices.GetPTPDeviceInfo("aFakeInterface", ctx)
			Expect(info.DeviceID).To(Equal(devID))
			Expect(info.VendorID).To(Equal(vendor))
			Expect(info.GNSSDev).To(Equal("/dev/" + gnssDev))
		})
	})
	When("called GetDevDPLLInfo", func() {
		It("should return a valid PTPDeviceInfo", func() {
			state := "10"
			offset := "-34"

			response["cat /sys/class/net/aFakeInterface/device/dpll_1_state"] = []byte(state)
			response["cat /sys/class/net/aFakeInterface/device/dpll_1_offset"] = []byte(offset)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			info := devices.GetDevDPLLInfo(ctx, "aFakeInterface")
			Expect(info.State).To(Equal(state))
			Expect(info.Offset).To(Equal(offset))

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
			key := fmt.Sprintf("timeout %d head -n %s %s", timeout, strconv.Itoa(nLines), devInfo.GNSSDev)

			response[key] = []byte(line)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())
			GNSSLines := devices.ReadGNSSDev(ctx, devInfo, nLines, timeout)
			Expect(GNSSLines.Dev).To(Equal(devInfo.GNSSDev))
			Expect(GNSSLines.Lines).To(Equal(line))
		})
	})
})

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devices Suite")
}
