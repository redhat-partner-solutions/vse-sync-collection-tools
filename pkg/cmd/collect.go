// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/runner"
)

var (
	pollCount              int
	pollInterval           int
	devInfoAnnouceInterval int
	collectorNames         []string
)

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Run the collector tool",
	Long:  `Run the collector tool to gather data from your target cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		collectionRunner := runner.NewCollectorRunner(collectorNames)
		collectionRunner.Run(
			kubeConfig,
			outputFile,
			pollCount,
			pollInterval,
			devInfoAnnouceInterval,
			ptpInterface,
			useAnalyserJSON,
		)
	},
}

func init() { //nolint:funlen // Allow this to get a little long
	rootCmd.AddCommand(collectCmd)

	AddCommonFlags(collectCmd)

	collectCmd.Flags().IntVarP(
		&pollCount,
		"count",
		"c",
		defaultCount,
		"Number of data points to collect from the cluster. Use `-1` to poll forever",
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
				"\toptional collectors: %s",
			strings.Join(runner.RequiredCollectorNames, ", "),
			strings.Join(runner.OptionalCollectorNames, ", "),
		),
	)
}
