// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/fetcher"
)

type PMCInfo struct {
	Timestamp               string `fetcherKey:"date"                    json:"timestamp"`
	TimeSource              string `fetcherKey:"timeSource"              json:"timeSource"`
	ClockAccuracy           string `fetcherKey:"clockAccuracy"           json:"clockAccuracy"`
	OffsetScaledLogVariance string `fetcherKey:"offsetScaledLogVariance" json:"offsetScaledLogVariance"`
	ClockClass              int    `fetcherKey:"clockClass"              json:"clock_class"` //nolint:tagliatelle // needs to match the parser in vse-sync-pp
	CurrentUtcOffset        int    `fetcherKey:"currentUtcOffset"        json:"currentUtcOffset"`
	Leap61                  int    `fetcherKey:"leap61"                  json:"leap61"`
	Leap59                  int    `fetcherKey:"leap59"                  json:"leap59"`
	CurrentUtcOffsetValid   int    `fetcherKey:"currentUtcOffsetValid"   json:"currentUtcOffsetValid"`
	PtpTimescale            int    `fetcherKey:"ptpTimescale"            json:"ptpTimescale"`
	TimeTraceable           int    `fetcherKey:"timeTraceable"           json:"timeTraceable"`
	FrequencyTraceable      int    `fetcherKey:"frequencyTraceable"      json:"frequencyTraceable"`
}

// GetAnalyserFormat returns the json expected by the analysers
func (gmSetting *PMCInfo) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	formatted := callbacks.AnalyserFormatType{
		ID:   "phc/gm-settings",
		Data: gmSetting,
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

// MapStringToInt converts map string to map int
func MapStringToInt(inputMap map[string]string) (map[string]int, error) {
	convertedMap := make(map[string]int)
	for key, value := range inputMap {
		convertedValue, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %s into and int", value)
		}

		convertedMap[key] = convertedValue
	}

	return convertedMap, nil
}

var (
	pmcFetcher *fetcher.Fetcher
	pmcRegEx   = regexp.MustCompile(
		`\sclockClass\s+(\d+)` +
			`\s*clockAccuracy\s+(.+)\n` +
			`\s*offsetScaledLogVariance\s+(.+)\n` +
			`\s*currentUtcOffset\s+(\d+)\n` +
			`\s*leap61\s+(\d+)\n` +
			`\s*leap59\s+(\d+)\n` +
			`\s*currentUtcOffsetValid\s+(\d+)\n` +
			`\s*ptpTimescale\s+(\d+)\n` +
			`\s*timeTraceable\s+(\d+)\n` +
			`\s*frequencyTraceable\s+(\d+)\n` +
			`\s*timeSource\s+(.+)`,
	// sending: GET GRANDMASTER_SETTINGS_NP
	// 	507c6f.fffe.30fbe8-0 seq 0 RESPONSE MANAGEMENT GRANDMASTER_SETTINGS_NP
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
)

func init() {
	pmcFetcher = fetcher.NewFetcher()
	pmcFetcher.SetPostProcessor(processPMC)
	pmcFetcher.AddCommand(getDateCommand())
	err := pmcFetcher.AddNewCommand(
		"PMC",
		"pmc -u -f /var/run/ptp4l.0.config  'GET GRANDMASTER_SETTINGS_NP'",
		true,
	)
	if err != nil {
		panic(fmt.Errorf("failed to setup PMC fetcher %w", err))
	}
}

func processPMC(result map[string]string) (map[string]any, error) { //nolint:funlen // allow slightly long function
	processedResult := make(map[string]any)
	match := pmcRegEx.FindStringSubmatch(result["PMC"])

	if len(match) == 0 {
		return processedResult, fmt.Errorf("unable to parse pmc output: %s", result["PMC"])
	}

	valuesToConvert := map[string]string{
		"clockClass":            match[1],
		"currentUtcOffset":      match[4],
		"leap61":                match[5],
		"leap59":                match[6],
		"currentUtcOffsetValid": match[7],
		"ptpTimescale":          match[8],
		"timeTraceable":         match[9],
		"frequencyTraceable":    match[10],
	}

	convertedMap, err := MapStringToInt(valuesToConvert)
	if err != nil {
		return processedResult, err
	}

	processedResult["timeSource"] = match[11]
	processedResult["clockAccuracy"] = match[2]
	processedResult["offsetScaledLogVariance"] = match[3]
	processedResult["clockClass"] = convertedMap["clockClass"]
	processedResult["currentUtcOffset"] = convertedMap["currentUtcOffset"]
	processedResult["leap61"] = convertedMap["leap61"]
	processedResult["leap59"] = convertedMap["leap59"]
	processedResult["currentUtcOffsetValid"] = convertedMap["currentUtcOffsetValid"]
	processedResult["ptpTimescale"] = convertedMap["ptpTimescale"]
	processedResult["timeTraceable"] = convertedMap["timeTraceable"]
	processedResult["frequencyTraceable"] = convertedMap["frequencyTraceable"]

	return processedResult, nil
}

// GetPMC returns PMCInfo
func GetPMC(ctx clients.ExecContext) (PMCInfo, error) {
	gmSetting := PMCInfo{}
	err := pmcFetcher.Fetch(ctx, &gmSetting)
	if err != nil {
		log.Debugf("failed to fetch gmSetting %s", err.Error())
		return gmSetting, fmt.Errorf("failed to fetch gmSetting %w", err)
	}
	return gmSetting, nil
}
