// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/logging"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type RootParams struct {
	LogLevel string `mapstructure:"verbosity"`
}

func (p *RootParams) CheckForRequiredFields() error {
	return nil
}

var (
	configFile string
	logLevel   string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "vse-sync-collection-tools",
		Short: "A monitoring tool for PTP related metrics",
		Long:  `A monitoring tool for PTP related metrics.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if configFile != "" {
				log.Debugf("config: %v", configFile)
				viper.SetConfigFile(configFile)
				err := viper.ReadInConfig()
				utils.IfErrorExitOrPanic(err)
			}
			params := RootParams{}
			err := populateParams(cmd, &params)
			if err == nil {
				logging.SetupLogging(params.LogLevel, os.Stdout)
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
	rootCmd.PersistentFlags().StringVarP(
		&logLevel,
		"verbosity",
		"v",
		log.WarnLevel.String(),
		"Log level (debug, info, warn, error, fatal, panic)",
	)
	err := viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))
	utils.IfErrorExitOrPanic(err)
	viper.RegisterAlias("log_level", "verbosity")
}
