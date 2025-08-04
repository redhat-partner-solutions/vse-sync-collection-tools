// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/constants"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/detect"
)

// detectCards represents the detect command which prints the configured interfaces
var detectCards = &cobra.Command{
	Use:   "detect",
	Short: "Run the interface detection tool",
	Long:  `Run the interface detection tool to check for multi-card setups`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate clock type
		clockTypeUpper := strings.ToUpper(clockType)
		if clockTypeUpper != constants.ClockTypeGM && clockTypeUpper != constants.ClockTypeBC {
			fmt.Fprintf(os.Stderr, "Error: Invalid clock type '%s'. Must be either '%s' or '%s'\n", clockType, constants.ClockTypeGM, constants.ClockTypeBC)
			os.Exit(1)
		}
		detect.Detect(kubeConfig, nodeName, useAnalyserJSON, clockTypeUpper)
	},
}

func init() {
	rootCmd.AddCommand(detectCards)
	AddKubeconfigFlag(detectCards)
	AddFormatFlag(detectCards)
	AddNodeNameFlag(detectCards)
	AddClockTypeFlag(detectCards)
}
