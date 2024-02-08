// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/clients"
)

const (
	clusterVersionID  = TGMEnvVerPath + "/RHOCP/"
	MinClusterVersion = "4.14.0-0" // trailing -0 is required to allow preGA version
)

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
	dynamicClient := dynamic.NewForConfigOrDie(client.RestConfig)

	resourceID := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamicClient.Resource(resourceID).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch cluster version %w", err)
	}

	for _, item := range list.Items {
		value := item.Object["status"]
		status := &Status{}
		marsh, err := json.Marshal(value)
		if err != nil {
			log.Debug("failed to marshal cluster version status", err)
			continue
		}
		err = json.Unmarshal(marsh, status)
		if err != nil {
			log.Debug("failed to marshal cluster version status", err)
			continue
		}
		return status.Desired.Version, nil
	}
	return "", errors.New("failed to find PTP Operator CSV")
}

func NewClusterVersion(client *clients.Clientset) *VersionWithErrorCheck {
	version, err := getClusterVersion(
		"config.openshift.io",
		"v1",
		"clusterversions",
		client,
	)
	return &VersionWithErrorCheck{
		VersionCheck: VersionCheck{
			id:           clusterVersionID,
			Version:      version,
			checkVersion: version,
			MinVersion:   MinClusterVersion,
			order:        clusterVersionOrdering,
		},
		Error: err,
	}
}
