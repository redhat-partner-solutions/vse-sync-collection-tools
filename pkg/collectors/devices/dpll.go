// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"errors"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/fetcher"
)

const (
	unitConversionFactor = 100
)

type DevDPLLInfo struct {
	Timestamp string  `fetcherKey:"date"          json:"timestamp"`
	EECState  string  `fetcherKey:"dpll_0_state"  json:"eecstate"`
	PPSState  string  `fetcherKey:"dpll_1_state"  json:"state"`
	PPSOffset float64 `fetcherKey:"dpll_1_offset" json:"terror"`
}

// AnalyserJSON returns the json expected by the analysers
func (dpllInfo *DevDPLLInfo) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID: "dpll/time-error",
		Data: map[string]any{
			"timestamp": dpllInfo.Timestamp,
			"eecstate":  dpllInfo.EECState,
			"state":     dpllInfo.PPSState,
			"terror":    dpllInfo.PPSOffset / unitConversionFactor,
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

var (
	dpllFetcher map[string]*fetcher.Fetcher
)

func init() {
	dpllFetcher = make(map[string]*fetcher.Fetcher)
}

func postProcessDPLL(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any)
	offset, err := strconv.ParseFloat(result["dpll_1_offset"], 32)
	if err != nil {
		return processedResult, fmt.Errorf("failed converting dpll_1_offset %w to an int", err)
	}
	processedResult["dpll_1_offset"] = offset
	return processedResult, nil
}

// BuildDPLLInfoFetcher popluates the fetcher required for
// collecting the DPLLInfo
func BuildDPLLInfoFetcher(interfaceName string) error { //nolint:dupl // Further dedup risks be too abstract or fragile
	fetcherInst, err := fetcher.FetcherFactory(
		[]*clients.Cmd{dateCmd},
		[]fetcher.AddCommandArgs{
			{
				Key:     "dpll_0_state",
				Command: fmt.Sprintf("cat /sys/class/net/%s/device/dpll_0_state", interfaceName),
				Trim:    true,
			},
			{
				Key:     "dpll_1_state",
				Command: fmt.Sprintf("cat /sys/class/net/%s/device/dpll_1_state", interfaceName),
				Trim:    true,
			},
			{
				Key:     "dpll_1_offset",
				Command: fmt.Sprintf("cat /sys/class/net/%s/device/dpll_1_offset", interfaceName),
				Trim:    true,
			},
		},
	)
	if err != nil {
		log.Errorf("failed to create fetcher for dpll: %s", err.Error())
		return fmt.Errorf("failed to create fetcher for dpll: %w", err)
	}
	dpllFetcher[interfaceName] = fetcherInst
	fetcherInst.SetPostProcessor(postProcessDPLL)
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
		log.Debugf("failed to fetch dpllInfo %s", err.Error())
		return dpllInfo, fmt.Errorf("failed to fetch dpllInfo %w", err)
	}
	return dpllInfo, nil
}
