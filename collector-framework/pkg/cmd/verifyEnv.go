// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

type verifyFunc func(kubeconfig string, useAnalyserJSON bool)

var verify verifyFunc

func SetVerifyFunc(f verifyFunc) {
	verify = f
}

var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "environment based actions",
	Long:  `environment based actions`,
}

// VerifyEnvCmd represents the verifyEnv command
var VerifyEnvCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify the environment is ready for collection",
	Long:  `verify the environment is ready for collection`,
	Run: func(cmd *cobra.Command, args []string) {
		if verify == nil {
			log.Fatal("Verify command was not registered")
		}
		verify(kubeConfig, useAnalyserJSON)
	},
}

func init() {
	RootCmd.AddCommand(EnvCmd)
	EnvCmd.AddCommand(VerifyEnvCmd)
	AddKubeconfigFlag(VerifyEnvCmd)
	AddOutputFlag(VerifyEnvCmd)
	AddFormatFlag(VerifyEnvCmd)
}
