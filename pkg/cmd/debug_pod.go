// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

var startDebug = &cobra.Command{
	Use:   "start-debug",
	Short: "Schedule debug container",
	Long:  `Schedule debug container`,
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := clients.GetClientset(kubeConfig)
		utils.IfErrorExitOrPanic(err)
		ctx, err := contexts.GetNetlinkContext(clientset, nodeName, false)
		utils.IfErrorExitOrPanic(err)
		err = ctx.CreatePodAndWait()
		utils.IfErrorExitOrPanic(err)
	},
}

var stopDebug = &cobra.Command{
	Use:   "stop-debug",
	Short: "Remove debug container",
	Long:  `Remove debug container`,
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := clients.GetClientset(kubeConfig)
		utils.IfErrorExitOrPanic(err)
		ctx, err := contexts.GetNetlinkContext(clientset, nodeName, false)
		utils.IfErrorExitOrPanic(err)
		err = ctx.DeletePodAndWait()
		utils.IfErrorExitOrPanic(err)
	},
}

func init() {
	rootCmd.AddCommand(startDebug)
	AddKubeconfigFlag(startDebug)
	AddNodeNameFlag(startDebug)

	rootCmd.AddCommand(stopDebug)
	AddKubeconfigFlag(stopDebug)
	AddNodeNameFlag(stopDebug)
}
