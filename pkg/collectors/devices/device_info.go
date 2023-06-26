// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/fetcher"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type PTPDeviceInfo struct {
	Timestamp       string        `json:"date" fetcherKey:"date"`
	VendorID        string        `json:"vendorId" fetcherKey:"vendorID"`
	DeviceID        string        `json:"deviceInfo" fetcherKey:"devID"`
	GNSSDev         string        `json:"GNSSDev" fetcherKey:"gnss"`
	Timeoffset      time.Duration `json:"timeOffset" fetcherKey:"timeOffset"`
	FirmwareVersion string        `json:"firmwareVersion" fetcherKey:"firmwareVersion"`
}

// AnalyserJSON returns the json expected by the analysers
func (ptpDevInfo *PTPDeviceInfo) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID: "devInfo",
		Data: []any{
			time.Now().Add(ptpDevInfo.Timeoffset).UTC().Format(time.RFC3339Nano),
			ptpDevInfo.Timestamp,
			ptpDevInfo.VendorID,
			ptpDevInfo.DeviceID,
			ptpDevInfo.GNSSDev,
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

var (
	devFetcher map[string]*fetcher.Fetcher
	devDateCmd *clients.Cmd
)

func init() {
	devFetcher = make(map[string]*fetcher.Fetcher)
	devDateCmdInst, err := clients.NewCmd("date", "date +%s.%N")
	if err != nil {
		panic(err)
	}

	devDateCmd = devDateCmdInst
	devDateCmd.SetOutputProcessor(fetcher.TrimSpace)
}

func extractOffsetFromTimestamp(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any, 0)
	timestamp, err := utils.ParseTimestamp(result["date"])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse timestamp  %w", err)
	}
	processedResult["date"] = timestamp.Format(time.RFC3339Nano)
	processedResult["timeOffset"] = time.Since(timestamp)
	return processedResult, nil
}

func processGNSSPath(s string) (string, error) {
	return "/dev/" + strings.TrimSpace(s), nil
}

func failedToAddError(cmdKey string, err error) error {
	log.Errorf("failed to add command %s %s", cmdKey, err.Error())
	return fmt.Errorf("failed to fetch devInfo %w", err)
}

// BuildPTPDeviceInfo popluates the fetcher required for
// collecting the PTPDeviceInfo
func BuildPTPDeviceInfo(interfaceName string) error {
	fetcherInst := fetcher.NewFetcher()
	devFetcher[interfaceName] = fetcherInst
	fetcherInst.SetPostProcesser(extractOffsetFromTimestamp)
	fetcherInst.AddCommand(devDateCmd)

	gnssCmd, err := clients.NewCmd("gnss", fmt.Sprintf("ls /sys/class/net/%s/device/gnss/", interfaceName))
	if err != nil {
		return failedToAddError("gnss", err)
	}
	gnssCmd.SetOutputProcessor(processGNSSPath)
	fetcherInst.AddCommand(gnssCmd)

	err = fetcherInst.AddNewCommand(
		"devID",
		fmt.Sprintf("cat /sys/class/net/%s/device/device", interfaceName),
		true,
	)
	if err != nil {
		return failedToAddError("devID", err)
	}

	err = fetcherInst.AddNewCommand("vendorID",
		fmt.Sprintf("cat /sys/class/net/%s/device/vendor", interfaceName),
		true,
	)
	if err != nil {
		return failedToAddError("vendorID", err)
	}

	err = fetcherInst.AddNewCommand("firmwareVersion",
		fmt.Sprintf("ethtool -i %s | grep firmware --color=never", interfaceName),
		true,
	)
	if err != nil {
		return failedToAddError("firmwareVersion", err)
	}
	return nil
}

// GetPTPDeviceInfo returns the PTPDeviceInfo for an interface
func GetPTPDeviceInfo(interfaceName string, ctx clients.ContainerContext) (PTPDeviceInfo, error) {
	devInfo := PTPDeviceInfo{}
	// Find the dev for the GNSS for this interface
	fetcherInst, fetchedInstanceOk := devFetcher[interfaceName]
	if !fetchedInstanceOk {
		err := BuildPTPDeviceInfo(interfaceName)
		if err != nil {
			return devInfo, err
		}
		fetcherInst, fetchedInstanceOk = devFetcher[interfaceName]
		if !fetchedInstanceOk {
			return devInfo, errors.New("failed to create fetcher for PTPDeviceInfo")
		}
	}

	err := fetcherInst.Fetch(ctx, &devInfo)
	if err != nil {
		log.Debugf("failed to fetch devInfo %s", err.Error())
		return devInfo, fmt.Errorf("failed to fetch devInfo %w", err)
	}
	return devInfo, nil
}
