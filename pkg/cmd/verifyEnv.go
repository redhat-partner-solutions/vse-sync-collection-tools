// SPDX-License-Identifier: GPL-2.0-or-later

/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/verify"
)

// verifyEnvCmd represents the verifyEnv command
var verifyEnvCmd = &cobra.Command{
	Use:   "verifyEnv",
	Short: "verify the environment for collection",
	Long:  `verify the environment for collection`,
	Run: func(cmd *cobra.Command, args []string) {
		verify.Verify(ptpInterface, kubeConfig, useAnalyserJSON)
	},
}

func init() {
	rootCmd.AddCommand(verifyEnvCmd)
	AddCommonFlags(verifyEnvCmd)
}
