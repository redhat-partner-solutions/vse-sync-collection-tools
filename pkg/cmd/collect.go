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
	"github.com/spf13/viper"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors"
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
)

type CollectionParams struct {
	KubeConfig             string   `mapstructure:"kubeconfig"`
	PTPInterface           string   `mapstructure:"ptp_interface"`
	OutputFile             string   `mapstructure:"output_file"`
	Duration               string   `mapstructure:"duration"`
	LogsOutputFile         string   `mapstructure:"logs_output"`
	TempDir                string   `mapstructure:"tempdir"`
	CollectorNames         []string `mapstructure:"collectors"`
	PollInterval           int      `mapstructure:"poll_interval"`
	DevInfoAnnouceInterval int      `mapstructure:"announce_interval"`
	UseAnalyserJSON        bool     `mapstructure:"use_analyser_format"`
	IncludeLogTimestamps   bool     `mapstructure:"log_timestamps"`
	KeepDebugFiles         bool     `mapstructure:"keep_debug_files"`
}

func (p *CollectionParams) CheckForRequiredFields() error {
	missing := make([]string, 0)
	if p.KubeConfig == "" {
		missing = append(missing, "kubeconfig")
	}
	if p.PTPInterface == "" {
		missing = append(missing, "interface")
	}
	if len(missing) > 0 {
		return fmt.Errorf(`required flag(s) "%s" not set`, strings.Join(missing, `", "`))
	}
	return nil
}

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Run the collector tool",
	Long:  `Run the collector tool to gather data from your target cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		runtimeConfig := &CollectionParams{}
		err := populateParams(cmd, runtimeConfig)
		utils.IfErrorExitOrPanic(err)

		requestedDuration, err := time.ParseDuration(runtimeConfig.Duration)
		utils.IfErrorExitOrPanic(err)
		if requestedDuration.Nanoseconds() < 0 {
			log.Panicf("Requested duration must be positive")
		}

		for _, c := range collectorNames {
			if (c == collectors.LogsCollectorName || c == runner.All) && logsOutputFile == "" {
				utils.IfErrorExitOrPanic(utils.NewMissingInputError(
					errors.New("if Logs collector is selected you must also provide a log output file")),
				)
			}
		}

		if strings.Contains(tempDir, "~") {
			usr, err := user.Current()
			if err != nil {
				log.Fatal("Failed to fetch current user so could not resolve tempdir")
			}
			if tempDir == "~" {
				tempDir = usr.HomeDir
			} else if strings.HasPrefix(tempDir, "~/") {
				tempDir = filepath.Join(usr.HomeDir, tempDir[2:])
			}
		}

		if err := os.MkdirAll(tempDir, tempdirPerm); err != nil {
			log.Fatal(err)
		}

		collectionRunner := runner.NewCollectorRunner(runtimeConfig.CollectorNames)
		collectionRunner.Run(
			runtimeConfig.KubeConfig,
			runtimeConfig.OutputFile,
			requestedDuration,
			runtimeConfig.PollInterval,
			runtimeConfig.DevInfoAnnouceInterval,
			runtimeConfig.PTPInterface,
			runtimeConfig.UseAnalyserJSON,
			runtimeConfig.LogsOutputFile,
			runtimeConfig.IncludeLogTimestamps,
			runtimeConfig.TempDir,
			runtimeConfig.KeepDebugFiles,
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
	err := viper.BindPFlag("duration", collectCmd.Flags().Lookup("duration"))
	utils.IfErrorExitOrPanic(err)

	collectCmd.Flags().IntVarP(
		&pollInterval,
		"rate",
		"r",
		defaultPollInterval,
		"Poll interval for querying the cluster. The value will be polled once every interval. "+
			"Using --rate 10 will cause the value to be polled once every 10 seconds",
	)
	err = viper.BindPFlag("poll_interval", collectCmd.Flags().Lookup("rate"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("poll_rate", "poll_interval")
	viper.RegisterAlias("rate", "poll_interval")

	collectCmd.Flags().IntVarP(
		&devInfoAnnouceInterval,
		"announce",
		"a",
		defaultDevInfoInterval,
		"interval at which to emit the device info summary to the targeted output.",
	)
	err = viper.BindPFlag("announce_interval", collectCmd.Flags().Lookup("announce"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("announce_rate", "announce_interval")
	viper.RegisterAlias("announce", "announce_interval")

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
	err = viper.BindPFlag("collectors", collectCmd.Flags().Lookup("collector"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("collector", "collectors")

	collectCmd.Flags().StringVarP(
		&logsOutputFile,
		"logs-output", "l", "",
		"Path to the logs output file. This is required when using the logs collector",
	)
	err = viper.BindPFlag("logs_output", collectCmd.Flags().Lookup("logs-output"))
	utils.IfErrorExitOrPanic(err)

	collectCmd.Flags().BoolVar(
		&includeLogTimestamps,
		"log-timestamps", defaultIncludeLogTimestamps,
		"Specifies if collected logs should include timestamps or not. (default is false)",
	)
	err = viper.BindPFlag("log_timestamps", collectCmd.Flags().Lookup("log-timestamps"))
	utils.IfErrorExitOrPanic(err)

	collectCmd.Flags().StringVarP(&tempDir, "tempdir", "t", defaultTempDir,
		"Directory for storing temp/debug files. Must exist.")
	err = viper.BindPFlag("tempdir", collectCmd.Flags().Lookup("tempdir"))
	utils.IfErrorExitOrPanic(err)

	collectCmd.Flags().BoolVar(&keepDebugFiles, "keep", defaultKeepDebugFiles, "Keep debug files")
	err = viper.BindPFlag("keep_debug_files", collectCmd.Flags().Lookup("keep"))
	utils.IfErrorExitOrPanic(err)
}
