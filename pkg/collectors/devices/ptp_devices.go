// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type PTPDeviceInfo struct {
	Timestamp string `json:"date" fetcherKey:"date"`
	VendorID  string `json:"vendorId" fetcherKey:"vendorID"`
	DeviceID  string `json:"deviceInfo" fetcherKey:"devID"`
	GNSSDev   string `json:"GNSSDev" fetcherKey:"gnss"` //nolint:tagliatelle // Because GNSS is an ancronym
}

type DevDPLLInfo struct {
	Timestamp string `json:"date" fetcherKey:"date"`
	EECState  string `json:"EECState" fetcherKey:"dpll_0_state"`   //nolint:tagliatelle // Because EEC is dpll name
	PPSState  string `json:"PPSState" fetcherKey:"dpll_1_state"`   //nolint:tagliatelle // Because PPS is dpll name
	PPSOffset string `json:"PPSOffset" fetcherKey:"dpll_1_offset"` //nolint:tagliatelle // Because PPS is dpll name
}
type GNSSDevLines struct {
	Timestamp string `json:"date" fetcherKey:"date"`
	Dev       string `json:"dev"`
	Lines     string `json:"lines" fetcherKey:"lines"`
}

var (
	devFetcher  map[string]*fetcher
	gnssFetcher map[string]*fetcher
	dpllFetcher map[string]*fetcher
	dateCmd     *clients.Cmd
)

func init() {
	devFetcher = make(map[string]*fetcher)
	gnssFetcher = make(map[string]*fetcher)
	dpllFetcher = make(map[string]*fetcher)
	dateCmdInst, err := clients.NewCmd("date", []string{"date +%s.%N"})
	if err != nil {
		panic(err)
	}
	dateCmd = dateCmdInst
	dateCmd.SetCleanupFunc(strings.TrimSpace)
}

func ifCmdErrorLog(key string, err error) {
	if err != nil {
		log.Errorf("failed to add command %s %s", key, err.Error())
	}
}

func GetPTPDeviceInfo(interfaceName string, ctx clients.ContainerContext) (devInfo PTPDeviceInfo) {
	// Find the dev for the GNSS for this interface
	fetcherInst, ok := devFetcher[interfaceName]
	if !ok {
		fetcherInst = NewFetcher()
		devFetcher[interfaceName] = fetcherInst

		fetcherInst.AddCommand(dateCmd)

		err := fetcherInst.AddNewCommand("gnss", []string{
			"ls", "/sys/class/net/" + interfaceName + "/device/gnss/",
		}, true)
		ifCmdErrorLog("gnss", err)

		err = fetcherInst.AddNewCommand("devID", []string{
			"cat", "/sys/class/net/" + interfaceName + "/device/device",
		}, true)
		ifCmdErrorLog("devId", err)

		err = fetcherInst.AddNewCommand("vendorID", []string{
			"cat", "/sys/class/net/" + interfaceName + "/device/vendor",
		}, true)
		ifCmdErrorLog("vendorID", err)
	}

	err := fetcherInst.Fetch(ctx, &devInfo)
	if err != nil {
		log.Errorf("failed to fetch devInfo %s", err.Error())
	}
	devInfo.GNSSDev = "/dev/" + devInfo.GNSSDev
	return devInfo
}

// Read lines from the GNSSDev of the passed devInfo.
func ReadGNSSDev(ctx clients.ContainerContext, devInfo PTPDeviceInfo, lines, timeoutSeconds int) GNSSDevLines {
	fetcherInst, ok := gnssFetcher[devInfo.GNSSDev]
	if !ok {
		fetcherInst = NewFetcher()
		gnssFetcher[devInfo.GNSSDev] = fetcherInst

		fetcherInst.AddCommand(dateCmd)

		err := fetcherInst.AddNewCommand("lines", []string{
			"timeout", strconv.Itoa(timeoutSeconds), "head", "-n", strconv.Itoa(lines), devInfo.GNSSDev,
		}, true)
		ifCmdErrorLog("lines", err)
	}

	gnssInfo := GNSSDevLines{
		Dev: devInfo.GNSSDev,
	}
	err := fetcherInst.Fetch(ctx, &gnssInfo)
	if err != nil {
		log.Errorf("failed to fetch gnssInfo %s", err.Error())
	}
	return gnssInfo
}

// GetDevDPLLInfo returns the device DPLL info for an interface.
func GetDevDPLLInfo(ctx clients.ContainerContext, interfaceName string) (dpllInfo DevDPLLInfo) {
	fetcherInst, ok := dpllFetcher[interfaceName]
	if !ok {
		fetcherInst = NewFetcher()
		dpllFetcher[interfaceName] = fetcherInst

		fetcherInst.AddCommand(dateCmd)

		err := fetcherInst.AddNewCommand("dpll_0_state", []string{
			"cat", "/sys/class/net/" + interfaceName + "/device/dpll_0_state",
		}, true)
		ifCmdErrorLog("dpll_0_state", err)

		err = fetcherInst.AddNewCommand("dpll_1_state", []string{
			"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_state",
		}, true)
		ifCmdErrorLog("dpll_1_state", err)

		err = fetcherInst.AddNewCommand("dpll_1_offset", []string{
			"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_offset",
		}, true)
		ifCmdErrorLog("dpll_1_offset", err)
	}
	err := fetcherInst.Fetch(ctx, &dpllInfo)
	if err != nil {
		log.Errorf("failed to fetch dpllInfo %s", err.Error())
	}
	return
}
