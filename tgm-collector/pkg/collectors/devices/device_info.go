// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/fetcher"
)

type PTPDeviceInfo struct {
	Timestamp       string        `fetcherKey:"date"            json:"date"`
	VendorID        string        `fetcherKey:"vendorID"        json:"vendorId"`
	DeviceID        string        `fetcherKey:"devID"           json:"deviceInfo"`
	GNSSDev         string        `fetcherKey:"gnss"            json:"GNSSDev"`
	FirmwareVersion string        `fetcherKey:"firmwareVersion" json:"firmwareVersion"`
	DriverVersion   string        `fetcherKey:"driverVersion"   json:"driverVersion"`
	Timeoffset      time.Duration `fetcherKey:"timeOffset"      json:"timeOffset"`
}

// AnalyserJSON returns the json expected by the analysers
func (ptpDevInfo *PTPDeviceInfo) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID: "devInfo",
		Data: map[string]any{
			"timestamp":         time.Now().Add(ptpDevInfo.Timeoffset).UTC().Format(time.RFC3339Nano),
			"fetched_timestamp": ptpDevInfo.Timestamp,
			"vendorID":          ptpDevInfo.VendorID,
			"devID":             ptpDevInfo.DeviceID,
			"gnss":              ptpDevInfo.GNSSDev,
			"firmwareVersion":   ptpDevInfo.FirmwareVersion,
			"driverVersion":     ptpDevInfo.DriverVersion,
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

var (
	devFetcher   map[string]*fetcher.Fetcher
	ethtoolRegex = regexp.MustCompile(`version: (.*)\nfirmware-version: (.*)\n`)
	// driver: ice
	// version: 1.11.20.7
	// firmware-version: 4.20 0x8001778b 1.3346.0
	// expansion-rom-version:
	// bus-info: 0000:86:00.0
	// supports-statistics: yes
	// supports-test: yes
	// supports-eeprom-access: yes
	// supports-register-dump: yes
	// supports-priv-flags: yes
)

func init() {
	devFetcher = make(map[string]*fetcher.Fetcher)
}

func extractOffsetFromTimestamp(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any, 0)
	timestamp, err := time.Parse(time.RFC3339Nano, result["date"])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse timestamp  %w", err)
	}
	processedResult["timeOffset"] = time.Since(timestamp)
	return processedResult, nil
}

func extractEthtoolsInfo(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any, 0)
	match := ethtoolRegex.FindStringSubmatch(result["ethtoolOut"])
	if len(match) == 0 {
		return processedResult, fmt.Errorf(
			"failed to extract ethtoolOut from %s",
			result["ethtoolOut"],
		)
	}
	processedResult["driverVersion"] = match[1]
	processedResult["firmwareVersion"] = match[2]
	return processedResult, nil
}

func devInfoPostProcessor(result map[string]string) (map[string]any, error) {
	processedResult, err := extractOffsetFromTimestamp(result)
	if err != nil {
		return processedResult, err
	}
	firmwareResult, err := extractEthtoolsInfo(result)
	if err != nil {
		return processedResult, err
	}

	for key, value := range firmwareResult {
		processedResult[key] = value
	}
	return processedResult, nil
}

func processGNSSPath(s string) (string, error) {
	return "/dev/" + strings.TrimSpace(s), nil
}

// BuildPTPDeviceInfo popluates the fetcher required for
// collecting the PTPDeviceInfo
func BuildPTPDeviceInfo(interfaceName string) error { //nolint:dupl // Further dedup risks be too abstract or fragile
	gnssCmd, err := clients.NewCmd("gnss", fmt.Sprintf("ls /sys/class/net/%s/device/gnss/", interfaceName))
	if err != nil {
		return fmt.Errorf("failed to create fetcher for devInfo: failed to create command for %s: %w", "gnss", err)
	}
	gnssCmd.SetOutputProcessor(processGNSSPath)

	fetcherInst, err := fetcher.FetcherFactory(
		[]*clients.Cmd{
			dateCmd,
			gnssCmd,
		},
		[]fetcher.AddCommandArgs{
			{
				Key:     "devID",
				Command: fmt.Sprintf("cat /sys/class/net/%s/device/device", interfaceName),
				Trim:    true,
			},
			{
				Key:     "vendorID",
				Command: fmt.Sprintf("cat /sys/class/net/%s/device/vendor", interfaceName),
				Trim:    true,
			},
			{
				Key:     "ethtoolOut",
				Command: fmt.Sprintf("ethtool -i %s", interfaceName),
				Trim:    true,
			},
		},
	)
	if err != nil {
		log.Errorf("failed to create fetcher for devInfo: %s", err.Error())
		return fmt.Errorf("failed to create fetcher for devInfo: %w", err)
	}
	devFetcher[interfaceName] = fetcherInst
	fetcherInst.SetPostProcessor(devInfoPostProcessor)
	return nil
}

// GetPTPDeviceInfo returns the PTPDeviceInfo for an interface
func GetPTPDeviceInfo(interfaceName string, ctx clients.ExecContext) (PTPDeviceInfo, error) {
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
