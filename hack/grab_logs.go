// SPDX-License-Identifier: GPL-2.0-or-later

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	ptpContainers = []string{"gpsd", "linuxptp-daemon-container"}

	kubeconfigPath string
	outputPath     string
	logWindow      string
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func buildCommand(containerName string) *exec.Cmd {
	params := []string{"logs", "--kubeconfig", kubeconfigPath, "--namespace=openshift-ptp", "daemonset/linuxptp-daemon", "-c", containerName} //nolint:lll //easier to read unwrapped
	if logWindow != "" {
		params = append(params, "--since", logWindow)
	}

	cmd := exec.Command("oc", params...)
	return cmd
}

func buildFilename(name string, timestamp time.Time) string {
	return fmt.Sprintf("%s/%s-%s", outputPath, name, timestamp.Format("060102T150405"))
}

func getLogsForContainer(containerName string, start time.Time) {
	cmd := buildCommand(containerName)
	log.Printf("Running command: %s", cmd)
	outputfile, err := os.Create(buildFilename(containerName, start))
	check(err)
	log.Printf("Outputting to file: %v", outputfile.Name())
	defer outputfile.Close()
	cmd.Stdout = outputfile

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		panic(err)
	}
}

func main() {
	flag.StringVar(&kubeconfigPath, "k", "", "Path to kubeconfig. Required.")
	flag.StringVar(
		&outputPath,
		"o",
		".",
		"Optional. Specify the output directory. Target must exist. Defaults to working directory.",
	)
	flag.StringVar(
		&logWindow,
		"since",
		"",
		"Optional. Only get logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs if omitted.",
	)
	flag.Parse()

	if kubeconfigPath == "" {
		log.Println("Kubeconfig path (-k) is Required")
		flag.Usage()
		os.Exit(1)
	}

	var wg sync.WaitGroup //nolint:varnamelen // `wg` is a common abbreviation for waitgroup.
	wg.Add(len(ptpContainers))

	now := time.Now()
	for i, s := range ptpContainers {
		log.Printf("Starting goroutine %d for %s", i, s)
		go func(s string) {
			defer wg.Done()
			getLogsForContainer(s, now)
		}(s)
	}
	wg.Wait()
}
