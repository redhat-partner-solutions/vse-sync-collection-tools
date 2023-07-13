package vaildations

import (
	"context"
	"fmt"

	// "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const ptpOperatorVersionID = "PTP Operactor Version is valid"

type OperatorVersion struct {
	Version   string             `json:"version"`
	group     string             `json:"-"`
	version   string             `json:"-"`
	resource  string             `json:"-"`
	namespace string             `json:"-"`
	client    *clients.Clientset `json:"-"`
}

func (opVer *OperatorVersion) Verify() error {
	dynamic := dynamic.NewForConfigOrDie(opVer.client.RestConfig)

	resourceId := schema.GroupVersionResource{
		Group:    opVer.group,
		Version:  opVer.version,
		Resource: opVer.resource,
	}
	list, err := dynamic.Resource(resourceId).Namespace(opVer.namespace).
		List(context.Background(), metav1.ListOptions{})

	fmt.Println(err)
	fmt.Println(list)
	// v1alpha1.ClusterServiceVersion{}
	return nil
}

func (opVer *OperatorVersion) GetID() string {
	return ptpOperatorVersionID
}

func (opVer *OperatorVersion) GetData() any { //nolint:ireturn // data will very for each validation
	return opVer
}

func NewOperatorVersion(client *clients.Clientset) *OperatorVersion {
	return &OperatorVersion{
		group:     "",
		version:   "v1alpha1",
		resource:  "clusterserviceversions",
		namespace: "openshift-ptp",
		client:    client,
	}
}
