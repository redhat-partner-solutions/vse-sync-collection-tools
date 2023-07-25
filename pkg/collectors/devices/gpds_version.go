// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/fetcher"
)

type GPSDVersion struct {
	Timestamp string `json:"timestamp" fetcherKey:"date"`
	Version   string `json:"version" fetcherKey:"GPSD"`
}

func (gpsd *GPSDVersion) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID: "gpsd-version",
		Data: map[string]any{
			"timestamp": gpsd.Timestamp,
			"version":   gpsd.Version,
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

var (
	gpsdVersionFetcher *fetcher.Fetcher
)

func init() {
	gpsdVersionFetcher = fetcher.NewFetcher()

	gpdsDateCmdInst, err := clients.NewCmd("date", "date +%s.%N")
	if err != nil {
		panic(fmt.Errorf("failed to setup GPS fetcher %w", err))
	}
	gpdsDateCmdInst.SetOutputProcessor(formatTimestampAsRFC3339Nano)
	gpsdVersionFetcher.AddCommand(gpdsDateCmdInst)

	gpdsVersionCmdInst, err := clients.NewCmd("GPSD", "gpsd --version")
	if err != nil {
		panic(fmt.Errorf("failed to setup GPS fetcher %w", err))
	}
	gpdsVersionCmdInst.SetOutputProcessor(processGPSDVersion)
	gpsdVersionFetcher.AddCommand(gpdsVersionCmdInst)
}

// processUBXNav parses the output of the ubxtool extracting the required values for gpsNav
func processGPSDVersion(s string) (string, error) {
	version := strings.TrimSpace(s)
	return version, nil
}

// GetGPSDVersion returns GPSDVersion of the host
func GetGPSDVersion(ctx clients.ContainerContext) (GPSDVersion, error) {
	gpsdVersion := GPSDVersion{}
	err := gpsdVersionFetcher.Fetch(ctx, &gpsdVersion)
	if err != nil {
		log.Debugf("failed to fetch gpsdVersion %s", err.Error())
		return gpsdVersion, fmt.Errorf("failed to fetch gpsdVersion %w", err)
	}
	return gpsdVersion, nil
}
