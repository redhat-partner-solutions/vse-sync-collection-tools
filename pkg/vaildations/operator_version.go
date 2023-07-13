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
	ptpOperatorVersionID = "PTP Operactor Version is valid"
	MinOperatorVersion   = "4.14.0"
)

type OperatorVersion struct {
	Version string `json:"version"`
	Error   error  `json:"fetchError"`
}

func (ver *OperatorVersion) MarshalJSON() ([]byte, error) {
	var err string
	if ver.Error != nil {
		err = ver.Error.Error()
	}
	return json.Marshal(&struct {
		Version string `json:"version"`
		Error   string `json:"fetchError"`
	}{
		Version: ver.Version,
		Error:   err,
	})
}

type CSV struct {
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
}

func (opVer *OperatorVersion) Verify() error {
	if opVer.Error != nil {
		return opVer.Error
	}
	if semver.Compare(fmt.Sprintf("v%s", opVer.Version), fmt.Sprintf("v%s", MinOperatorVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid firmware version: %s < %s",
				opVer.Version,
				MinOperatorVersion,
			),
		)
	}
	return nil
}

func getOperatorVersion(
	group,
	version,
	resource,
	namespace string,
	client *clients.Clientset,
) (string, error) {
	dynamic := dynamic.NewForConfigOrDie(client.RestConfig)

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamic.Resource(resourceId).Namespace(namespace).
		List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return "", err
	}

	for _, item := range list.Items {
		value := item.Object["spec"]
		crd := &CSV{}
		marsh, _ := json.Marshal(value)
		json.Unmarshal(marsh, crd)
		if crd.DisplayName == "PTP Operator" {
			return crd.Version, nil
		}
	}
	return "", errors.New("failed to find PTP Operator CSV")
}

func (opVer *OperatorVersion) GetID() string {
	return ptpOperatorVersionID
}

func (opVer *OperatorVersion) GetData() any { //nolint:ireturn // data will very for each validation
	return opVer
}

func NewOperatorVersion(client *clients.Clientset) *OperatorVersion {
	version, err := getOperatorVersion(
		"operators.coreos.com",
		"v1alpha1",
		"clusterserviceversions",
		"openshift-ptp",
		client,
	)
	return &OperatorVersion{Version: version, Error: err}
}
