// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/runner"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	defaultDuration        string = "1000s"
	defaultPollInterval    int    = 1
	defaultDevInfoInterval int    = 60
)

var (
	requestedDurationStr   string
	pollInterval           int
	devInfoAnnouceInterval int
	collectorNames         []string
	logsOutputFile         string
)

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Run the collector tool",
	Long:  `Run the collector tool to gather data from your target cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		collectionRunner := runner.NewCollectorRunner(collectorNames)

		requestedDuration, err := time.ParseDuration(requestedDurationStr)
		if requestedDuration.Nanoseconds() < 0 {
			log.Panicf("Requested duration must be positive")
		}
		utils.IfErrorExitOrPanic(err)

		for _, c := range collectorNames {
			if c == collectors.LogsCollectorName && logsOutputFile == "" {
				utils.IfErrorExitOrPanic(utils.NewMissingInputError(
					errors.New("if Logs collector is selected you must also provide a log output file")),
				)
			}
		}

		collectionRunner.Run(
			kubeConfig,
			outputFile,
			requestedDuration,
			pollInterval,
			devInfoAnnouceInterval,
			ptpInterface,
			useAnalyserJSON,
			logsOutputFile,
		)
	},
}

func init() { //nolint:funlen // Allow this to get a little long
	rootCmd.AddCommand(collectCmd)

	AddKubeconfigFlag(collectCmd)
	AddOutputFlag(collectCmd)
	AddFormatFlag(collectCmd)
	AddInterfaceFlag(collectCmd)

	collectCmd.Flags().StringVarP(
		&requestedDurationStr,
		"duration",
		"d",
		defaultDuration,
		"A positive duration string sequence of decimal numbers and a unit suffix, such as \"300ms\", \"1.5h\" or \"2h45m\"."+
			" Valid time units are \"s\", \"m\", \"h\".",
	)
	collectCmd.Flags().IntVarP(
		&pollInterval,
		"rate",
		"r",
		defaultPollInterval,
		"Poll interval for querying the cluster. The value will be polled once every interval. "+
			"Using --rate 10 will cause the value to be polled once every 10 seconds",
	)
	collectCmd.Flags().IntVarP(
		&devInfoAnnouceInterval,
		"announce",
		"a",
		defaultDevInfoInterval,
		"interval at which to emit the device info summary to the targeted output.",
	)
	defaultCollectorNames := make([]string, 0)
	defaultCollectorNames = append(defaultCollectorNames, runner.All)
	collectCmd.Flags().StringSliceVarP(
		&collectorNames,
		"collector",
		"s",
		defaultCollectorNames,
		fmt.Sprintf(
			"the collectors you wish to run (case-insensitive):\n"+
				"\trequired collectors: %s (will be automatically added)\n"+
				"\toptional collectors: %s\n"+
				"\topt in collectors: %s (note: these are not included by default)",
			strings.Join(runner.RequiredCollectorNames, ", "),
			strings.Join(runner.OptionalCollectorNames, ", "),
			strings.Join(runner.OptInCollectorNames, ", "),
		),
	)

	collectCmd.Flags().StringVarP(
		&logsOutputFile,
		"logs-output", "l", "",
		"Path to the logs output file. This is required when using the opt in logs collector",
	)
}
