// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

var (
	kubeConfig      string
	outputFile      string
	useAnalyserJSON bool
	ptpInterface    string
)

func AddCommonFlags(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to the kubeconfig file")
	err := targetCmd.MarkFlagRequired("kubeconfig")
	utils.IfErrorExitOrPanic(err)

	targetCmd.Flags().StringVarP(&ptpInterface, "interface", "i", "", "Name of the PTP interface")
	err = targetCmd.MarkFlagRequired("interface")
	utils.IfErrorExitOrPanic(err)

	targetCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Path to the output file")

	targetCmd.Flags().BoolVarP(
		&useAnalyserJSON,
		"use-analyser-format",
		"j",
		false,
		"Output in a format to be used by analysers from vse-sync-pp",
	)
}
