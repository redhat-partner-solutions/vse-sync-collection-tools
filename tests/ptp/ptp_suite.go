// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	DpllReads     int    `yaml:"dpll_reads"`
}

var ptpConfig = PtpConfig{}

func init() {
	config.RegisterCustomConfig(ptpCustomConfigKey, &ptpConfig)
}
