// SPDX-License-Identifier: GPL-2.0-or-later

package runner

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors"
)

var (
	OptionalCollectorNames []string
	RequiredCollectorNames []string
	All                    string = "all"
)

func init() {
	registry := collectors.GetRegistry()
	OptionalCollectorNames = registry.GetOptionalNames()
	RequiredCollectorNames = registry.GetRequiredNames()
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
	collectorNames = append(collectorNames, RequiredCollectorNames...)
	for _, name := range selectedCollectors {
		switch {
		case strings.EqualFold(name, "all"):
			collectorNames = append(collectorNames, OptionalCollectorNames...)
		case strings.EqualFold(name, "defaults"):
			collectorNames = append(collectorNames, OptionalCollectorNames...)
		case isIn(name, collectorNames):
			continue
		case isIn(name, OptionalCollectorNames):
			collectorNames = append(collectorNames, name)
		default:
			log.Errorf("Unknown collector %s. Ignored", name)
		}
	}
	collectorNames = removeDuplicates(collectorNames)
	return collectorNames
}
