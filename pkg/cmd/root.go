// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/logging"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/runner"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	defaultCount              int     = 10
	defaultPollRate           float64 = 1.0
	defaultDevInfoAnnouceRate float64 = 0.17
)

var (
	kubeConfig         string
	pollCount          int
	pollRate           float64
	devInfoAnnouceRate float64
	ptpInterface       string
	outputFile         string
	logLevel           string
	useAnalyserJSON    bool
	collectorNames     []string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "vse-sync-testsuite",
		Short: "A monitoring tool for PTP related metrics",
		Long:  `A monitoring tool for PTP related metrics.`,
		Run: func(cmd *cobra.Command, args []string) {
			logging.SetupLogging(logLevel, os.Stdout)
			collectionRunner := runner.NewCollectorRunner(collectorNames)
			collectionRunner.Run(
				kubeConfig,
				outputFile,
				pollCount,
				pollRate,
				devInfoAnnouceRate,
				ptpInterface,
				useAnalyserJSON,
			)
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() { //nolint:funlen // Allow this to get a little long
	rootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to the kubeconfig file")
	err := rootCmd.MarkPersistentFlagRequired("kubeconfig")
	utils.IfErrorPanic(err)

	rootCmd.PersistentFlags().StringVarP(&ptpInterface, "interface", "i", "", "Name of the PTP interface")
	err = rootCmd.MarkPersistentFlagRequired("interface")
	utils.IfErrorPanic(err)

	rootCmd.PersistentFlags().IntVarP(
		&pollCount,
		"count",
		"c",
		defaultCount,
		"Number of queries the cluster (-1) means infinite",
	)
	rootCmd.PersistentFlags().Float64VarP(
		&pollRate,
		"rate",
		"r",
		defaultPollRate,
		"Poll rate for querying the cluster",
	)
	rootCmd.PersistentFlags().Float64VarP(
		&devInfoAnnouceRate,
		"announce",
		"a",
		defaultDevInfoAnnouceRate,
		"Rate announcing the dev info",
	)

	rootCmd.PersistentFlags().StringVarP(
		&logLevel,
		"verbosity",
		"v",
		log.WarnLevel.String(),
		"Log level (debug, info, warn, error, fatal, panic)",
	)
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Path to the output file")

	rootCmd.PersistentFlags().BoolVarP(
		&useAnalyserJSON,
		"use-analyser-format",
		"j",
		false,
		"Output in a format to be used by analysers from vse-sync-pp",
	)

	defaultCollectorNames := make([]string, 0)
	defaultCollectorNames = append(defaultCollectorNames, runner.All)
	rootCmd.PersistentFlags().StringSliceVarP(
		&collectorNames,
		"collector",
		"s",
		defaultCollectorNames,
		fmt.Sprintf(
			"the collectors you wish to run (case-insensitive):\n"+
				"\trequired collectors: %s (will be automatically added)\n"+
				"\toptional collectors: %s",
			strings.Join(runner.RequiredCollectorNames, ", "),
			strings.Join(runner.OptionalCollectorNames, ", "),
		),
	)
}
