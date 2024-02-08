// SPDX-License-Identifier: GPL-2.0-or-later

package clients

import (
	"context"
	"fmt"
	"strings"
	"time"

	ocpconfig "github.com/openshift/client-go/config/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
)

// A Clientset contains clients for the different k8s API groups in one place
type Clientset struct {
	RestConfig      *rest.Config
	DynamicClient   dynamic.Interface
	OcpClient       ocpconfig.Interface
	K8sClient       kubernetes.Interface
	K8sRestClient   rest.Interface
	KubeConfigPaths []string
	ready           bool
}

var clientset = Clientset{}

// GetClientset returns the singleton clientset object.
func GetClientset(kubeconfigPaths ...string) (*Clientset, error) {
	if clientset.ready {
		return &clientset, nil
	}

	if len(kubeconfigPaths) == 0 {
		return nil, utils.NewMissingInputError(
			fmt.Errorf("must have at least one kubeconfig to initialise a new Clientset"),
		)
	}
	clientset, err := newClientset(kubeconfigPaths...)
	if err != nil {
		return nil, utils.NewMissingInputError(
			fmt.Errorf("failed to create k8s clients holder: %w", err),
		)
	}
	return clientset, nil
}

// newClientset will initialise the singleton clientset using provided kubeconfigPath
func newClientset(kubeconfigPaths ...string) (*Clientset, error) {
	log.Infof("creating new Clientset from %v", kubeconfigPaths)
	clientset.KubeConfigPaths = kubeconfigPaths
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

	clientset.K8sRestClient = clientset.K8sClient.CoreV1().RESTClient()
	clientset.ready = true
	return &clientset, nil
}

func ClearClientSet() {
	clientset = Clientset{}
}

func (clientsholder *Clientset) FindPodNameFromPrefix(namespace, prefix string) (string, error) {
	podList, err := clientsholder.K8sClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to getting pod list: %w", err)
	}
	podNames := make([]string, 0)

	for i := range podList.Items {
		hasPrefix := strings.HasPrefix(podList.Items[i].Name, prefix)
		isDebug := strings.HasSuffix(podList.Items[i].Name, "-debug")
		if hasPrefix && !isDebug {
			podNames = append(podNames, podList.Items[i].Name)
		}
	}

	switch len(podNames) {
	case 0:
		return "", fmt.Errorf("no pod with prefix %v found in namespace %v", prefix, namespace)
	case 1:
		return podNames[0], nil
	default:
		return "", fmt.Errorf("too many (%v) pods with prefix %v found in namespace %v", len(podNames), prefix, namespace)
	}
}
