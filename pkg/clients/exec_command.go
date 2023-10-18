// SPDX-License-Identifier: GPL-2.0-or-later

package clients

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

var NewSPDYExecutor = remotecommand.NewSPDYExecutor

// ContainerContext encapsulates the context in which a command is run; the namespace, pod, and container.
type ContainerContext struct {
	clientset     *Clientset
	namespace     string
	podName       string
	containerName string
	podNamePrefix string
}

func (c *ContainerContext) Refresh() error {
	newPodname, err := c.clientset.FindPodNameFromPrefix(c.namespace, c.podNamePrefix)
	if err != nil {
		return err
	}
	c.podName = newPodname
	return nil
}

func NewContainerContext(
	clientset *Clientset,
	namespace, podNamePrefix, containerName string,
) (ContainerContext, error) {
	podName, err := clientset.FindPodNameFromPrefix(namespace, podNamePrefix)
	if err != nil {
		return ContainerContext{}, err
	}
	ctx := ContainerContext{
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,
		podNamePrefix: podNamePrefix,
		clientset:     clientset,
	}
	return ctx, nil
}

func (c *ContainerContext) GetNamespace() string {
	return c.namespace
}

func (c *ContainerContext) GetPodName() string {
	return c.podName
}

func (c *ContainerContext) GetContainerName() string {
	return c.containerName
}

// ExecCommand runs command in a container and returns output buffers
//
//nolint:lll,funlen // allow slightly long function definition and allow a slightly long function
func (c *ContainerContext) ExecCommandContainer(command []string) (stdout, stderr string, err error) {
	commandStr := command
	var buffOut bytes.Buffer
	var buffErr bytes.Buffer
	log.Debugf(
		"execute command on ns=%s, pod=%s container=%s, cmd: %s\n",
		c.GetNamespace(),
		c.GetPodName(),
		c.GetContainerName(),
		strings.Join(commandStr, " "),
	)
	req := c.clientset.K8sRestClient.Post().
		Namespace(c.GetNamespace()).
		Resource("pods").
		Name(c.GetPodName()).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: c.GetContainerName(),
			Command:   commandStr,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := NewSPDYExecutor(c.clientset.RestConfig, "POST", req.URL())
	if err != nil {
		log.Debug(err)
		return stdout, stderr, fmt.Errorf("error setting up remote command: %w", err)
	}

	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &buffOut,
		Stderr: &buffErr,
	})
	stdout, stderr = buffOut.String(), buffErr.String()
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("Pod %s was not found, likely restarted so refreshing context", c.GetPodName())
			refreshErr := c.Refresh()
			if refreshErr != nil {
				log.Debug("Failed to refresh container context", refreshErr)
			}
		}
		log.Debug(err)
		log.Debug(req.URL())
		log.Debug("command: ", command)
		log.Debug("stderr: ", stderr)
		log.Debug("stdout: ", stdout)
		return stdout, stderr, fmt.Errorf("error running remote command: %w", err)
	}
	return stdout, stderr, nil
}

//nolint:lll // allow slightly long function definition
func (c *ContainerContext) ExecCommandContainerStdIn(command []string, buffIn bytes.Buffer) (stdout, stderr string, err error) {
	commandStr := command
	var buffOut bytes.Buffer
	var buffErr bytes.Buffer
	log.Debugf(
		"execute command on ns=%s, pod=%s container=%s, cmd: %s",
		c.GetNamespace(),
		c.GetPodName(),
		c.GetContainerName(),
		strings.Join(commandStr, " "),
	)
	req := c.clientset.K8sRestClient.Post().
		Namespace(c.GetNamespace()).
		Resource("pods").
		Name(c.GetPodName()).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: c.GetContainerName(),
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := NewSPDYExecutor(c.clientset.RestConfig, "POST", req.URL())
	if err != nil {
		log.Debug(err)
		return stdout, stderr, fmt.Errorf("error setting up remote command: %w", err)
	}

	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  &buffIn,
		Stdout: &buffOut,
		Stderr: &buffErr,
	})
	stdin, stdout, stderr := buffIn.String(), buffOut.String(), buffErr.String()
	if err != nil {
		log.Debug(err)
		log.Debug(req.URL())
		log.Debug("command: ", command)
		log.Debug("stdin: ", stdin)
		log.Debug("stderr: ", stderr)
		log.Debug("stdout: ", stdout)
		return stdout, stderr, fmt.Errorf("error running remote command: %w", err)
	}
	return stdout, stderr, nil
}
