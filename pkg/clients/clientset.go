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

package clients

import (
	"fmt"
	"time"

	ocpconfig "github.com/openshift/client-go/config/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// A Clientset contains clients for the different k8s API groups in one place
type Clientset struct {
	RestConfig    *rest.Config
	DynamicClient dynamic.Interface
	OcpClient     ocpconfig.Interface
	K8sClient     kubernetes.Interface
	ready         bool
}

var clientset = Clientset{}

// GetClientset returns the singleton clientset object.
func GetClientset(kubeconfigPaths ...string) *Clientset {
	if clientset.ready {
		return &clientset
	}

	if len(kubeconfigPaths) == 0 {
		log.Panic("must have at least one kubeconfig to initialise a new Clientset")
	}
	clientset, err := newClientset(kubeconfigPaths...)
	if err != nil {
		log.Panic("Failed to create k8s clients holder: ", err)
	}
	return clientset
}

// newClientset will initialise the singleton clientset using provided kubeconfigPath
func newClientset(kubeconfigPaths ...string) (*Clientset, error) {
	log.Infof("creating new Clientset from %v", kubeconfigPaths)
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	loadingRules.Precedence = kubeconfigPaths // This means it will not load the value from $KUBECONFIG
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)
	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	var err error
	clientset.RestConfig, err = kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate rest config: %w", err)
	}

	DefaultTimeout := 10 * time.Second
	clientset.RestConfig.Timeout = DefaultTimeout

	clientset.DynamicClient, err = dynamic.NewForConfig(clientset.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate dynamic client (unstructured/dynamic): %w", err)
	}
	clientset.K8sClient, err = kubernetes.NewForConfig(clientset.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate k8sclient: %w", err)
	}
	// create the oc client
	clientset.OcpClient, err = ocpconfig.NewForConfig(clientset.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate ocClient: %w", err)
	}

	clientset.ready = true
	return &clientset, nil
}
