// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"bytes"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type PTPDeviceInfo struct {
	Timestamp string `json:"date"`
	VendorID  string `json:"vendorId"`
	DeviceID  string `json:"deviceInfo"`
	GNSSDev   string `json:"GNSSDev"` //nolint:tagliatelle // Because GNSS is an ancronym
}

type DevDPLLInfo struct {
	Timestamp string `json:"date"`
	EECState  string `json:"EECState"`  //nolint:tagliatelle // Because EEC is dpll name
	PPSState  string `json:"PPSState"`  //nolint:tagliatelle // Because PPS is dpll name
	PPSOffset string `json:"PPSOffset"` //nolint:tagliatelle // Because PPS is dpll name
}
type GNSSDevLines struct {
	Timestamp string `json:"date"`
	Dev       string `json:"dev"`
	Lines     string `json:"lines"`
}

func GetPTPDeviceInfo(interfaceName string, ctx clients.ContainerContext) (devInfo PTPDeviceInfo) {
	// Find the dev for the GNSS for this interface
	cmds := &clients.CmdGroup{}

	dateCmd, err := clients.NewCmd("date", []string{"date +%s%N"})
	if err != nil {
		panic(err)
	}
	dateCmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(dateCmd)

	gnssCmd, err := clients.NewCmd("gnss", []string{
		"ls", "/sys/class/net/" + interfaceName + "/device/gnss/",
	})
	if err != nil {
		panic(err)
	}
	gnssCmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(gnssCmd)

	devIDCmd, err := clients.NewCmd("devID", []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/device",
	})
	if err != nil {
		panic(err)
	}
	devIDCmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(devIDCmd)

	vendorIDCmd, err := clients.NewCmd("vendorID", []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/vendor",
	})
	if err != nil {
		panic(err)
	}
	vendorIDCmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(vendorIDCmd)

	collectedValues := runCommands(ctx, cmds)

	devInfo.GNSSDev = "/dev/" + collectedValues["gnss"]
	log.Debugf("got dev for %s:  %s", interfaceName, devInfo.GNSSDev)

	// expecting a string like 0x1593
	devInfo.DeviceID = collectedValues["devID"]
	log.Debugf("got deviceID for %s:  %s", interfaceName, devInfo.DeviceID)

	// expecting a string like 0x8086
	devInfo.VendorID = collectedValues["vendorID"]
	log.Debugf("got vendorID for %s:  %s", interfaceName, devInfo.VendorID)

	devInfo.Timestamp = collectedValues["date"]
	return
}

func runCommands(ctx clients.ContainerContext, cmdGrp clients.Cmder) (result map[string]string) { //nolint:lll // allow slightly long function definition
	clientset := clients.GetClientset()
	cmd := cmdGrp.GetCommand()
	command := []string{"/usr/bin/sh"}
	var buffIn bytes.Buffer
	buffIn.WriteString(strings.Join(cmd, " "))

	stdout, _, err := clientset.ExecCommandContainerStdIn(ctx, command, buffIn)
	if err != nil {
		log.Errorf("command in container failed unexpectedly. context: %v", ctx)
		log.Errorf("command in container failed unexpectedly. command: %v", command)
		log.Errorf("command in container failed unexpectedly. error: %v", err)
		return result
	}
	result, err = cmdGrp.ExtractResult(stdout)
	if err != nil {
		log.Errorf("extraction failed %s", err.Error())
		log.Errorf("output was %s", stdout)
	}
	return
}

// Read lines from the GNSSDev of the passed devInfo.
func ReadGNSSDev(ctx clients.ContainerContext, devInfo PTPDeviceInfo, lines, timeoutSeconds int) GNSSDevLines {
	cmds := &clients.CmdGroup{}

	dateCmd, err := clients.NewCmd("date", []string{"date +%s%N"})
	if err != nil {
		panic(err)
	}
	dateCmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(dateCmd)

	linesCmd, err := clients.NewCmd("lines", []string{
		"timeout", strconv.Itoa(timeoutSeconds), "head", "-n", strconv.Itoa(lines), devInfo.GNSSDev,
	})
	if err != nil {
		panic(err)
	}
	linesCmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(linesCmd)

	collectedValues := runCommands(ctx, cmds)

	return GNSSDevLines{
		Timestamp: collectedValues["date"],
		Dev:       devInfo.GNSSDev,
		Lines:     collectedValues["lines"],
	}
}

// GetDevDPLLInfo returns the device DPLL info for an interface.
func GetDevDPLLInfo(ctx clients.ContainerContext, interfaceName string) (dpllInfo DevDPLLInfo) {
	cmds := &clients.CmdGroup{}

	dateCmd, err := clients.NewCmd("date", []string{"date +%s%N"})
	if err != nil {
		panic(err)
	}
	dateCmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(dateCmd)

	dpllState0Cmd, err := clients.NewCmd("dpll_0_state", []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_0_state",
	})
	if err != nil {
		panic(err)
	}
	dpllState0Cmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(dpllState0Cmd)

	dpllState1Cmd, err := clients.NewCmd("dpll_1_state", []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_state",
	})
	if err != nil {
		panic(err)
	}
	dpllState1Cmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(dpllState1Cmd)

	dpllOffset1Cmd, err := clients.NewCmd("dpll_1_offset", []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_offset",
	})
	if err != nil {
		panic(err)
	}
	dpllOffset1Cmd.SetCleanupFunc(strings.TrimSpace)
	cmds.AddCommand(dpllOffset1Cmd)

	collectedValues := runCommands(ctx, cmds)

	dpllInfo.Timestamp = collectedValues["date"]
	dpllInfo.EECState = collectedValues["dpll_0_state"]
	dpllInfo.PPSState = collectedValues["dpll_1_state"]
	dpllInfo.PPSOffset = collectedValues["dpll_1_offset"]
	return
}
