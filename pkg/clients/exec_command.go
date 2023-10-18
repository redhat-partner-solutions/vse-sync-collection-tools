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

//nolint:lll,funlen // allow slightly long function definition and function length
func (c *ContainerContext) execCommand(command []string, buffInPtr *bytes.Buffer) (stdout, stderr string, err error) {
	commandStr := command
	var buffOut bytes.Buffer
	var buffErr bytes.Buffer

	useBuffIn := buffInPtr != nil

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
			Stdin:     useBuffIn,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := NewSPDYExecutor(c.clientset.RestConfig, "POST", req.URL())
	if err != nil {
		log.Debug(err)
		return stdout, stderr, fmt.Errorf("error setting up remote command: %w", err)
	}

	var streamOptions remotecommand.StreamOptions

	if useBuffIn {
		streamOptions = remotecommand.StreamOptions{
			Stdin:  buffInPtr,
			Stdout: &buffOut,
			Stderr: &buffErr,
		}
	} else {
		streamOptions = remotecommand.StreamOptions{
			Stdout: &buffOut,
			Stderr: &buffErr,
		}
	}

	err = exec.StreamWithContext(context.TODO(), streamOptions)
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
		if useBuffIn {
			log.Debug("stdin: ", buffInPtr.String())
		}
		log.Debug("stderr: ", stderr)
		log.Debug("stdout: ", stdout)
		return stdout, stderr, fmt.Errorf("error running remote command: %w", err)
	}
	return stdout, stderr, nil
}

// ExecCommand runs command in a container and returns output buffers
//
//nolint:lll,funlen // allow slightly long function definition and allow a slightly long function
func (c *ContainerContext) ExecCommandContainer(command []string) (stdout, stderr string, err error) {
	return c.execCommand(command, nil)
}

//nolint:lll // allow slightly long function definition
func (c *ContainerContext) ExecCommandContainerStdIn(command []string, buffIn bytes.Buffer) (stdout, stderr string, err error) {
	return c.execCommand(command, &buffIn)
}
