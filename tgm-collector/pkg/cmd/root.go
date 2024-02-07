// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	fCmd "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/cmd"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors"
)

func Execute() {
	fCmd.Execute()
}

func init() {
	collectors.IncludeCollectorsNoOp()
}
