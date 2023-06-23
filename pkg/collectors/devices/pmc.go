// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type PMCGrandMaster struct {
	Timestamp  string `json:"timestamp" fetcherKey:"date"`
	ClockID    string `json:"clockId" fetcherKey:"clockID"`
	ClockClass string `json:"clockClass" fetcherKey:"clockClass"`
	Leap61     string `json:"leap61" fetcherKey:"leap61"`
	Leap59     string `json:"leap59" fetcherKey:"leap59"`
}

func (pmcGM *PMCGrandMaster) GetAnalyserFormat() (*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID: "ptpGmAnouncement",
		Data: []string{
			pmcGM.Timestamp,
			pmcGM.ClockID,
			pmcGM.ClockClass,
			pmcGM.Leap61,
			pmcGM.Leap59,
		},
	}
	return &formatted, nil
}

var (
	pmcGMRegex = regexp.MustCompile(
		`sending: GET GRANDMASTER_SETTINGS_NP` +
			`\s*([[:xdigit:]]{6}\.[[:xdigit:]]{4}\.[[:xdigit:]]{6}-\d+) seq (\d+) RESPONSE MANAGEMENT GRANDMASTER_SETTINGS_NP` +
			`\s*clockClass\s+(\d+)` +
			`\s*clockAccuracy\s+((?:0x[[:xdigit:]]+)|(?:\d*))` +
			`\s*offsetScaledLogVariance\s+((?:0x[[:xdigit:]]+)|(?:\d*))` +
			`\s*currentUtcOffset\s+(\d+)` +
			`\s*leap61\s+(\d+)` +
			`\s*leap59\s+(\d+)` +
			`\s*currentUtcOffsetValid\s+(\d+)` +
			`\s*ptpTimescale\s+(\d+)` +
			`\s*timeTraceable\s+(\d+)` +
			`\s*frequencyTraceable\s+(\d+)` +
			`\s*timeSource\s+((?:0x[[:xdigit:]]+)|(?:\d*))`,
	// PMC output example:
	// sending: GET GRANDMASTER_SETTINGS_NP
	// 507c6f.fffe.1fb54d-0 seq 0 RESPONSE MANAGEMENT GRANDMASTER_SETTINGS_NP
	// 		clockClass              248
	// 		clockAccuracy           0xfe
	// 		offsetScaledLogVariance 0xffff
	// 		currentUtcOffset        37
	// 		leap61                  0
	// 		leap59                  0
	// 		currentUtcOffsetValid   0
	// 		ptpTimescale            1
	// 		timeTraceable           0
	// 		frequencyTraceable      0
	// 		timeSource              0xa0
	)
	pmcFetcher *fetcher
)

func init() {
	pmcFetcher = NewFetcher()
	pmcFetcher.SetPostProcesser(processPMCGrandMaster)

	pmcDateCmdInst, err := clients.NewCmd("date", "date +%s.%N")
	if err != nil {
		panic(err)
	}
	pmcDateCmdInst.SetOutputProcessor(formatTimestampAsRFC3339Nano)
	pmcFetcher.AddCommand(pmcDateCmdInst)

	err = pmcFetcher.AddNewCommand(
		"PMC-GM",
		"pmc -u -f /var/run/ptp4l.0.config 'GET GRANDMASTER_SETTINGS_NP'",
		true,
	)
	if err != nil {
		log.Errorf("failed to add command %s %s", "PMC-GM", err.Error())
		panic(fmt.Errorf("failed to setup PMC-GM fetcher %w", err))
	}
}

// processPMCGrandMaster parses the output of the pmc extracting the required values for PMCGrandMaster
func processPMCGrandMaster(result map[string]string) (map[string]string, error) {
	processedResult := make(map[string]string)
	match := pmcGMRegex.FindStringSubmatch(result["PMC-GM"])
	if len(match) == 0 {
		return processedResult, fmt.Errorf(
			"unable to parse  from %s",
			result["PMC-GM"],
		)
	}
	processedResult["date"] = result["date"]
	processedResult["clockID"] = match[1]
	processedResult["clockClass"] = match[3]
	processedResult["leap61"] = match[7]
	processedResult["leap59"] = match[8]
	return processedResult, nil
}

// GetPMCGrandMaster returns PMCGrandMaster of the host
func GetPMCGrandMaster(ctx clients.ContainerContext) (PMCGrandMaster, error) {
	pmcGM := PMCGrandMaster{}
	err := pmcFetcher.Fetch(ctx, &pmcGM)
	if err != nil {
		log.Errorf("failed to fetch pmcGM %s", err.Error())
		return pmcGM, err
	}
	return pmcGM, nil
}
