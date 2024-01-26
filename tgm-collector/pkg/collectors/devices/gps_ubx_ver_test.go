// SPDX-License-Identifier: GPL-2.0-or-later

package devices_test

import (
	"bufio"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/tools/remotecommand"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/testutils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
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
			expectedInput := "echo '<UBXMonVer>';ubxtool -t -p MON-VER -P 29.20;echo '</UBXMonVer>';"
			expectedInput += "echo '<UBXVersion>';ubxtool -V;echo '</UBXVersion>';"
			expectedInput += "echo '<GPSDVersion>';gpsd --version;echo '</GPSDVersion>';"
			expectedInput += "echo '<GNSSDevices>';ls -1 /dev | grep gnss;echo '</GNSSDevices>';"

			expectedOutput := strings.Join([]string{
				"<UBXMonVer>",
				"1689260332.4728",
				"UBX-MON-VER:",
				"  swVersion EXT CORE 1.00 (3fda8e)",
				"  hwVersion 00190000",
				"  extension ROM BASE 0x118B2060",
				"  extension FWVER=TIM 2.20",
				"  extension PROTVER=29.20",
				"  extension MOD=ZED-F9T",
				"  extension GPS;GLO;GAL;BDS",
				"  extension SBAS;QZSS",
				"  extension NAVIC",
				"",
				"</UBXMonVer>",
				"<UBXVersion>",
				"ubxtool: Version 3.25.1~dev",
				"</UBXVersion>",
				"<GPSDVersion>",
				"gpsd: 3.25.1~dev (revision release-3.25-109-g1a04cfab8)",
				"</GPSDVersion>",
				"<GNSSDevices>",
				"gnss0",
				"</GNSSDevices>",
				"",
			}, "\n")
			response[expectedInput] = []byte(expectedOutput)

			ctx, err := clients.NewContainerContext(clientset, "TestNamespace", "Test", "TestContainer")
			Expect(err).NotTo(HaveOccurred())

			gpsInfo, err := devices.GetGPSVersions(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(gpsInfo.Timestamp).To(Equal("2023-07-13T14:58:52.4728Z"))
			Expect(gpsInfo.FirmwareVersion).To(Equal("TIM 2.20"))
			Expect(gpsInfo.ProtoVersion).To(Equal("29.20"))
			Expect(gpsInfo.Module).To(Equal("ZED-F9T"))
			Expect(gpsInfo.UBXVersion).To(Equal("3.25.1~dev"))
			Expect(gpsInfo.GPSDVersion).To(Equal("3.25.1~dev (revision release-3.25-109-g1a04cfab8)"))
			Expect(gpsInfo.GNSSDevices).To(Equal([]string{"/dev/gnss0"}))

		})
	})
})
