// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/runner"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
)

const (
	defaultDuration        string = "1000s"
	defaultPollInterval    int    = 1
	defaultDevInfoInterval int    = 60
)

type CollectorArgFunc func([]string) map[string]map[string]any

var (
	requestedDurationStr   string
	pollInterval           int
	devInfoAnnouceInterval int
	collectorNames         []string
	runFunc                CollectorArgFunc
)

func SetCollecterArgsFunc(f CollectorArgFunc) {
	runFunc = f
}

// CollectCmd represents the collect command
var CollectCmd = &cobra.Command{
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

		collectorArgs := make(map[string]map[string]any)
		if runFunc != nil {
			log.Debug("No runFunc function is defined")
			collectorArgs = runFunc(collectorNames)
		}

		collectionRunner.Run(
			kubeConfig,
			outputFile,
			requestedDuration,
			pollInterval,
			devInfoAnnouceInterval,
			useAnalyserJSON,
			collectorArgs,
		)
	},
}

func init() { //nolint:funlen // Allow this to get a little long
	RootCmd.AddCommand(CollectCmd)

	AddKubeconfigFlag(CollectCmd)
	AddOutputFlag(CollectCmd)
	AddFormatFlag(CollectCmd)

	CollectCmd.Flags().StringVarP(
		&requestedDurationStr,
		"duration",
		"d",
		defaultDuration,
		"A positive duration string sequence of decimal numbers and a unit suffix, such as \"300ms\", \"1.5h\" or \"2h45m\"."+
			" Valid time units are \"s\", \"m\", \"h\".",
	)
	CollectCmd.Flags().IntVarP(
		&pollInterval,
		"rate",
		"r",
		defaultPollInterval,
		"Poll interval for querying the cluster. The value will be polled once every interval. "+
			"Using --rate 10 will cause the value to be polled once every 10 seconds",
	)
	CollectCmd.Flags().IntVarP(
		&devInfoAnnouceInterval,
		"announce",
		"a",
		defaultDevInfoInterval,
		"interval at which to emit the device info summary to the targeted output.",
	)
	defaultCollectorNames := make([]string, 0)
	defaultCollectorNames = append(defaultCollectorNames, runner.All)
	CollectCmd.Flags().StringSliceVarP(
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
