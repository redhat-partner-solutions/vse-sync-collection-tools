package clients

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	ocpconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// A Clientset contains clients for the different k8s API groups in one place
type Clientset struct {
	RestConfig    *rest.Config
	DynamicClient dynamic.Interface
	OcpClient     ocpconfigv1.ConfigV1Interface
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
		return nil, fmt.Errorf("cannot instantiate rest config: %s", err)
	}

	DefaultTimeout := 10 * time.Second
	clientset.RestConfig.Timeout = DefaultTimeout

	clientset.DynamicClient, err = dynamic.NewForConfig(clientset.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate dynamic client (unstructured/dynamic): %s", err)
	}
	clientset.K8sClient, err = kubernetes.NewForConfig(clientset.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate k8sclient: %s", err)
	}
	// create the oc client
	clientset.OcpClient, err = ocpconfigv1.NewForConfig(clientset.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate ocClient: %s", err)
	}

	clientset.ready = true
	return &clientset, nil
}
