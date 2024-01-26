// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"
	"strings"
	"time"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/utils"
)

var dateCmd *clients.Cmd

func formatTimestampAsRFC3339Nano(s string) (string, error) {
	timestamp, err := utils.ParseTimestamp(strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("failed to parse timestamp %w", err)
	}
	return timestamp.Format(time.RFC3339Nano), nil
}

func getDateCommand() *clients.Cmd {
	if dateCmd == nil {
		dateCmdInst, err := clients.NewCmd("date", "date +%s.%N")
		if err != nil {
			panic(err)
		}
		dateCmdInst.SetOutputProcessor(formatTimestampAsRFC3339Nano)
		dateCmd = dateCmdInst
	}
	return dateCmd
}

func init() {
	getDateCommand()
}
