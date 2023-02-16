package devices

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	log "github.com/sirupsen/logrus"
)

type PTPDeviceInfo struct {
	VendorID string
	DeviceID string
	TtyGNSS  string
}

func GetPTPDeviceInfo(interfaceName string, ctx clients.ContainerContext) (devInfo PTPDeviceInfo) {
	// expecting a string like "../../../0000:86:00.0" here, just keep the last path section with filepath.Base
	busID := commandWithPostprocessFunc(ctx, filepath.Base, []string{
		"readlink", "/sys/class/net/" + interfaceName + "/device",
	})

	devInfo.TtyGNSS = "/dev/" + busToGNSS(busID)
	log.Debugf("got busID for %s:  %s", interfaceName, busID)
	log.Debugf("got tty for %s:  %s", interfaceName, devInfo.TtyGNSS)

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

// transform a bus ID to an expected GNSS TTY name.
// e.g. "0000:86:00.0" -> "ttyGNSS_8600", "0000:51:02.1" -> "ttyGNSS_5102"
func busToGNSS(busID string) (gnss string) {
	log.Debugf("convert %s to GNSS tty", busID)
	parts := strings.Split(busID, ":")
	ttyGNSS := parts[1] + strings.Split(parts[2], ".")[0]
	return "ttyGNSS_" + ttyGNSS
}

func commandWithPostprocessFunc(ctx clients.ContainerContext, cleanupFunc func(string) string, command []string) (result string) {
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

// Read lines from the ttyGNSS of the passed devInfo.
func ReadTtyGNSS(ctx clients.ContainerContext, devInfo PTPDeviceInfo, lines, timeoutSeconds int) string {
	return commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"timeout", strconv.Itoa(timeoutSeconds),  "head", "-n", strconv.Itoa(lines), devInfo.TtyGNSS,
	})
}
