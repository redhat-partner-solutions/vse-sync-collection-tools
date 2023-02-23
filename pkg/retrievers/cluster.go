package retrievers

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UnknownClusterVersion    = "UnknownClusterVersion"
	apiserverClusterOperator = "openshift-apiserver"
)

// GetClusterVersion retrieves the version of the "openshift-apiserver" cluster operator
// as a proxy for the version of OCP.
// If no version can be established, will return a `version` of "UnknownClusterVersion" along with an error.
func GetClusterVersion(clientSet *clients.Clientset) (version string, err error) {
	version = UnknownClusterVersion
	clusterOperator, err := clientSet.OcpClient.ClusterOperators().Get(context.TODO(), apiserverClusterOperator, metav1.GetOptions{})
	if err != nil {
		return
	}
	err = fmt.Errorf("could not establish version", apiserverClusterOperator)
	log.Debugf("ClusterOperator versions: %v", clusterOperator.Status.Versions)
	for _, ver := range clusterOperator.Status.Versions {
		if ver.Name == apiserverClusterOperator {
			version = ver.Version
			err = nil
			break
		}
	}
	return
}
