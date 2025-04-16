// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/detect"
)

// detectCards represents the detect command which prints the configured interfaces
var detectCards = &cobra.Command{
	Use:   "detect",
	Short: "Run the interface detection tool",
	Long:  `Run the interface detection tool to check for multi-card setups`,
	Run: func(cmd *cobra.Command, args []string) {
		detect.Detect(kubeConfig, nodeName, useAnalyserJSON)
	},
}

func init() {
	rootCmd.AddCommand(detectCards)
	AddKubeconfigFlag(detectCards)
	AddFormatFlag(detectCards)
	AddNodeNameFlag(detectCards)
}
