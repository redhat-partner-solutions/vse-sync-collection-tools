// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/logging"
)

var (
	logLevel string

	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "vse-sync-testsuite",
		Short: "A monitoring tool for PTP related metrics",
		Long:  `A monitoring tool for PTP related metrics.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logging.SetupLogging(logLevel, os.Stdout)
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(
		&logLevel,
		"verbosity",
		"v",
		log.WarnLevel.String(),
		"Log level (debug, info, warn, error, fatal, panic)",
	)
}
