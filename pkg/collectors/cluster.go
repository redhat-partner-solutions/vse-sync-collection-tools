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

package collectors

import (
	"context"
	"fmt"

	ocpconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UnknownClusterVersion    = "UnknownClusterVersion"
	APIServerClusterOperator = "openshift-apiserver"
)

// GetClusterVersion retrieves the version of the "openshift-apiserver" cluster operator
// as a proxy for the version of OCP.
// If no version can be established, will return a `version` of "UnknownClusterVersion" along with an error.
func GetClusterVersion(ocpClient ocpconfigv1.ConfigV1Interface) (version string, err error) {
	version = UnknownClusterVersion
	clusterOperator, err := ocpClient.ClusterOperators().Get(context.TODO(), APIServerClusterOperator, metav1.GetOptions{})
	if err != nil {
		return
	}
	err = fmt.Errorf("could not establish version: %v", APIServerClusterOperator)
	log.Debugf("ClusterOperator versions: %v", clusterOperator.Status.Versions)
	for _, ver := range clusterOperator.Status.Versions {
		if ver.Name == APIServerClusterOperator {
			version = ver.Version
			err = nil
			break
		}
	}
	return
}
