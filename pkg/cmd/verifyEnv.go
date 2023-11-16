// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/verify"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "environment based actions",
	Long:  `environment based actions`,
}

type VerifyParams struct {
	KubeConfig      string `mapstructure:"kubeconfig"`
	PTPInterface    string `mapstructure:"ptp_interface"`
	UseAnalyserJSON bool   `mapstructure:"use_analyser_format"`
}

func (p *VerifyParams) CheckForRequiredFields() error {
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

// verifyEnvCmd represents the verifyEnv command
var verifyEnvCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify the environment is ready for collection",
	Long:  `verify the environment is ready for collection`,
	Run: func(cmd *cobra.Command, args []string) {
		runtimeConfig := &VerifyParams{}
		err := populateParams(cmd, runtimeConfig)
		if err == nil {
			verify.Verify(
				runtimeConfig.PTPInterface,
				runtimeConfig.KubeConfig,
				runtimeConfig.UseAnalyserJSON,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(verifyEnvCmd)
	AddKubeconfigFlag(verifyEnvCmd)
	AddOutputFlag(verifyEnvCmd)
	AddFormatFlag(verifyEnvCmd)
	AddInterfaceFlag(verifyEnvCmd)
}
