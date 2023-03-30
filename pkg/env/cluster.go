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

package env

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Node struct {
	Hostname string `yaml:"hostname"`
	Ip       string `yaml:"ip"`
}

type ClusterInfo struct {
	KubeconfigPath string `yaml:"kubeconfig_path"`
	ClusterName    string `yaml:"cluster_name"`
	Nodes          []Node `yaml:"nodes"`
}

var c ClusterInfo

func LoadClusterDefFromFile(filePath string) (*ClusterInfo, error) {
	log.Debugf("Loading cluster definition from file: %s", filePath)

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
