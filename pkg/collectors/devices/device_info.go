// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type PTPDeviceInfo struct {
	Timestamp  string `json:"date" fetcherKey:"date"`
	VendorID   string `json:"vendorId" fetcherKey:"vendorID"`
	DeviceID   string `json:"deviceInfo" fetcherKey:"devID"`
	GNSSDev    string `json:"GNSSDev" fetcherKey:"gnss"`
	Timeoffset string `json:"timeOffset" fetcherKey:"timeOffset"`
}

// AnalyserJSON returns the json expected by the analysers
func (ptpDevInfo *PTPDeviceInfo) GetAnalyserFormat() (*callbacks.AnalyserFormatType, error) {
	offset, err := time.ParseDuration(ptpDevInfo.Timeoffset)
	if err != nil {
		return &callbacks.AnalyserFormatType{}, fmt.Errorf("failed to parse offset %s %w", ptpDevInfo.Timeoffset, err)
	}

	formatted := callbacks.AnalyserFormatType{
		ID: "devInfo",
		Data: []string{
			time.Now().Add(offset).UTC().Format(time.RFC3339Nano),
			ptpDevInfo.Timestamp,
			ptpDevInfo.VendorID,
			ptpDevInfo.DeviceID,
			ptpDevInfo.GNSSDev,
		},
	}
	return &formatted, nil
}

var (
	devFetcher map[string]*fetcher
	devDateCmd *clients.Cmd
)

func init() {
	devFetcher = make(map[string]*fetcher)
	devDateCmdInst, err := clients.NewCmd("date", "date +%s.%N")
	if err != nil {
		panic(err)
	}

	devDateCmd = devDateCmdInst
	devDateCmd.SetOutputProcessor(TrimSpace)
}

func extractOffsetFromTimestamp(result map[string]string) (map[string]string, error) {
	timestamp, err := utils.ParseTimestamp(result["date"])
	if err != nil {
		return result, fmt.Errorf("failed to parse timestamp  %w", err)
	}
	result["date"] = timestamp.Format(time.RFC3339Nano)
	result["timeOffset"] = fmt.Sprintf("%fs", time.Since(timestamp).Seconds())
	return result, nil
}

// BuildPTPDeviceInfo popluates the fetcher required for
// collecting the PTPDeviceInfo
func BuildPTPDeviceInfo(interfaceName string) error {
	fetcherInst := NewFetcher()
	devFetcher[interfaceName] = fetcherInst
	fetcherInst.SetPostProcesser(extractOffsetFromTimestamp)
	fetcherInst.AddCommand(devDateCmd)

	err := fetcherInst.AddNewCommand(
		"gnss",
		fmt.Sprintf("ls /sys/class/net/%s/device/gnss/", interfaceName),
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "gnss", err.Error())
		return fmt.Errorf("failed to fetch devInfo %w", err)
	}

	err = fetcherInst.AddNewCommand(
		"devID",
		fmt.Sprintf("cat /sys/class/net/%s/device/device", interfaceName),
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "devId", err.Error())
		return fmt.Errorf("failed to fetch devInfo %w", err)
	}
	err = fetcherInst.AddNewCommand("vendorID",
		fmt.Sprintf("cat /sys/class/net/%s/device/vendor", interfaceName),
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "vendorID", err.Error())
		return fmt.Errorf("failed to fetch devInfo %w", err)
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
		log.Errorf("failed to fetch devInfo %s", err.Error())
		return devInfo, fmt.Errorf("failed to fetch devInfo %w", err)
	}
	devInfo.GNSSDev = "/dev/" + devInfo.GNSSDev
	return devInfo, nil
}
