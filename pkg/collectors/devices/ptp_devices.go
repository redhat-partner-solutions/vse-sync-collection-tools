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

package devices

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type PTPDeviceInfo struct {
	VendorID string
	DeviceID string
	TtyGNSS  string
}

type DevDPLLInfo struct {
	State  string
	Offset string
}

func GetPTPDeviceInfo(interfaceName string, ctx clients.ContainerContext) (devInfo PTPDeviceInfo) {
	// expecting a string like "../../../0000:86:00.0" here, just keep the last path section with filepath.Base
	busID := commandWithPostprocessFunc(ctx, filepath.Base, []string{
		"readlink", "/sys/class/net/" + interfaceName + "/device",
	})

	devInfo.TtyGNSS = "/dev/" + busToGNSS(busID)
	log.Debugf("got busID for %s:  %s", interfaceName, busID)
	log.Debugf("got tty for %s:  %s", interfaceName, devInfo.TtyGNSS)

	// expecting a string like 0x1593
	devInfo.DeviceID = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/device",
	})
	log.Debugf("got deviceID for %s:  %s", interfaceName, devInfo.DeviceID)

	// expecting a string like 0x8086
	devInfo.VendorID = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/vendor",
	})
	log.Debugf("got vendorID for %s:  %s", interfaceName, devInfo.VendorID)
	return
}

// transform a bus ID to an expected GNSS TTY name.
// e.g. "0000:86:00.0" -> "ttyGNSS_8600", "0000:51:02.1" -> "ttyGNSS_5102"
func busToGNSS(busID string) (string) {
	log.Debugf("convert %s to GNSS tty", busID)
	parts := strings.Split(busID, ":")
	ttyGNSS := parts[1] + strings.Split(parts[2], ".")[0]
	return "ttyGNSS_" + ttyGNSS
}

func commandWithPostprocessFunc(ctx clients.ContainerContext, cleanupFunc func(string) string, command []string) (result string) {
	clientset := clients.GetClientset()
	stdout, _, err := clientset.ExecCommandContainer(ctx, command)
	if err != nil {
		log.Errorf("command in container failed unexpectedly. context: %v", ctx)
		log.Errorf("command in container failed unexpectedly. command: %v", command)
		log.Errorf("command in container failed unexpectedly. error: %v", err)
		return ""
	}
	return cleanupFunc(stdout)
}

// Read lines from the ttyGNSS of the passed devInfo.
func ReadTtyGNSS(ctx clients.ContainerContext, devInfo PTPDeviceInfo, lines, timeoutSeconds int) string {
	return commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"timeout", strconv.Itoa(timeoutSeconds), "head", "-n", strconv.Itoa(lines), devInfo.TtyGNSS,
	})
}

// GetDevDPLLInfo returns the device DPLL info for an interface.
func GetDevDPLLInfo(ctx clients.ContainerContext, interfaceName string) (dpllInfo DevDPLLInfo) {
	dpllInfo.State = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_state",
	})
	dpllInfo.Offset = commandWithPostprocessFunc(ctx, strings.TrimSpace, []string{
		"cat", "/sys/class/net/" + interfaceName + "/device/dpll_1_offset",
	})
	return
}

// Write Logs in file. This will write a lot of data, so bufio.Writer will be recommended
func writeLogs(reader *bufio.Reader, writer io.Writer, timeout time.Duration) {
	if _, ok := writer.(*bufio.Writer); !ok {
		writer = bufio.NewWriter(writer)
	}
	for start := time.Now(); time.Since(start) < timeout; {
		str, readErr := reader.ReadString('\n')
		if readErr == io.EOF {
			break
		}
		_, err := writer.Write([]byte(str))
		if err != nil {
			return
		}
	}
}

// GetPtpLogs2File writes in a file the logs for a given pod move to ptplogs
func GetPtpDeviceLogsToFile(ctx clients.ContainerContext, timeout time.Duration, filename string) error {
	clientset := clients.GetClientset()
	// if the file does not exist, create it
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("could not open file for logging: %w", err)
	}
	defer file.Close()
	// get the logs
	logOptions := corev1.PodLogOptions{
		Container: ctx.GetContainerName(),
		Follow:    true,
	}
	logRequest := clientset.K8sClient.CoreV1().Pods(ctx.GetNamespace()).GetLogs(ctx.GetPodName(), &logOptions)
	stream, err := logRequest.Stream(context.TODO())
	defer stream.Close()
	if err != nil {
		return fmt.Errorf("could not retrieve log in ns=%s pod=%s, container=%s, err=%w", ctx.GetNamespace(), ctx.GetPodName(), ctx.GetContainerName(), err)
	}
	reader := bufio.NewReader(stream)
	writer := io.Writer(file)
	writeLogs(reader, writer, timeout)
	return nil
}
