// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/mod/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	clusterVersionID  = "Opensift Cluster Version is valid"
	MinClusterVersion = "4.13.3"
)

type ClusterVersion struct {
	Version string
	Error   error
}

func (ver *ClusterVersion) MarshalJSON() ([]byte, error) {
	return MarshalVersionAndError(&VersionWithError{
		Version: ver.Version,
		Error:   ver.Error,
	})
}

func (clusterVer *ClusterVersion) Verify() error {
	if clusterVer.Error != nil {
		return clusterVer.Error
	}
	ver := fmt.Sprintf("v%s", clusterVer.Version)
	if !semver.IsValid(ver) {
		return fmt.Errorf("could not parse version %s", ver)
	}
	if semver.Compare(ver, fmt.Sprintf("v%s", MinClusterVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid firmware version: %s < %s",
				clusterVer.Version,
				MinClusterVersion,
			),
		)
	}
	return nil
}

type Status struct {
	Desired clusterVersion `json:"desired"`
}

type clusterVersion struct {
	Version string `json:"version"`
}

func getClusterVersion(
	group,
	version,
	resource string,
	client *clients.Clientset,
) (string, error) {
	dynamic := dynamic.NewForConfigOrDie(client.RestConfig)

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamic.Resource(resourceId).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, item := range list.Items {
		value := item.Object["status"]
		status := &Status{}
		marsh, _ := json.Marshal(value)
		json.Unmarshal(marsh, status)
		return status.Desired.Version, nil
	}
	return "", errors.New("failed to find PTP Operator CSV")
}

func (clusterVer *ClusterVersion) GetID() string {
	return clusterVersionID
}

func (clusterVer *ClusterVersion) GetData() any { //nolint:ireturn // data will very for each validation
	return clusterVer
}

func NewClusterVersion(client *clients.Clientset) *ClusterVersion {
	version, err := getClusterVersion(
		"config.openshift.io",
		"v1",
		"clusterversions",
		client,
	)
	return &ClusterVersion{Version: version, Error: err}
}
