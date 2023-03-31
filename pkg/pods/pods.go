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

package pods

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

// Builder provides a struct for pod object from the cluster and a pod definition.
type Builder struct {
	// Pod definition, used to create the pod object.
	Definition *v1.Pod
	// Created pod object.
	Object *v1.Pod
	// api clients to interact with the cluster.
	apiClients *clients.Clientset
	// Used to store latest error message upon defining or mutating pod definition.
	errorMsg string
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(name, nsname, image string, args []string) *Builder {
	log.Infof(
		"Initializing new pod structure with the following params: "+
			"name: %s, namespace: %s, image: %s",
		name, nsname, image)

	clientset := clients.GetClientset()

	builder := &Builder{
		apiClients: clientset,
		Definition: getDefinition(name, nsname, image, args),
	}

	if name == "" {
		log.Infof("The name of the pod is empty")

		builder.errorMsg = "pod's name is empty"
	}

	if nsname == "" {
		log.Infof("The namespace of the pod is empty")

		builder.errorMsg = "namespace's name is empty"
	}

	if image == "" {
		log.Infof("The image of the pod is empty")

		builder.errorMsg = "pod's image is empty"
	}

	return builder
}

// Create makes a pod according to the pod definition and stores the created object in the associated builder.
func (builder *Builder) Create() (*Builder, error) {
	log.Infof("Creating pod %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if builder.errorMsg != "" {
		return nil, fmt.Errorf("%s", builder.errorMsg)
	}

	var err error
	if !builder.Exists() {
		log.Infof("Nex is Creating pod %s in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
		builder.Object, err = builder.apiClients.K8sClient.CoreV1().Pods(builder.Definition.Namespace).Create(
			context.TODO(), builder.Definition, metaV1.CreateOptions{})
		fmt.Printf("the output of err: %s\n", err)

	}

	return builder, err
}

// Exists checks whether the given pod in the namespace exists.
func (builder *Builder) Exists() bool {
	return false
}

// create pod definition
func getDefinition(name, nsName, image string, args []string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			GenerateName: name,
			Namespace:    nsName},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: pointer.Int64Ptr(0),
			Containers: []v1.Container{
				{
					Name:  name + "-container",
					Image: image,
					Args:  args,
				},
			},
		},
	}
}
