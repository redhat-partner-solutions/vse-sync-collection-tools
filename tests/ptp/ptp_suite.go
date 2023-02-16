package ptp

import (
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/config"
)

const ptpCustomConfigKey = "ptp_tests_config"

type PtpConfig struct {
	Namespace     string `yaml:"namespace"`
	PodName       string `yaml:"pod_name"`
	Container     string `yaml:"container"`
	InterfaceName string `yaml:"interface_name"`
	TtyTimeout    int    `yaml:"tty_timeout"`
}

var ptpConfig = PtpConfig{}

func init() {
	config.RegisterCustomConfig(ptpCustomConfigKey, &ptpConfig)
}
