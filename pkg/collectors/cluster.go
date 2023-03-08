package collectors

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	ocpconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UnknownClusterVersion    = "UnknownClusterVersion"
	ApiServerClusterOperator = "openshift-apiserver"
)

// GetClusterVersion retrieves the version of the "openshift-apiserver" cluster operator
// as a proxy for the version of OCP.
// If no version can be established, will return a `version` of "UnknownClusterVersion" along with an error.
func GetClusterVersion(ocpClient ocpconfigv1.ConfigV1Interface) (version string, err error) {
	version = UnknownClusterVersion
	// clusterOperator, err := clientSet.OcpClient.ClusterOperators().Get(context.TODO(), apiserverClusterOperator, metav1.GetOptions{})
	clusterOperator, err := ocpClient.ClusterOperators().Get(context.TODO(), ApiServerClusterOperator, metav1.GetOptions{})
	if err != nil {
		return
	}
	err = fmt.Errorf("could not establish version: %v", ApiServerClusterOperator)
	log.Debugf("ClusterOperator versions: %v", clusterOperator.Status.Versions)
	for _, ver := range clusterOperator.Status.Versions {
		if ver.Name == ApiServerClusterOperator {
			version = ver.Version
			err = nil
			break
		}
	}
	return
}
