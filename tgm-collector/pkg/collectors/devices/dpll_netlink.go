// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/fetcher"
)

var states = map[string]string{
	"unknown":       "-1",
	"invalid":       "0",
	"freerun":       "1",
	"locked":        "2",
	"locked-ho-acq": "3",
	"holdover":      "4",
}

type DevNetlinkDPLLInfo struct {
	Timestamp string `fetcherKey:"date" json:"timestamp"`
	EECState  string `fetcherKey:"eec"  json:"eecstate"`
	PPSState  string `fetcherKey:"pps"  json:"state"`
}

// AnalyserJSON returns the json expected by the analysers
func (dpllInfo *DevNetlinkDPLLInfo) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID: "dpll/states",
		Data: map[string]any{
			"timestamp": dpllInfo.Timestamp,
			"eecstate":  dpllInfo.EECState,
			"state":     dpllInfo.PPSState,
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

type NetlinkEntry struct {
	LockStatus string `json:"lock-status"` //nolint:tagliatelle // not my choice
	Driver     string `json:"module-name"` //nolint:tagliatelle // not my choice
	ClockType  string `json:"type"`        //nolint:tagliatelle // not my choice
	ClockID    int64  `json:"clock-id"`    //nolint:tagliatelle // not my choice
	ID         int    `json:"id"`          //nolint:tagliatelle // not my choice
}

// # Example output
// [{'clock-id': 5799633565435100136,
//   'id': 0,
//   'lock-status': 'locked-ho-acq',
//   'mode': 'automatic',
//   'mode-supported': ['automatic'],
//   'module-name': 'ice',
//   'type': 'eec'},
//  {'clock-id': 5799633565435100136,
//   'id': 1,
//   'lock-status': 'locked-ho-acq',
//   'mode': 'automatic',
//   'mode-supported': ['automatic'],
//   'module-name': 'ice',
//   'type': 'pps'}]

var (
	dpllNetlinkFetcher map[int64]*fetcher.Fetcher
	dpllClockIDFetcher map[string]*fetcher.Fetcher
)

func init() {
	dpllNetlinkFetcher = make(map[int64]*fetcher.Fetcher)
	dpllClockIDFetcher = make(map[string]*fetcher.Fetcher)
}

func buildPostProcessDPLLNetlink(clockID int64) fetcher.PostProcessFuncType {
	return func(result map[string]string) (map[string]any, error) {
		processedResult := make(map[string]any)

		entries := make([]NetlinkEntry, 0)
		cleaned := strings.ReplaceAll(result["dpll-netlink"], "'", "\"")
		err := json.Unmarshal([]byte(cleaned), &entries)
		if err != nil {
			log.Errorf("Failed to unmarshal netlink output: %s", err.Error())
		}

		log.Debug("entries: ", entries)
		for _, entry := range entries {
			if entry.ClockID == clockID {
				state, ok := states[entry.LockStatus]
				if !ok {
					log.Errorf("Unknown state: %s", state)
					state = "-1"
				}
				processedResult[entry.ClockType] = state
			}
		}
		return processedResult, nil
	}
}

// BuildDPLLNetlinkInfoFetcher popluates the fetcher required for
// collecting the DPLLInfo
func BuildDPLLNetlinkInfoFetcher(clockID int64) error { //nolint:dupl // Further dedup risks be too abstract or fragile
	fetcherInst, err := fetcher.FetcherFactory(
		[]*clients.Cmd{dateCmd},
		[]fetcher.AddCommandArgs{
			{
				Key:     "dpll-netlink",
				Command: "/linux/tools/net/ynl/cli.py --spec /linux/Documentation/netlink/specs/dpll.yaml --dump device-get",
				Trim:    true,
			},
		},
	)
	if err != nil {
		log.Errorf("failed to create fetcher for dpll netlink: %s", err.Error())
		return fmt.Errorf("failed to create fetcher for dpll netlink: %w", err)
	}
	dpllNetlinkFetcher[clockID] = fetcherInst
	fetcherInst.SetPostProcessor(buildPostProcessDPLLNetlink(clockID))
	return nil
}

// GetDevDPLLInfo returns the device DPLL info for an interface.
func GetDevDPLLNetlinkInfo(ctx clients.ExecContext, clockID int64) (DevNetlinkDPLLInfo, error) {
	dpllInfo := DevNetlinkDPLLInfo{}
	fetcherInst, fetchedInstanceOk := dpllNetlinkFetcher[clockID]
	if !fetchedInstanceOk {
		err := BuildDPLLNetlinkInfoFetcher(clockID)
		if err != nil {
			return dpllInfo, err
		}
		fetcherInst, fetchedInstanceOk = dpllNetlinkFetcher[clockID]
		if !fetchedInstanceOk {
			return dpllInfo, errors.New("failed to create fetcher for DPLLInfo using netlink interface")
		}
	}
	err := fetcherInst.Fetch(ctx, &dpllInfo)
	if err != nil {
		log.Debugf("failed to fetch dpllInfo  via netlink: %s", err.Error())
		return dpllInfo, fmt.Errorf("failed to fetch dpllInfo via netlink: %w", err)
	}
	return dpllInfo, nil
}

func BuildClockIDFetcher(interfaceName string) error {
	fetcherInst, err := fetcher.FetcherFactory(
		[]*clients.Cmd{dateCmd},
		[]fetcher.AddCommandArgs{
			{
				Key: "dpll-netlink-clock-id",
				Command: fmt.Sprintf(
					`export IFNAME=%s; export BUSID=$(readlink /sys/class/net/$IFNAME/device | xargs basename | cut -d ':' -f 2,3);`+
						` echo $(("16#$(lspci -v | grep $BUSID -A 20 |grep 'Serial Number' | awk '{print $NF}' | tr -d '-')"))`,
					interfaceName,
				),
				Trim: true,
			},
		},
	)
	if err != nil {
		log.Errorf("failed to create fetcher for dpll clock ID: %s", err.Error())
		return fmt.Errorf("failed to create fetcher for dpll clock ID: %w", err)
	}
	fetcherInst.SetPostProcessor(postProcessDPLLNetlinkClockID)
	dpllClockIDFetcher[interfaceName] = fetcherInst
	return nil
}

func postProcessDPLLNetlinkClockID(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any)
	clockID, err := strconv.ParseInt(result["dpll-netlink-clock-id"], 10, 64)
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse int for clock id: %w", err)
	}
	processedResult["clockID"] = clockID
	return processedResult, nil
}

type NetlinkClockID struct {
	Timestamp string `fetcherKey:"date"          json:"timestamp"`
	ClockID   int64  `fetcherKey:"clockID"       json:"clockId"`
}

func GetClockID(ctx clients.ExecContext, interfaceName string) (NetlinkClockID, error) {
	clockID := NetlinkClockID{}
	fetcherInst, fetchedInstanceOk := dpllClockIDFetcher[interfaceName]
	if !fetchedInstanceOk {
		err := BuildClockIDFetcher(interfaceName)
		if err != nil {
			return clockID, err
		}
		fetcherInst, fetchedInstanceOk = dpllClockIDFetcher[interfaceName]
		if !fetchedInstanceOk {
			return clockID, errors.New("failed to create fetcher for DPLLInfo using netlink interface")
		}
	}
	err := fetcherInst.Fetch(ctx, &clockID)
	if err != nil {
		log.Debugf("failed to fetch netlink clockID %s", err.Error())
		return clockID, fmt.Errorf("failed to fetch netlink clockID %w", err)
	}
	return clockID, nil
}
