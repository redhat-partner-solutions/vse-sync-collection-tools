// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

			// Load config from file or from env vars
			if configFile != "" {
				log.Debugf("config: %v", configFile)
				viper.SetConfigFile(configFile)
			}
			err := viper.ReadInConfig()
			utils.IfErrorExitOrPanic(err)

			// Update flag values with values from viper
			log.Debugf("Updating CLI flags with config")
			cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				log.Debugf("flagName: %v", f.Name)
				if viper.IsSet(f.Name) {
					log.Debugf("flagValue: %v", viper.GetString(f.Name))
					err = cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
					utils.IfErrorExitOrPanic(err)
				}
			})
		},
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
	viper.AutomaticEnv()
	viper.SetEnvPrefix("COLLECTOR")

	rootCmd.Flags().StringVar(&configFile, "config", "", "Path to config file")

	rootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to the kubeconfig file")
	err := rootCmd.MarkPersistentFlagRequired("kubeconfig")
	utils.IfErrorExitOrPanic(err)

	rootCmd.PersistentFlags().StringVarP(&ptpInterface, "interface", "i", "", "Name of the PTP interface")
	err = rootCmd.MarkPersistentFlagRequired("interface")
	utils.IfErrorExitOrPanic(err)

	rootCmd.PersistentFlags().IntVarP(
		&pollCount,
		"count",
		"c",
		defaultCount,
		"Number of queries the cluster (-1) means infinite",
	)
	rootCmd.PersistentFlags().IntVarP(
		&pollInterval,
		"rate",
		"r",
		defaultPollInterval,
		"Poll interval for querying the cluster. The value will be polled once ever interval. "+
			"Using --rate 10 will cause the value to be polled once every 10 seconds",
	)
	rootCmd.PersistentFlags().IntVarP(
		&devInfoAnnouceInterval,
		"announce",
		"a",
		defaultDevInfoInterval,
		"interval for announcing the dev info",
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
