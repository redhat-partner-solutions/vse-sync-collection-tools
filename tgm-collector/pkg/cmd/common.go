// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
)

var (
	ptpInterface string
)

func AddInterfaceFlag(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&ptpInterface, "interface", "i", "", "Name of the PTP interface")
	err := targetCmd.MarkFlagRequired("interface")
	utils.IfErrorExitOrPanic(err)
}
