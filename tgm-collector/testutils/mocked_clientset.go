// SPDX-License-Identifier: GPL-2.0-or-later

package testutils

import (
	"net/url"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakeK8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
)

const kubeconfigPath string = "test_files/kubeconfig"

// Returns a clientset where K8sClient and K8sRestClient are faked
func GetMockedClientSet(k8APIObjects ...runtime.Object) *clients.Clientset {
	clients.ClearClientSet()
	clientset, err := clients.GetClientset(kubeconfigPath)
	if err != nil {
		panic("Failed to get clientset")
	}
	fakeK8sClient := fakeK8s.NewSimpleClientset(k8APIObjects...)

	config := rest.ClientContentConfig{
		GroupVersion: schema.GroupVersion{Version: "v1"},
	}
	fakeRestClient, err := rest.NewRESTClient(&url.URL{}, "", config, nil, nil)
	if err != nil {
		panic("Failed to create rest client")
	}
	clientset.K8sClient = fakeK8sClient
	clientset.K8sRestClient = fakeRestClient
	return clientset
}
