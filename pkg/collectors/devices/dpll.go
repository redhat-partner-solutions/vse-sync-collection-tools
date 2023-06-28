// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/fetcher"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	unitConversionFactor = 100
)

type DevDPLLInfo struct {
	Timestamp string `json:"date" fetcherKey:"date"`
	EECState  string `json:"EECState" fetcherKey:"dpll_0_state"`
	PPSState  string `json:"PPSState" fetcherKey:"dpll_1_state"`
	PPSOffset string `json:"PPSOffset" fetcherKey:"dpll_1_offset"`
}

// AnalyserJSON returns the json expected by the analysers
func (dpllInfo *DevDPLLInfo) GetAnalyserFormat() (*callbacks.AnalyserFormatType, error) {
	offset, err := strconv.ParseFloat(dpllInfo.PPSOffset, 32)
	if err != nil {
		return &callbacks.AnalyserFormatType{}, fmt.Errorf("failed converting PPSOffset %w", err)
	}
	formatted := callbacks.AnalyserFormatType{
		ID: "dpll/time-error",
		Data: []string{
			dpllInfo.Timestamp,
			dpllInfo.EECState,
			dpllInfo.PPSState,
			fmt.Sprintf("%f", offset/unitConversionFactor),
		},
	}
	return &formatted, nil
}

var (
	dpllFetcher map[string]*fetcher.Fetcher
	dpplDateCmd *clients.Cmd
)

func formatTimestampAsRFC3339Nano(s string) (string, error) {
	timestamp, err := utils.ParseTimestamp(strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("failed to parse timestamp %w", err)
	}
	return timestamp.Format(time.RFC3339Nano), nil
}

func init() {
	dpllFetcher = make(map[string]*fetcher.Fetcher)
	dpplDateCmdInst, err := clients.NewCmd("date", "date +%s.%N")
	if err != nil {
		panic(err)
	}
	dpplDateCmd = dpplDateCmdInst
	dpplDateCmd.SetOutputProcessor(formatTimestampAsRFC3339Nano)
}

// BuildDPLLInfoFetcher popluates the fetcher required for
// collecting the DPLLInfo
func BuildDPLLInfoFetcher(interfaceName string) error {
	fetcherInst := fetcher.NewFetcher()
	dpllFetcher[interfaceName] = fetcherInst
	fetcherInst.AddCommand(dpplDateCmd)

	err := fetcherInst.AddNewCommand(
		"dpll_0_state",
		fmt.Sprintf("cat /sys/class/net/%s/device/dpll_0_state", interfaceName),
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "dpll_0_state", err.Error())
		return fmt.Errorf("failed to add command %s %w", "dpll_0_state", err)
	}

	err = fetcherInst.AddNewCommand(
		"dpll_1_state",
		fmt.Sprintf("cat /sys/class/net/%s/device/dpll_1_state", interfaceName),
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "dpll_1_state", err.Error())
		return fmt.Errorf("failed to add command %s %w", "dpll_1_state", err)
	}

	err = fetcherInst.AddNewCommand(
		"dpll_1_offset",
		fmt.Sprintf("cat /sys/class/net/%s/device/dpll_1_offset", interfaceName),
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "dpll_1_offset", err.Error())
		return fmt.Errorf("failed to add command %s %w", "dpll_1_offset", err)
	}
	return nil
}

// GetDevDPLLInfo returns the device DPLL info for an interface.
func GetDevDPLLInfo(ctx clients.ContainerContext, interfaceName string) (DevDPLLInfo, error) {
	dpllInfo := DevDPLLInfo{}
	fetcherInst, fetchedInstanceOk := dpllFetcher[interfaceName]
	if !fetchedInstanceOk {
		err := BuildDPLLInfoFetcher(interfaceName)
		if err != nil {
			return dpllInfo, err
		}
		fetcherInst, fetchedInstanceOk = dpllFetcher[interfaceName]
		if !fetchedInstanceOk {
			return dpllInfo, errors.New("failed to create fetcher for DPLLInfo")
		}
	}
	err := fetcherInst.Fetch(ctx, &dpllInfo)
	if err != nil {
		log.Errorf("failed to fetch dpllInfo %s", err.Error())
		return dpllInfo, fmt.Errorf("failed to fetch dpllInfo %w", err)
	}
	return dpllInfo, nil
}
