// SPDX-License-Identifier: GPL-2.0-or-later

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	expect "github.com/Netflix/go-expect"
	"github.com/icza/backscanner"
	log "github.com/sirupsen/logrus"
)

const (
	timeout          = 5 * time.Second
	podSafetyWait    = 5 * time.Second // generous sleep so we don't create too many debug pods and break.
	defaultLogWindow = 2000 * time.Second
)

var (
	promptRE  = regexp.MustCompile(`sh-4\.4#`) // line has ANSI escape characters: \r\x1b[Ksh-4.4#
	logfileRE = regexp.MustCompile(`(\d)\.log(\..*)*`)
)

var (
	kubeconfigPath string
	outputPath     string
	tmpOutputPath  string
	logWindow      time.Duration
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func buildDebugCommand(nodeName string) []string {
	nodepath := "node/" + nodeName
	return []string{"oc", "debug", "--kubeconfig", kubeconfigPath, "--quiet", nodepath}
}

func getNodeName() string {
	nodeNameCommand := []string{"get", "nodes", "--kubeconfig", kubeconfigPath, "-o=jsonpath={.items[0].metadata.name}"}
	cmd := exec.Command("oc", nodeNameCommand...)
	log.Debug(cmd)
	var nodeName strings.Builder
	cmd.Stdout = &nodeName
	check(cmd.Run())
	log.Debug(nodeName.String())
	return nodeName.String()
}

func pickLogFiles(filePaths []string) []string {
	maxLogNumber := 0
	var candidateFiles []string //nolint:prealloc // performance not the goal here
	for _, filePath := range filePaths {
		_, fileName := path.Split(filePath)
		logNumber, err := strconv.Atoi(strings.SplitN(fileName, ".", 2)[0]) //nolint:gomnd // making constant does not improve code.
		if err != nil {
			continue // skip file
		}
		if logNumber < maxLogNumber {
			continue // only get files from the biggest log number
		} else if logNumber > maxLogNumber {
			maxLogNumber = logNumber
			candidateFiles = nil
		}
		candidateFiles = append(candidateFiles, filePath)
	}
	return candidateFiles
}

func downloadLogs(nodeName string, logPaths []string) []string {
	var createdFiles []string //nolint:prealloc // performance not the goal here
	var stderr strings.Builder
	for _, filePath := range logPaths {
		log.Infof("Getting logs from remote file: %s", filePath)
		_, fileName := path.Split(filePath)
		args := append(buildDebugCommand(nodeName), "--")
		if strings.HasSuffix(fileName, ".gz") {
			args = append(args, []string{"gzip", "-d", filePath, "--stdout"}...)
			fileName = strings.ReplaceAll(fileName, ".gz", "")
		} else {
			args = append(args, []string{"cat", filePath}...)
		}
		cmd := exec.Command(args[0], args[1:]...) //nolint:gosec // accepting security risk.

		localFile := fmt.Sprintf("%s/%s", tmpOutputPath, fileName)
		createdFiles = append(createdFiles, localFile)
		outputFile, err := os.Create(localFile)
		check(err)
		cmd.Stdout = outputFile
		cmd.Stderr = &stderr

		if err = cmd.Run(); err != nil {
			log.Error(stderr.String())
			check(err)
		}
		err = outputFile.Close()
		check(err)
		stderr.Reset()
		time.Sleep(podSafetyWait)
	}
	return createdFiles
}

func getLogsFileNamesForNode(nodeName string) []string {
	expecter, err := expect.NewConsole(expect.WithStdout(os.Stdout), expect.WithDefaultTimeout(timeout))
	check(err)
	defer expecter.Close()

	args := buildDebugCommand(nodeName)
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec // accepting security risk.
	cmd.Stdin = expecter.Tty()
	cmd.Stdout = expecter.Tty()
	cmd.Stderr = expecter.Tty()

	err = cmd.Start()
	check(err)
	defer cmd.Process.Kill() //nolint:errcheck // continue regardless of error.

	_, err = expecter.Expect(expect.Regexp(promptRE))
	check(err)

	_, err = expecter.Send("find /host/var/log/pods/openshift-ptp_linuxptp-daemon-*/linuxptp-daemon-container/*log*\n")
	check(err)
	logFiles, _ := expecter.Expect(expect.Regexp(promptRE)) // Just wait for timeout here as the prompt is full of invisible characters which makes regex a nightmare.

	var candidateFiles []string
	for _, line := range strings.Split(logFiles, "\n") {
		trimmedLine := strings.TrimSpace(line)
		log.Debugf("Got Line:'%q'", trimmedLine)
		if match := logfileRE.MatchString(trimmedLine); match {
			candidateFiles = append(candidateFiles, trimmedLine)
		}
	}
	logsToRetrieve := pickLogFiles(candidateFiles)
	return logsToRetrieve
}

func processLogLine(line string) (time.Time, string) { //nolint:gocritic //names not needed
	splitLine := strings.SplitN(line, " ", 2) //nolint:gomnd // making constant does not improve code.
	timestamp, err := time.Parse(time.RFC3339, splitLine[0])
	if err != nil {
		return time.Now(), ""
	}
	// 2023-09-08T11:03:22.315992414+00:00 stdout F ts2phc[1374422.785]: [ts2phc.0.config] nmea sentence: GNGGA,110322.00,4233.01417,N,07112.87799,W,1,12,0.48,58.2,M,-33.0,M,,
	// becomes
	// ts2phc[1374422.785]: [ts2phc.0.config] nmea sentence: GNGGA,110322.00,4233.01417,N,07112.87799,W,1,12,0.48,58.2,M,-33.0,M,,
	return timestamp, line[45:]
}

func scanFileInReverseToTime(fileName string, targetTime time.Time, output *os.File) (foundTime bool) {
	foundTime = false
	f, err := os.Open(fileName)
	check(err)
	defer f.Close()
	finfo, err := f.Stat()
	check(err)
	reverseReader := backscanner.New(f, int(finfo.Size()))
	for {
		line, _, err := reverseReader.Line()
		if errors.Is(err, io.EOF) {
			log.Infof("Finished processing file: %s", fileName)
			break
		} else if err != nil {
			log.Errorf("Unexpected error reading file %s: %s", fileName, err)
			break
		}
		ts, logLine := processLogLine(line)
		if ts.Before(targetTime) {
			foundTime = true
			return
		}
		if line != "" {
			_, _ = output.WriteString(logLine + "\n")
		}
	}
	return
}

func buildSingleLog(fromFiles []string, since time.Time) {
	reverseFilename := fmt.Sprintf("%s/logs.gol", tmpOutputPath)
	reverseFinalLogFile, err := os.Create(reverseFilename)
	check(err)
	defer reverseFinalLogFile.Close()

	// assuming order of fromFiles is consistent due to naming scheme: first item is most recent file, the rest are in ascending time order
	reverseTimeOrder := []string{fromFiles[0]}
	for i := len(fromFiles) - 1; i > 0; i-- {
		reverseTimeOrder = append(reverseTimeOrder, fromFiles[i])
	}
	foundTime := false
	// work over files in descending time order, and lines in each file from end to start, until we find a line before since time.
	for _, f := range reverseTimeOrder {
		log.Infof("Processing file %s", f)
		foundTime = scanFileInReverseToTime(f, since, reverseFinalLogFile)
		if foundTime {
			log.Infof("Found sinceTime in file %s", f)
			break
		}
	}
	if !foundTime {
		log.Warn("Processed all log files without satisfying since duration")
	}

	cmd := exec.Command("tac", reverseFilename)
	finalLogFile, err := os.Create(outputPath)
	check(err)
	log.Debugf("%s > %s", cmd.Args, finalLogFile.Name())
	cmd.Stdout = finalLogFile
	log.Infof("Writing final logs to %s", finalLogFile.Name())
	err = cmd.Run()
	check(err)
}

func main() {
	flag.StringVar(&kubeconfigPath, "k", "", "Path to kubeconfig. Required.")
	flag.StringVar(&outputPath, "o", "", "Output file for final logs. Target must exist. Required")
	flag.DurationVar(
		&logWindow,
		"d",
		defaultLogWindow,
		fmt.Sprintf("Optional. Duration of logs to finally output. Defaults to %s.", defaultLogWindow),
	)
	flag.StringVar(
		&tmpOutputPath,
		"t",
		".",
		"Optional. Specify the temporary output directory. Target must exist. Defaults to working directory.",
	)
	flag.Parse()

	if kubeconfigPath == "" {
		log.Error("Kubeconfig path (-k) is Required")
		flag.Usage()
		os.Exit(1)
	}
	if outputPath == "" {
		log.Error("Output path (-o) is Required")
		flag.Usage()
		os.Exit(1)
	}

	sinceTime := time.Now().Add(-logWindow)
	log.Infof("Getting logs since %s", sinceTime)
	nodeName := getNodeName()
	logFiles := getLogsFileNamesForNode(nodeName)
	time.Sleep(podSafetyWait)
	localFiles := downloadLogs(nodeName, logFiles)
	buildSingleLog(localFiles, sinceTime)
}
