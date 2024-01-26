// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/fetcher"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type GPSVersions struct {
	Timestamp       string   `fetcherKey:"timestamp"       json:"timestamp"`
	FirmwareVersion string   `fetcherKey:"firmwareVersion" json:"firmwareVersion"`
	ProtoVersion    string   `fetcherKey:"protocolVersion" json:"protocolVersion"`
	Module          string   `fetcherKey:"module"          json:"module"`
	UBXVersion      string   `fetcherKey:"UBXVersion"      json:"ubxVersion"`
	GPSDVersion     string   `fetcherKey:"GPSDVersion"     json:"gpsdVersion"`
	GNSSDevices     []string `fetcherKey:"GNSSDevices"     json:"gnssDevices"`
}

func (gpsVer *GPSVersions) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	messages := []*callbacks.AnalyserFormatType{
		{
			ID:   "gnss/time-error",
			Data: gpsVer,
		},
	}
	return messages, nil
}

var (
	ubxFirmwareVersion = regexp.MustCompile(
		timeStampPattern +
			`\nUBX-MON-VER:` +
			`\n\s+swVersion (.*)` +
			`\n\s+hwVersion (.*)` +
			`\n\s+((?:extension .*(?:\n\s+)?)+)`,
		// 1689260332.4728
		// UBX-MON-VER:
		//   swVersion EXT CORE 1.00 (3fda8e)
		//   hwVersion 00190000
		//   extension ROM BASE 0x118B2060
		//   extension FWVER=TIM 2.20
		//   extension PROTVER=29.20
		//   extension MOD=ZED-F9T
		//   extension GPS;GLO;GAL;BDS
		//   extension SBAS;QZSS
		//   extension NAVIC
	)
	fwVersionExtension    = regexp.MustCompile(`extension FWVER=(.*)`)
	protoVersionExtension = regexp.MustCompile(`extension PROTVER=(.*)`)
	moduleExtension       = regexp.MustCompile(`extension MOD=(.*)`)

	ubxVersion = regexp.MustCompile("ubxtool: Version (.*)")
	// ubxtool: Version 3.25.1~dev
	gpsdVersion = regexp.MustCompile(`gpsd: (.* \(revision .*\))`)
	// gpsd: 3.25.1~dev (revision release-3.25-109-g1a04cfab8)
	gpsVerFetcher *fetcher.Fetcher
)

func init() {
	gpsVerFetcherInst, err := fetcher.FetcherFactory(
		[]*clients.Cmd{},
		[]fetcher.AddCommandArgs{
			{
				Key:     "UBXMonVer",
				Command: "ubxtool -t -p MON-VER -P 29.20",
				Trim:    true,
			},
			{
				Key:     "UBXVersion",
				Command: "ubxtool -V",
				Trim:    true,
			},
			{
				Key:     "GPSDVersion",
				Command: "gpsd --version",
				Trim:    true,
			},
			{
				Key:     "GNSSDevices",
				Command: "ls -1 /dev | grep gnss", // using grep so we just get an empty string if there is nothing
				Trim:    true,
			},
		},
	)

	if err != nil {
		panic(fmt.Errorf("failed to setup GPSD Version fetcher %w", err))
	}
	gpsVerFetcherInst.SetPostProcessor(processGPSVer)
	gpsVerFetcher = gpsVerFetcherInst
}

// findFirstCaptureGroup returns the first capture group of supplied regex.
func findFirstCaptureGroup(input string, regex *regexp.Regexp, name string) (string, error) {
	version := regex.FindStringSubmatch(input)
	if len(version) == 0 {
		return "", fmt.Errorf(
			"unable to parse version from %s in %s",
			name, input,
		)
	}
	return version[1], nil
}

func processExtentions(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any)
	match := ubxFirmwareVersion.FindStringSubmatch(result["UBXMonVer"])
	if len(match) == 0 {
		return processedResult, fmt.Errorf(
			"unable to parse UBX MON Version from %s",
			result["UBXMonVer"],
		)
	}

	timestamp, err := utils.ParseTimestamp(match[1])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse versionTimestamp %w", err)
	}
	processedResult["timestamp"] = timestamp.Format(time.RFC3339Nano)

	fwVer, err := findFirstCaptureGroup(match[4], fwVersionExtension, "extension")
	if err != nil {
		return processedResult, err
	}
	processedResult["firmwareVersion"] = fwVer

	protoVer, err := findFirstCaptureGroup(match[4], protoVersionExtension, "extension")
	if err != nil {
		return processedResult, err
	}
	processedResult["protocolVersion"] = protoVer

	module, err := findFirstCaptureGroup(match[4], moduleExtension, "extension")
	if err != nil {
		return processedResult, err
	}
	processedResult["module"] = module
	return processedResult, nil
}

func processGPSVer(result map[string]string) (map[string]any, error) {
	processedResult, err := processExtentions(result)
	if err != nil {
		return processedResult, err
	}

	ubxVer, err := findFirstCaptureGroup(result["UBXVersion"], ubxVersion, "ubxtools version")
	if err != nil {
		return processedResult, err
	}
	processedResult["UBXVersion"] = ubxVer

	gpsdVer, err := findFirstCaptureGroup(result["GPSDVersion"], gpsdVersion, "gpsd version")
	if err != nil {
		return processedResult, err
	}
	processedResult["GPSDVersion"] = gpsdVer

	gnssDevices := make([]string, 0)
	for _, dev := range strings.Split(result["GNSSDevices"], "\n") {
		dev = strings.TrimSpace(dev)
		if len(dev) > 0 {
			gnssDevices = append(gnssDevices, "/dev/"+dev)
		}
	}
	processedResult["GNSSDevices"] = gnssDevices
	return processedResult, nil
}

// GetGPSVersions returns GPSVersions of the host
func GetGPSVersions(ctx clients.ExecContext) (GPSVersions, error) {
	gpsVer := GPSVersions{}
	err := gpsVerFetcher.Fetch(ctx, &gpsVer)
	if err != nil {
		log.Debugf("failed to fetch gpsVer %s", err.Error())
		return gpsVer, fmt.Errorf("failed to fetch gpsVer %w", err)
	}
	return gpsVer, nil
}
