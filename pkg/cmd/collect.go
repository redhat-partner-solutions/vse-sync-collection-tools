// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/constants"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/runner"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	defaultDuration             string = "1000s"
	defaultPollInterval         int    = 1
	defaultDevInfoInterval      int    = 60
	defaultIncludeLogTimestamps bool   = false
	defaultTempDir              string = "."
	defaultKeepDebugFiles       bool   = false
	tempdirPerm                        = 0755
)

var (
	requestedDurationStr   string
	pollInterval           int
	devInfoAnnouceInterval int
	collectorNames         []string
	logsOutputFile         string
	includeLogTimestamps   bool
	tempDir                string
	keepDebugFiles         bool
	unmanagedDebugPod      bool
)

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Run the collector tool",
	Long:  `Run the collector tool to gather data from your target cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate clock type
		clockTypeUpper := strings.ToUpper(clockType)
		if clockTypeUpper != constants.ClockTypeGM && clockTypeUpper != constants.ClockTypeBC {
			fmt.Fprintf(os.Stderr, "Error: Invalid clock type '%s'. Must be either '%s' or '%s'\n", clockType, constants.ClockTypeGM, constants.ClockTypeBC)
			os.Exit(1)
		}

		collectionRunner := runner.NewCollectorRunner(collectorNames)

		requestedDuration, err := time.ParseDuration(requestedDurationStr)
		if requestedDuration.Nanoseconds() < 0 {
			log.Panicf("Requested duration must be positive")
		}
		utils.IfErrorExitOrPanic(err)

		for _, c := range collectorNames {
			if (c == collectors.LogsCollectorName || c == runner.All) && logsOutputFile == "" {
				utils.IfErrorExitOrPanic(utils.NewMissingInputError(
					errors.New("if Logs collector is selected you must also provide a log output file")),
				)
			}
		}

		if strings.Contains(tempDir, "~") {
			var usr *user.User
			usr, err = user.Current()
			if err != nil {
				log.Fatal("Failed to fetch current user so could not resolve tempdir")
			}
			if tempDir == "~" {
				tempDir = usr.HomeDir
			} else if strings.HasPrefix(tempDir, "~/") {
				tempDir = filepath.Join(usr.HomeDir, tempDir[2:])
			}
		}

		if err = os.MkdirAll(tempDir, tempdirPerm); err != nil {
			log.Fatal(err)
		}

		constuctor, err := collectors.NewCollectionConstructor(
			kubeConfig,
			useAnalyserJSON,
			outputFile,
			ptpInterface,
			nodeName,
			logsOutputFile,
			tempDir,
			pollInterval,
			devInfoAnnouceInterval,
			includeLogTimestamps,
			keepDebugFiles,
			unmanagedDebugPod,
			clockTypeUpper,
		)
		utils.IfErrorExitOrPanic(err)

		collectionRunner.Run(
			requestedDuration,
			constuctor,
		)
	},
}

func init() { //nolint:funlen // Allow this to get a little long
	rootCmd.AddCommand(collectCmd)

	AddKubeconfigFlag(collectCmd)
	AddOutputFlag(collectCmd)
	AddFormatFlag(collectCmd)
	AddInterfaceFlag(collectCmd)
	AddNodeNameFlag(collectCmd)
	AddClockTypeFlag(collectCmd)

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
				"\toptional collectors: %s",
			strings.Join(runner.RequiredCollectorNames, ", "),
			strings.Join(runner.OptionalCollectorNames, ", "),
		),
	)

	collectCmd.Flags().StringVarP(
		&logsOutputFile,
		"logs-output", "l", "",
		"Path to the logs output file. This is required when using the logs collector",
	)
	collectCmd.Flags().BoolVar(
		&includeLogTimestamps,
		"log-timestamps", defaultIncludeLogTimestamps,
		"Specifies if collected logs should include timestamps or not. (default is false)",
	)

	collectCmd.Flags().StringVarP(&tempDir, "tempdir", "t", defaultTempDir,
		"Directory for storing temp/debug files. Must exist.")
	collectCmd.Flags().BoolVar(&keepDebugFiles, "keep", defaultKeepDebugFiles, "Keep debug files")

	collectCmd.Flags().BoolVar(&unmanagedDebugPod, "unmanaged-debug-pod", false, "Do not manage debug pod")
}
