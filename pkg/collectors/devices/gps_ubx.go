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
	TimestampMon    string `json:"timestampMon" fetcherKey:"monTimestamp"`
	AntBlockID      []int  `json:"antBlockId" fetcherKey:"antBlockId"`
	AntStatus       []int  `json:"antStatus" fetcherKey:"antStatus"`
	AntPower        []int  `json:"antPower" fetcherKey:"antPower"`
	TimeAcc         int    `json:"timeAcc" fetcherKey:"timeAcc"`
	FreqAcc         int    `json:"freqAcc" fetcherKey:"freqAcc"`
}

func (gpsNav *GPSNav) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	messages := []*callbacks.AnalyserFormatType{}
	messages = append(messages, &callbacks.AnalyserFormatType{
		ID: "gnss/time-error",
		Data: []any{
			gpsNav.TimestampClock,
			gpsNav.GPSFix,
			gpsNav.TimeAcc,
			gpsNav.FreqAcc,
		},
	})

	for blockIndex, blockID := range gpsNav.AntBlockID {
		messages = append(messages, &callbacks.AnalyserFormatType{
			ID: "gnss/rf-mon",
			Data: []any{
				gpsNav.TimestampMon,
				blockID,
				gpsNav.AntStatus[blockIndex],
				gpsNav.AntPower[blockIndex],
			},
		})
	}
	return messages, nil
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
	ubxAntFullBlockRegex = regexp.MustCompile(
		timeStampPattern +
			`\nUBX-MON-RF:\n` +
			`\s+version \d nBlocks (\d) reserved1 \d \d\n(?s:([^UBX-]*)[UBX-]?)`,
		// 1686916187.0584
		// UBX-MON-RF:
		//  version 0 nBlocks 2 reserved1 0 0
		//    blockId 0 flags x0 antStatus 2 antPower 1 postStatus 0 reserved2 0 0 0 0
		//     noisePerMS 82 agcCnt 6318 jamInd 3 ofsI 15 magI 154 ofsQ 2 magQ 145
		//     reserved3 0 0 0
		//    blockId 1 flags x0 antStatus 2 antPower 1 postStatus 0 reserved2 0 0 0 0
		//     noisePerMS 49 agcCnt 6669 jamInd 2 ofsI 11 magI 146 ofsQ 1 magQ 139
		//     reserved3 0 0 0
	)
	ubxAntInternalBlockRegex = regexp.MustCompile(
		`\s+blockId (\d) flags \w+ antStatus (\d) antPower (\d+) postStatus \d reserved2 \d \d \d \d\n` +
			`\s+noisePerMS \d+ agcCnt \d+ jamInd \d ofsI \d+ magI \d+ ofsQ \d magQ \d+\n` +
			`\s+reserved3 \d \d \d\n`,
		//    blockId 0 flags x0 antStatus 2 antPower 1 postStatus 0 reserved2 0 0 0 0
		//     noisePerMS 82 agcCnt 6318 jamInd 3 ofsI 15 magI 154 ofsQ 2 magQ 145
		//     reserved3 0 0 0
		//    blockId 1 flags x0 antStatus 2 antPower 1 postStatus 0 reserved2 0 0 0 0
		//     noisePerMS 49 agcCnt 6669 jamInd 2 ofsI 11 magI 146 ofsQ 1 magQ 139
		//     reserved3 0 0 0
	)
	gpsFetcher *fetcher.Fetcher
)

func init() {
	gpsFetcher = fetcher.NewFetcher()
	gpsFetcher.SetPostProcesser(processUBX)
	err := gpsFetcher.AddNewCommand(
		"GPS",
		"ubxtool -t -p NAV-STATUS -p NAV-CLOCK -p MON-RF -P 29.20",
		true,
	)
	if err != nil {
		panic(fmt.Errorf("failed to setup GPS fetcher %w", err))
	}
}

