// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

var (
	kubeConfig      string
	outputFile      string
	useAnalyserJSON bool
	ptpInterface    string
)

func AddKubeconfigFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to the kubeconfig file")
	err := viper.BindPFlag("kubeconfig", targetCmd.Flags().Lookup("kubeconfig"))
	utils.IfErrorExitOrPanic(err)
}

func AddOutputFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Path to the output file")
	err := viper.BindPFlag("output_file", targetCmd.Flags().Lookup("output"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("output", "output_file")
}

func AddFormatFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().BoolVarP(
		&useAnalyserJSON,
		"use-analyser-format",
		"j",
		false,
		"Output in a format to be used by analysers from vse-sync-pp",
	)
	err := viper.BindPFlag("use_analyser_format", targetCmd.Flags().Lookup("use-analyser-format"))
	utils.IfErrorExitOrPanic(err)
}

func AddInterfaceFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&ptpInterface, "interface", "i", "", "Name of the PTP interface")
	err := viper.BindPFlag("ptp_interface", targetCmd.Flags().Lookup("interface"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("interface", "ptp_interface")
}

type Params interface {
	CheckForRequiredFields() error
}

func populateParams(cmd *cobra.Command, params Params) error {
	err := viper.Unmarshal(params)
	utils.IfErrorExitOrPanic(err)
	err = params.CheckForRequiredFields()
	if err != nil {
		cmd.PrintErrln(err.Error())
		err = cmd.Usage()
		utils.IfErrorExitOrPanic(err)
		os.Exit(int(utils.InvalidArgs))
		return fmt.Errorf("failed to populate params: %w", err)
	}
	return nil
}
