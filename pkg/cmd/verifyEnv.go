// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/constants"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/verify"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "environment based actions",
	Long:  `environment based actions`,
}

// verifyEnvCmd represents the verifyEnv command
var verifyEnvCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify the environment is ready for collection",
	Long:  `verify the environment is ready for collection`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate clock type
		clockTypeUpper := strings.ToUpper(clockType)
		if clockTypeUpper != constants.ClockTypeGM && clockTypeUpper != constants.ClockTypeBC {
			fmt.Fprintf(os.Stderr, "Error: Invalid clock type '%s'. Must be either '%s' or '%s'\n", clockType, constants.ClockTypeGM, constants.ClockTypeBC)
			os.Exit(1)
		}
		verify.Verify(ptpInterface, kubeConfig, useAnalyserJSON, nodeName, clockTypeUpper)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(verifyEnvCmd)
	AddKubeconfigFlag(verifyEnvCmd)
	AddOutputFlag(verifyEnvCmd)
	AddFormatFlag(verifyEnvCmd)
	AddInterfaceFlag(verifyEnvCmd)
	AddNodeNameFlag(verifyEnvCmd)
	AddClockTypeFlag(verifyEnvCmd)
}
