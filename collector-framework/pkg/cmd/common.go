// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
)

var (
	kubeConfig      string
	outputFile      string
	useAnalyserJSON bool
)

func AddKubeconfigFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to the kubeconfig file")
	err := targetCmd.MarkFlagRequired("kubeconfig")
	utils.IfErrorExitOrPanic(err)
}

func AddOutputFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Path to the output file")
}

func AddFormatFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().BoolVarP(
		&useAnalyserJSON,
		"use-analyser-format",
		"j",
		false,
		"Output in a format to be used by analysers from vse-sync-pp",
	)
}

func GetKubeConfig() string {
	return kubeConfig
}

func GetUseAnalyserJSON() bool {
	return useAnalyserJSON
}
