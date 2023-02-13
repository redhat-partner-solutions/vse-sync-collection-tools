package devices

import (
	"path/filepath"
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
	// expecting a string like "../../../0000:86:00.0" here, just keep the last path section
	busID := commandWithCleanup(ctx, filepath.Base, []string{
		"readlink", "/sys/class/net/" + interfaceName + "/device",
	})

	devInfo.TtyGNSS = "/dev/" + busToGNSS(busID)
	log.Debugf("got busID for %s:  %s", interfaceName, busID)
	log.Debugf("got tty for %s:  %s", interfaceName, devInfo.TtyGNSS)

	// expecting a string like 0x1593
	devInfo.DeviceID = commandWithCleanup(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/device",
	})
	log.Debugf("got deviceID for %s:  %s", interfaceName, devInfo.DeviceID)

	// expecting a string like 0x8086
	devInfo.VendorID = commandWithCleanup(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/vendor",
	})
	log.Debugf("got vendorID for %s:  %s", interfaceName, devInfo.VendorID)
	return
}

// transform a bus ID to an expected GNSS TTY name.
// e.g. "0000:86:00.0" -> "ttyGNSS_8600", "0000:51:02.1" -> "ttyGNSS_5102"
func busToGNSS(busID string) (gnss string) {
	parts := strings.Split(busID, ":")
	ttyGNSS := parts[1] + strings.Split(parts[2], ".")[0]
	return "ttyGNSS_" + ttyGNSS
}

func commandWithCleanup(ctx clients.ContainerContext, cleanupFunc func(string) string, command []string) (result string) {
	clientset := clients.GetClientset()
	stdout, _, err := clientset.ExecCommandContainer(ctx, command)
	if err != nil {
		log.Error("command in container failed unexpectedly. context: %v", ctx)
		log.Error("command in container failed unexpectedly. command: %v", command)
		log.Error("command in container failed unexpectedly. error: %v", err)
		return ""
	}
	return cleanupFunc(stdout)
}

// Read from the ttyGNSS of the passed devInfo.
func ReadTtyGNSS(ctx clients.ContainerContext, devInfo PTPDeviceInfo) string {
	return commandWithCleanup(ctx, strings.TrimSpace, []string{
		"/bin/sh", "-c", "timeout 20 cat " + devInfo.TtyGNSS,
	})
}
