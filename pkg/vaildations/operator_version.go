package vaildations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	// "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
	"golang.org/x/mod/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	ptpOperatorVersionID = "PTP Operactor Version is valid"
	MinOperatorVersion   = "4.14.0"
)

type OperatorVersion struct {
	Version   string             `json:"version"`
	group     string             `json:"-"`
	version   string             `json:"-"`
	resource  string             `json:"-"`
	namespace string             `json:"-"`
	client    *clients.Clientset `json:"-"`
}

type CSV struct {
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
}

func (opVer *OperatorVersion) Verify() error {
	version, err := getOperatorVersion(opVer)
	opVer.Version = version
	if err != nil {
		return err
	}
	if semver.Compare(fmt.Sprintf("v%s", version), fmt.Sprintf("v%s", MinOperatorVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid firmware version: %s < %s",
				version,
				MinOperatorVersion,
			),
		)
	}

	return nil
}

func getOperatorVersion(opVer *OperatorVersion) (string, error) {
	dynamic := dynamic.NewForConfigOrDie(opVer.client.RestConfig)

	resourceId := schema.GroupVersionResource{
		Group:    opVer.group,
		Version:  opVer.version,
		Resource: opVer.resource,
	}
	list, _ := dynamic.Resource(resourceId).Namespace(opVer.namespace).
		List(context.Background(), metav1.ListOptions{})

	for _, item := range list.Items {
		value := item.Object["spec"]
		crd := &CSV{}
		marsh, _ := json.Marshal(value)
		json.Unmarshal(marsh, crd)
		if crd.DisplayName == "PTP Operator" {
			return crd.Version, nil
		}
	}
	return "", errors.New("failed to find new")
}

func (opVer *OperatorVersion) GetID() string {
	return ptpOperatorVersionID
}

func (opVer *OperatorVersion) GetData() any { //nolint:ireturn // data will very for each validation
	return opVer
}

func NewOperatorVersion(client *clients.Clientset) *OperatorVersion {
	return &OperatorVersion{
		group:     "operators.coreos.com",
		version:   "v1alpha1",
		resource:  "clusterserviceversions",
		namespace: "openshift-ptp",
		client:    client,
	}
}
