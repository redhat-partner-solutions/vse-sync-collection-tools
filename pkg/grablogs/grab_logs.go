// SPDX-License-Identifier: GPL-2.0-or-later

package grablogs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

var (
	ptpContainers = []string{"linuxptp-daemon-container"}
)

func buildCommand(containerName, kubeconfigPath, logWindow string) *exec.Cmd {
	params := []string{"logs", "--kubeconfig", kubeconfigPath, "--namespace=openshift-ptp", "daemonset/linuxptp-daemon", "-c", containerName} //nolint:lll //easier to read unwrapped
	if logWindow != "" {
		params = append(params, "--since", logWindow)
	}

	cmd := exec.Command("oc", params...)
	return cmd
}

func buildFilename(outputPath, name string, timestamp time.Time) string {
	return fmt.Sprintf("%s/%s-%s", outputPath, name, timestamp.Format("060102T150405"))
}

func getLogsForContainer(cmd *exec.Cmd, filename string) {
	log.Infof("Running command: %s", cmd)
	outputfile, err := os.Create(filename)
	utils.IfErrorExitOrPanic(err)
	log.Infof("Outputting to file: %v", outputfile.Name())
	defer outputfile.Close()
	cmd.Stdout = outputfile

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Run()
	utils.IfErrorExitOrPanic(err)
}

func GrabLogs(kubeconfigPath, logWindow, outputPath string) {
	var wg sync.WaitGroup //nolint:varnamelen // `wg` is a common abbreviation for waitgroup.
	wg.Add(len(ptpContainers))

	now := time.Now()
	for i, containerName := range ptpContainers {
		log.Infof("Starting goroutine %d for %s", i, containerName)
		go func(containerName string) {
			defer wg.Done()
			cmd := buildCommand(containerName, kubeconfigPath, logWindow)
			filename := buildFilename(outputPath, containerName, now)
			getLogsForContainer(cmd, filename)
		}(containerName)
	}
	wg.Wait()
}
