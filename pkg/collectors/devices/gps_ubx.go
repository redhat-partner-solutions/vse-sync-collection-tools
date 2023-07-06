// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/fetcher"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

type GPSNav struct {
	TimestampStatus string `json:"timestampStatus" fetcherKey:"navStatusTimestamp"`
	TimestampClock  string `json:"timestampClock" fetcherKey:"navClockTimestamp"`
	GPSFix          string `json:"GPSFix" fetcherKey:"gpsFix"`
	TimeAcc         int    `json:"timeAcc" fetcherKey:"timeAcc"`
	FreqAcc         int    `json:"freqAcc" fetcherKey:"freqAcc"`
}

func (gpsNav *GPSNav) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID: "gnss/time-error",
		Data: []any{
			gpsNav.TimestampClock,
			gpsNav.GPSFix,
			gpsNav.TimeAcc,
			gpsNav.FreqAcc,
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

var (
	timeStampPattern = `(\d+.\d+)`
	ubxNavRegex      = regexp.MustCompile(
		timeStampPattern +
			`\nUBX-NAV-STATUS:\n\s+iTOW (\d+) gpsFix (\d) flags (.*) fixStat ` +
			`(.*) flags2\s(.*)\n\s+ttff\s(\d+), msss (\d+)\n\n` +
			timeStampPattern +
			`\nUBX-NAV-CLOCK:\n\s+iTOW (\d+) clkB (\d+) clkD (\d+) tAcc (\d+) fAcc (\d+)`,
		// ubxtool output example:
		// 1686916187.0584
		// UBX-NAV-STATUS:
		//   iTOW 474605000 gpsFix 3 flags 0xdd fixStat 0x0 flags2 0x8
		//   ttff 25030, msss 4294967295
		//
		// 1686916187.0586
		// UBX-NAV-CLOCK:
		//   iTOW 474605000 clkB 61594 clkD 56 tAcc 5 fAcc 164
		//
	)
	gpsFetcher *fetcher.Fetcher
)

func init() {
	gpsFetcher = fetcher.NewFetcher()
	gpsFetcher.SetPostProcesser(processUBXNav)
	err := gpsFetcher.AddNewCommand(
		"GPS",
		"ubxtool -t -p NAV-STATUS -p NAV-CLOCK -P 29.20",
		true,
	)
	if err != nil {
		panic(fmt.Errorf("failed to setup GPS fetcher %w", err))
	}
}

// processUBXNav parses the output of the ubxtool extracting the required values for gpsNav
func processUBXNav(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any)
	match := ubxNavRegex.FindStringSubmatch(result["GPS"])
	if len(match) == 0 {
		return processedResult, fmt.Errorf(
			"unable to parse UBX Nav Status or Clock from %s",
			result["GPS"],
		)
	}
	timestampSatus, err := utils.ParseTimestamp(match[1])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse navStatusTimestamp %w", err)
	}
	processedResult["navStatusTimestamp"] = timestampSatus.Format(time.RFC3339Nano)

	timestampClock, err := utils.ParseTimestamp(match[9])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse navClockTimestamp %w", err)
	}
	timeAcc, err := strconv.Atoi(match[13])
	if err != nil {
		return processedResult, fmt.Errorf("failed to convert %s into and int", match[13])
	}
	freqAcc, err := strconv.Atoi(match[14])
	if err != nil {
		return processedResult, fmt.Errorf("failed to convert %s into and int", match[14])
	}

	processedResult["navClockTimestamp"] = timestampClock.Format(time.RFC3339Nano)
	processedResult["gpsFix"] = match[3]
	processedResult["timeAcc"] = timeAcc
	processedResult["freqAcc"] = freqAcc

	return processedResult, nil
}

// GetGPSNav returns GPSNav of the host
func GetGPSNav(ctx clients.ContainerContext) (GPSNav, error) {
	gpsNav := GPSNav{}
	err := gpsFetcher.Fetch(ctx, &gpsNav)
	if err != nil {
		log.Debugf("failed to fetch gpsNav %s", err.Error())
		return gpsNav, fmt.Errorf("failed to fetch gpsNav %w", err)
	}
	return gpsNav, nil
}
