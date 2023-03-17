package ptp

import (
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/config"
)

const ptpCustomConfigKey = "ptp_tests_config"


type PtpConfig struct {
	Namespace     string    `yaml:"namespace"`
	Container     string    `yaml:"container"`
	GrandMaster   TgmConfig `yaml:"tgm_config_group"`
	BoundaryClock TbcConfig `yaml:"tbc_config_group"`
}

type TgmConfig struct {
	InterfaceName string `yaml:"interface_name"`
	PodName       string `yaml:"pod_name"`
	TtyTimeout    int    `yaml:"tty_timeout"`
	DpllReads     int    `yaml:"dpll_reads"`
}

type TbcConfig struct {
	InterfaceName      string `yaml:"interface_name"`
	PodName            string `yaml:"pod_name"`
	MaxTransientPeriod int    `yaml:"max_transient_period"`
	MinSteadyPeriod    int    `yaml:"min_steady_period"`
}



var ptpConfig = PtpConfig{}

func init() {
	config.RegisterCustomConfig(ptpCustomConfigKey, &ptpConfig)
}
