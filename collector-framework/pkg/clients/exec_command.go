// SPDX-License-Identifier: GPL-2.0-or-later

package clients

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	startTimeout    = 2 * time.Minute
	deletionTimeout = 10 * time.Minute
)

type APIContext interface {
	GetNamespace() string
	GetPodName() string
	GetContainerName() string
	GetClientset() *Clientset
}

type ExecContext interface {
	ExecCommand([]string) (string, string, error)
	ExecCommandStdIn([]string, bytes.Buffer) (string, string, error)
}

var NewSPDYExecutor = remotecommand.NewSPDYExecutor

// ContainerExecContext encapsulates the context in which a command is run; the namespace, pod, and container.
type ContainerExecContext struct {
	clientset     *Clientset
	namespace     string
	podName       string
	containerName string
	podNamePrefix string
}

func (c *ContainerExecContext) Refresh() error {
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
) (*ContainerExecContext, error) {
	podName, err := clientset.FindPodNameFromPrefix(namespace, podNamePrefix)
	if err != nil {
		return &ContainerExecContext{}, err
	}
	ctx := ContainerExecContext{
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,
		podNamePrefix: podNamePrefix,
		clientset:     clientset,
	}
	return &ctx, nil
}

func (c *ContainerExecContext) GetNamespace() string {
	return c.namespace
}

func (c *ContainerExecContext) GetPodName() string {
	return c.podName
}

func (c *ContainerExecContext) GetContainerName() string {
	return c.containerName
}

func (c *ContainerExecContext) GetClientset() *Clientset {
	return c.clientset
}

//nolint:lll,funlen // allow slightly long function definition and function length
func (c *ContainerExecContext) execCommand(command []string, buffInPtr *bytes.Buffer) (stdout, stderr string, err error) {
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
		if k8sErrors.IsNotFound(err) {
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
func (c *ContainerExecContext) ExecCommand(command []string) (stdout, stderr string, err error) {
	return c.execCommand(command, nil)
}

//nolint:lll // allow slightly long function definition
func (c *ContainerExecContext) ExecCommandStdIn(command []string, buffIn bytes.Buffer) (stdout, stderr string, err error) {
	return c.execCommand(command, &buffIn)
}

// ContainerExecContext encapsulates the context in which a command is run; the namespace, pod, and container.
type ContainerCreationExecContext struct {
	*ContainerExecContext
	labels                   map[string]string
	pod                      *corev1.Pod
	containerSecurityContext *corev1.SecurityContext
	containerImage           string
	command                  []string
	volumes                  []*Volume
	hostNetwork              bool
}

type Volume struct {
	VolumeSource corev1.VolumeSource
	Name         string
	MountPath    string
}

func (c *ContainerCreationExecContext) createPod() error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.podName,
			Namespace: c.namespace,
			Labels:    c.labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            c.containerName,
					Image:           c.containerImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
				},
			},
			HostNetwork: c.hostNetwork,
		},
	}
	if len(c.command) > 0 {
		pod.Spec.Containers[0].Command = c.command
	}
	if c.containerSecurityContext != nil {
		pod.Spec.Containers[0].SecurityContext = c.containerSecurityContext
	}
	if len(c.volumes) > 0 {
		volumes := make([]corev1.Volume, 0)
		volumeMounts := make([]corev1.VolumeMount, 0)

		for _, v := range c.volumes {
			volumes = append(volumes, corev1.Volume{Name: v.Name, VolumeSource: v.VolumeSource})
			pod.Spec.Volumes = volumes
			volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: v.Name, MountPath: v.MountPath})
			pod.Spec.Containers[0].VolumeMounts = volumeMounts
		}
	}

	pod, err := c.clientset.K8sClient.CoreV1().Pods(pod.Namespace).Create(
		context.TODO(),
		pod,
		metav1.CreateOptions{},
	)
	c.pod = pod
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}
	return nil
}

func (c *ContainerCreationExecContext) listPods(options *metav1.ListOptions) (*corev1.PodList, error) {
	pods, err := c.clientset.K8sClient.CoreV1().Pods(c.pod.Namespace).List(
		context.TODO(),
		*options,
	)
	if err != nil {
		return pods, fmt.Errorf("failed to find pods: %s", err.Error())
	}
	return pods, nil
}

func (c *ContainerCreationExecContext) refeshPod() error {
	pods, err := c.listPods(&metav1.ListOptions{
		FieldSelector:   fields.OneTermEqualSelector("metadata.name", c.podName).String(),
		ResourceVersion: c.pod.ResourceVersion,
	})
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("failed to find pod: %s", c.podName)
	}
	c.pod = &pods.Items[0]

	return nil
}

func (c *ContainerCreationExecContext) isPodRunning() (bool, error) {
	err := c.refeshPod()
	if err != nil {
		return false, err
	}
	if c.pod.Status.Phase == corev1.PodRunning {
		return true, nil
	}
	return false, nil
}

func (c *ContainerCreationExecContext) waitForPodToStart() error {
	start := time.Now()
	for time.Since(start) <= startTimeout {
		running, err := c.isPodRunning()
		if err != nil {
			return err
		}
		if running {
			return nil
		}
		time.Sleep(time.Microsecond)
	}
	return errors.New("timed out waiting for pod to start")
}

func (c *ContainerCreationExecContext) CreatePodAndWait() error {
	var err error
	running := false
	if c.pod != nil {
		running, err = c.isPodRunning()
		if err != nil {
			return err
		}
	}
	if !running {
		err := c.createPod()
		if err != nil {
			return err
		}
	}
	return c.waitForPodToStart()
}

func (c *ContainerCreationExecContext) deletePod() error {
	deletePolicy := metav1.DeletePropagationForeground
	err := c.clientset.K8sClient.CoreV1().Pods(c.pod.Namespace).Delete(
		context.TODO(),
		c.pod.Name,
		metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	return nil
}

func (c *ContainerCreationExecContext) waitForPodToDelete() error {
	start := time.Now()
	for time.Since(start) <= deletionTimeout {
		pods, err := c.listPods(&metav1.ListOptions{})
		if err != nil {
			return err
		}
		found := false
		for _, pod := range pods.Items { //nolint:gocritic // This isn't my object I can't use a pointer
			if pod.Name == c.podName {
				found = true
			}
		}
		if !found {
			return nil
		}
		time.Sleep(time.Microsecond)
	}
	return errors.New("pod has not terminated within the timeout")
}

func (c *ContainerCreationExecContext) DeletePodAndWait() error {
	err := c.deletePod()
	if err != nil {
		return err
	}
	return c.waitForPodToDelete()
}

func NewContainerCreationExecContext(
	clientset *Clientset,
	namespace, podName, containerName, containerImage string,
	labels map[string]string,
	command []string,
	containerSecurityContext *corev1.SecurityContext,
	hostNetwork bool,
	volumes []*Volume,
) *ContainerCreationExecContext {
	ctx := ContainerExecContext{
		namespace:     namespace,
		podNamePrefix: podName,
		podName:       podName,
		containerName: containerName,
		clientset:     clientset,
	}

	return &ContainerCreationExecContext{
		ContainerExecContext:     &ctx,
		containerImage:           containerImage,
		labels:                   labels,
		command:                  command,
		containerSecurityContext: containerSecurityContext,
		hostNetwork:              hostNetwork,
		volumes:                  volumes,
	}
}
