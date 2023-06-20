// SPDX-License-Identifier: GPL-2.0-or-later

package runner

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors"
)

var (
	optionalCollectorNames []string
	requiredCollectorNames []string
	All                    string = "all"
)

func init() {
	optionalCollectorNames = []string{collectors.DPLLCollectorName, collectors.GPSCollectorName}
	requiredCollectorNames = []string{collectors.DevInfoCollectorName}
}

func isIn(name string, arr []string) bool {
	for _, arrVal := range arr {
		if name == arrVal {
			return true
		}
	}
	return false
}

func removeDuplicates(arr []string) []string {
	res := make([]string, 0)
	for _, name := range arr {
		if !isIn(name, res) {
			res = append(res, name)
		}
	}
	return res
}

// GetCollectorsToRun returns a slice containing the names of the
// collectors to be run it will enfore that required colletors
// are returned
func GetCollectorsToRun(selectedCollectors []string) []string {
	collectorNames := make([]string, 0)
	for _, name := range selectedCollectors {
		switch {
		case strings.EqualFold(name, "all"):
			collectorNames = append(collectorNames, requiredCollectorNames...)
			collectorNames = append(collectorNames, optionalCollectorNames...)
			collectorNames = removeDuplicates(collectorNames)
			return collectorNames
		case isIn(name, collectorNames):
			continue
		case isIn(name, requiredCollectorNames), isIn(name, optionalCollectorNames):
			collectorNames = append(collectorNames, name)
		default:
			log.Errorf("Unknown collector %s. Ignored", name)
		}
	}
	missingCollectors := make([]string, 0)
	for _, requiredName := range requiredCollectorNames {
		if !isIn(requiredName, collectorNames) {
			missingCollectors = append(missingCollectors, requiredName)
		}
	}
	if len(missingCollectors) > 0 {
		log.Warnf(
			"The following required collectors were missing %s. They will be added",
			strings.Join(missingCollectors, ","),
		)
		collectorNames = append(collectorNames, missingCollectors...)
	}
	return collectorNames
}