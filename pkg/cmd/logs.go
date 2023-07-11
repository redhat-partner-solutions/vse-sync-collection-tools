// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/grablogs"
)

var (
	outputDirPath string
	logWindow     string

	// logsCmd represents the logs command
	logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "collect container logs",
		Long:  `collect container logs`,
		Run: func(cmd *cobra.Command, args []string) {
			grablogs.GrabLogs(kubeConfig, logWindow, outputDirPath)
		},
	}
)

func init() {
	rootCmd.AddCommand(logsCmd)
	AddKubeconfigFlag(logsCmd)

	logsCmd.PersistentFlags().StringVarP(
		&outputDirPath,
		"output-dir",
		"o",
		".",
		"Optional. Specify the output directory. Target must exist. Defaults to working directory.",
	)

	logsCmd.PersistentFlags().StringVar(
		&logWindow,
		"since",
		"",
		"Optional. Only get logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs if omitted.",
	)
}
