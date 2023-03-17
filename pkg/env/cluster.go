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
	NgenImgPath    string `yaml:"ngen_path"`
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
