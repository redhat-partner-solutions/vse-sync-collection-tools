// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type GPSNav struct {
	TimestampStatus string `json:"timestampStatus" fetcherKey:"navStatusTimestamp"`
	TimestampClock  string `json:"timestampClock" fetcherKey:"navClockTimestamp"`
	GPSFix          string `json:"GPSFix" fetcherKey:"gpsFix"`
	TimeAcc         string `json:"timeAcc" fetcherKey:"timeAcc"`
	FreqAcc         string `json:"freqAcc" fetcherKey:"freqAcc"`
}

var (
	timeStampPattern = `\d+(.\d+) (\w{3} \w{3} \d{1,2} \d{1,2}\:\d{1,2}\:\d{1,2}) (\d{4})`
	ubxNavRegex      = regexp.MustCompile(
		timeStampPattern +
			`\nUBX-NAV-STATUS:\n\s+iTOW (\d+) gpsFix (\d) flags (.*) fixStat ` +
			`(.*) flags2\s(.*)\n\s+ttff\s(\d+), msss (\d+)\n\n` +
			timeStampPattern +
			`\nUBX-NAV-CLOCK:\n\s+iTOW (\d+) clkB (\d+) clkD (\d+) tAcc (\d+) fAcc (\d+)`,
	)
	timeForm   = "Mon Jan _2 15:04:05 2006"
	gpsFetcher *fetcher
)

func init() {
	gpsFetcher = NewFetcher()
	gpsFetcher.SetPostProcesser(processUBXNav)
	err := gpsFetcher.AddNewCommand(
		"GPS",
		"ubxtool -tt -p NAV-STATUS -p NAV-CLOCK -P 29.20",
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "GPS", err.Error())
		panic(fmt.Errorf("failed to setup GPS fetcher %w", err))
	}
}

func parseTimestamp(parts ...string) (time.Time, error) {
	timestamp, err := time.Parse(timeForm, strings.Join(parts, " "))
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"unable to parse timestamp from %s, %w,",
			strings.Join(parts, " "),
			err,
		)
	}
	return timestamp, nil
}

func processUBXNav(result map[string]string) (map[string]string, error) {
	processedResult := make(map[string]string)
	match := ubxNavRegex.FindStringSubmatch(result["GPS"])
	if len(match) == 0 {
		return processedResult, fmt.Errorf(
			"unable to parse UBX Nav Status or Clock from %s",
			result["GPS"],
		)
	}

	timestampSatus, err := parseTimestamp(match[2]+match[1], match[3])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse navStatusTimestamp %w", err)
	}
	processedResult["navStatusTimestamp"] = timestampSatus.String()

	timestampClock, err := parseTimestamp(match[12]+match[11], match[13])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse navClockTimestamp %w", err)
	}

	processedResult["navClockTimestamp"] = timestampClock.String()
	processedResult["gpsFix"] = match[5]
	processedResult["timeAcc"] = match[17]
	processedResult["freqAcc"] = match[18]

	return processedResult, nil
}

func GetGPSNav(ctx clients.ContainerContext) (GPSNav, error) {
	gpsNav := GPSNav{}
	err := gpsFetcher.Fetch(ctx, &gpsNav)
	if err != nil {
		log.Errorf("failed to fetch gpsNav %s", err.Error())
		return gpsNav, err
	}
	return gpsNav, nil
}
