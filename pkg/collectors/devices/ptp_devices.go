// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type PTPDeviceInfo struct {
	VendorID string `json:"vendorId"`
	DeviceID string `json:"deviceInfo"`
	GNSSDev  string `json:"GNSSDev"` //nolint:tagliatelle // Because GNSS is an ancronym
}

type DevDPLLInfo struct {
	EECState  string `json:"EECState"`  //nolint:tagliatelle // Because EEC is dpll name
	PPSState  string `json:"PPSState"`  //nolint:tagliatelle // Because PPS is dpll name
	PPSOffset string `json:"PPSOffset"` //nolint:tagliatelle // Because PPS is dpll name
}
type GNSSDevLines struct {
	Dev   string `json:"dev"`
	Lines string `json:"lines"`
}

func GetPTPDeviceInfo(interfaceName string, ctx clients.ContainerContext) (devInfo PTPDeviceInfo) {
	// Find the dev for the GNSS for this interface
	gnssDev := commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"ls", "/sys/class/net/" + interfaceName + "/device/gnss/",
	})

	devInfo.GNSSDev = "/dev/" + gnssDev
	log.Debugf("got dev for %s:  %s", interfaceName, devInfo.GNSSDev)

	// expecting a string like 0x1593
	devInfo.DeviceID = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/device",
	})
	log.Debugf("got deviceID for %s:  %s", interfaceName, devInfo.DeviceID)

	// expecting a string like 0x8086
	devInfo.VendorID = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/vendor",
	})
	log.Debugf("got vendorID for %s:  %s", interfaceName, devInfo.VendorID)
	return
}

func commandWithPostprocessFunc(ctx clients.ContainerContext, cleanupFunc func(string) string, command []string) (result string) { //nolint:lll // allow slightly long function definition
	clientset := clients.GetClientset()
	stdout, _, err := clientset.ExecCommandContainer(ctx, command)
	if err != nil {
		log.Errorf("command in container failed unexpectedly. context: %v", ctx)
		log.Errorf("command in container failed unexpectedly. command: %v", command)
		log.Errorf("command in container failed unexpectedly. error: %v", err)
		return ""
	}
	return cleanupFunc(stdout)
}

// Read lines from the GNSSDev of the passed devInfo.
func ReadGNSSDev(ctx clients.ContainerContext, devInfo PTPDeviceInfo, lines, timeoutSeconds int) GNSSDevLines {
	content := commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"timeout", strconv.Itoa(timeoutSeconds), "head", "-n", strconv.Itoa(lines), devInfo.GNSSDev,
	})
	return GNSSDevLines{
		Dev:   devInfo.GNSSDev,
		Lines: content,
	}
}

// GetDevDPLLInfo returns the device DPLL info for an interface.
func GetDevDPLLInfo(ctx clients.ContainerContext, interfaceName string) (dpllInfo DevDPLLInfo) {
	dpllInfo.EECState = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_0_state",
	})
	dpllInfo.PPSState = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_state",
	})
	dpllInfo.PPSOffset = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_offset",
	})
	return
}
