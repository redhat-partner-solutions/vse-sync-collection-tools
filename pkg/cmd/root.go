// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/logging"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/runner"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	defaultCount           int = 10
	defaultPollInterval    int = 1
	defaultDevInfoInterval int = 60
)

type Params struct {
	KubeConfig             string   `mapstructure:"kubeconfig"`
	PTPInterface           string   `mapstructure:"ptp_interface"`
	OutputFile             string   `mapstructure:"output_file"`
	LogLevel               string   `mapstructure:"log_level"`
	CollectorNames         []string `mapstructure:"collectors"`
	PollCount              int      `mapstructure:"poll_count"`
	PollInterval           int      `mapstructure:"poll_rate"`
	DevInfoAnnouceInterval int      `mapstructure:"announce_rate"`
	UseAnalyserJSON        bool     `mapstructure:"use_analyser_json"`
}

func (p *Params) checkForRequiredFields() error {
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

var (
	configFile             string
	kubeConfig             string
	pollCount              int
	pollInterval           int
	devInfoAnnouceInterval int
	ptpInterface           string
	outputFile             string
	logLevel               string
	useAnalyserJSON        bool
	collectorNames         []string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "vse-sync-testsuite",
		Short: "A monitoring tool for PTP related metrics",
		Long:  `A monitoring tool for PTP related metrics.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			logging.SetupLogging(logLevel, os.Stdout)

			if configFile != "" {
				log.Debugf("config: %v", configFile)
				viper.SetConfigFile(configFile)
				err := viper.ReadInConfig()
				utils.IfErrorExitOrPanic(err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			runtimeConfig := &Params{}
			err := viper.Unmarshal(runtimeConfig)
			utils.IfErrorExitOrPanic(err)
			err = runtimeConfig.checkForRequiredFields()
			if err != nil {
				cmd.PrintErrln(err.Error())
				err = cmd.Usage()
				utils.IfErrorExitOrPanic(err)
				os.Exit(1)
			} else {
				collectionRunner := runner.NewCollectorRunner(runtimeConfig.CollectorNames)
				collectionRunner.Run(
					runtimeConfig.KubeConfig,
					runtimeConfig.OutputFile,
					runtimeConfig.PollCount,
					runtimeConfig.PollInterval,
					runtimeConfig.DevInfoAnnouceInterval,
					runtimeConfig.PTPInterface,
					runtimeConfig.UseAnalyserJSON,
				)
			}
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

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("COLLECTOR")
	configureFlags()
}

func configureFlags() { //nolint:funlen // Allow this to get a little long
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to config file")

	rootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to the kubeconfig file")
	err := viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	utils.IfErrorExitOrPanic(err)

	rootCmd.PersistentFlags().StringVarP(&ptpInterface, "interface", "i", "", "Name of the PTP interface")
	err = viper.BindPFlag("interface", rootCmd.PersistentFlags().Lookup("interface"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("ptp_interface", "interface")

	rootCmd.PersistentFlags().IntVarP(
		&pollCount,
		"count",
		"c",
		defaultCount,
		"Number of queries the cluster (-1) means infinite",
	)
	err = viper.BindPFlag("count", rootCmd.PersistentFlags().Lookup("count"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("poll_count", "count")

	rootCmd.PersistentFlags().IntVarP(
		&pollInterval,
		"rate",
		"r",
		defaultPollInterval,
		"Poll interval for querying the cluster. The value will be polled once ever interval. "+
			"Using --rate 10 will cause the value to be polled once every 10 seconds",
	)
	err = viper.BindPFlag("rate", rootCmd.PersistentFlags().Lookup("rate"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("poll_rate", "rate")

	rootCmd.PersistentFlags().IntVarP(
		&devInfoAnnouceInterval,
		"announce",
		"a",
		defaultDevInfoInterval,
		"interval for announcing the dev info",
	)
	err = viper.BindPFlag("announce", rootCmd.PersistentFlags().Lookup("announce"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("announce_rate", "announce")

	rootCmd.PersistentFlags().StringVarP(
		&logLevel,
		"verbosity",
		"v",
		log.WarnLevel.String(),
		"Log level (debug, info, warn, error, fatal, panic)",
	)
	err = viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("log_level", "verbosity")

	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Path to the output file")
	err = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	utils.IfErrorExitOrPanic(err)

	rootCmd.PersistentFlags().BoolVarP(
		&useAnalyserJSON,
		"use-analyser-format",
		"j",
		false,
		"Output in a format to be used by analysers from vse-sync-pp",
	)
	err = viper.BindPFlag("use_analyser_format", rootCmd.PersistentFlags().Lookup("use-analyser-format"))
	utils.IfErrorExitOrPanic(err)

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
	err = viper.BindPFlag("collectors", rootCmd.PersistentFlags().Lookup("collector"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("collector", "collectors")
}
