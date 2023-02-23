package clients

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

// ContainerContext encapsulates the context in which a command is run; the namespace, pod, and container.
type ContainerContext struct {
	namespace     string
	podName       string
	containerName string
}

func NewContainerContext(namespace, podName, containerName string) ContainerContext {
	return ContainerContext{
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,
	}
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

type Command interface {
	ExecCommandContainer(ContainerContext, string) (string, string, error)
}

// ExecCommand runs command in a container and returns output buffers
func (clientsholder *Clientset) ExecCommandContainer(ctx ContainerContext, command []string) (stdout, stderr string, err error) {
	// commandStr := []string{"sh", "-c"}
	// commandStr = append(commandStr, command...)
	commandStr := command
	var buffOut bytes.Buffer
	var buffErr bytes.Buffer
	log.Debug(fmt.Sprintf("execute command on ns=%s, pod=%s container=%s, cmd: %s", ctx.GetNamespace(), ctx.GetPodName(), ctx.GetContainerName(), strings.Join(commandStr, " ")))
	req := clientsholder.K8sClient.CoreV1().RESTClient().
		Post().
		Namespace(ctx.GetNamespace()).
		Resource("pods").
		Name(ctx.GetPodName()).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: ctx.GetContainerName(),
			Command:   commandStr,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(clientsholder.RestConfig, "POST", req.URL())
	if err != nil {
		logrus.Error(err)
		return stdout, stderr, err
	}
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &buffOut,
		Stderr: &buffErr,
	})
	stdout, stderr = buffOut.String(), buffErr.String()
	if err != nil {
		log.Error(err)
		log.Error(req.URL())
		log.Error("command: ", command)
		log.Error("stderr: ", stderr)
		log.Error("stdout: ", stdout)
		return stdout, stderr, err
	}
	return stdout, stderr, err
}