// processUBXNav parses the output of the ubxtool extracting the required values for gpsNav
func processUBXNav(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any)
	matchNav := ubxNavRegex.FindStringSubmatch(result["GPS"])
	if len(matchNav) == 0 {
		return processedResult, fmt.Errorf(
			"unable to parse UBX Nav Status or Clock from %s",
			result["GPS"],
		)
	}
	timestampSatus, err := utils.ParseTimestamp(matchNav[1])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse navStatusTimestamp %w", err)
	}
	processedResult["navStatusTimestamp"] = timestampSatus.Format(time.RFC3339Nano)

	timestampClock, err := utils.ParseTimestamp(matchNav[9])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse navClockTimestamp %w", err)
	}
	timeAcc, err := strconv.Atoi(matchNav[13])
	if err != nil {
		return processedResult, fmt.Errorf("failed to convert %s into and int", matchNav[13])
	}
	freqAcc, err := strconv.Atoi(matchNav[14])
	if err != nil {
		return processedResult, fmt.Errorf("failed to convert %s into and int", matchNav[14])
	}

	processedResult["navClockTimestamp"] = timestampClock.Format(time.RFC3339Nano)
	processedResult["gpsFix"] = matchNav[3]
	processedResult["timeAcc"] = timeAcc
	processedResult["freqAcc"] = freqAcc
	return processedResult, nil
}

func processUBXMon(result map[string]string) (map[string]any, error) { //nolint:funlen // allow for a slightly long function
	processedResult := make(map[string]any)

	antFullMatch := ubxAntFullBlockRegex.FindStringSubmatch(result["GPS"])
	if len(antFullMatch) == 0 {
		return processedResult, fmt.Errorf("failed to match UBX MON in %s", result["GPS"])
	}

	timestampMon, err := utils.ParseTimestamp(antFullMatch[1])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse monTimestamp %w", err)
	}
	processedResult["monTimestamp"] = timestampMon.Format(time.RFC3339Nano)

	nBlocks, err := strconv.Atoi(antFullMatch[2])
	if err != nil {
		return processedResult, fmt.Errorf("failed to parse UBX antenna monitoring %w", err)
	}

	antBlockMatches := ubxAntInternalBlockRegex.FindAllStringSubmatch(antFullMatch[3], nBlocks)
	antBlockID := make([]int, 0)
	antStatus := make([]int, 0)
	antPower := make([]int, 0)
	for _, antBlock := range antBlockMatches {
		antBlockIDValue, err := strconv.Atoi(antBlock[1])
		if err != nil {
			return processedResult, fmt.Errorf("failed to convert %s to an int for antBlockIDValue %w", antBlock[1], err)
		}
		antBlockID = append(antBlockID, antBlockIDValue)

		antStatusValue, err := strconv.Atoi(antBlock[2])
		if err != nil {
			return processedResult, fmt.Errorf("failed to convert %s to an int for antStatusValue %w", antBlock[2], err)
		}
		antStatus = append(antStatus, antStatusValue)

		antPowerValue, err := strconv.Atoi(antBlock[3])
		if err != nil {
			return processedResult, fmt.Errorf("failed to convert %s to an int for antPowerValue %w", antBlock[3], err)
		}
		antPower = append(antPower, antPowerValue)
	}

	processedResult["antBlockId"] = antBlockID
	processedResult["antStatus"] = antStatus
	processedResult["antPower"] = antPower
	return processedResult, nil
}

func processUBX(result map[string]string) (map[string]any, error) {
	processedResult := make(map[string]any)
	processedUBXNav, err := processUBXNav(result)
	if err != nil {
		log.Errorf("processUBXNav Failed: %s", err.Error())
		return processedResult, err
	}
	for key, value := range processedUBXNav {
		processedResult[key] = value
	}

	processedUBXMon, err := processUBXMon(result)
	if err != nil {
		log.Errorf("processUBXMon Failed: %s", err.Error())
		return processedResult, err
	}
	for key, value := range processedUBXMon {
		processedResult[key] = value
	}

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
