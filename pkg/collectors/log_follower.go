// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	lineChanLength      = 100
	lineDelim           = '\n'
	streamingBufferSize = 2000
	logPollInterval     = 2
	logFilePermissions  = 0666
)

var (
	followDuration = logPollInterval * time.Second
	keepDuration   = 10 * followDuration
	followTimeout  = 2 * followDuration
)

type LogsCollector struct {
	lastPoll           time.Time
	quit               chan os.Signal
	lines              chan string
	seenLines          map[time.Time][]string
	client             *clients.Clientset
	logsOutputFileName string
	wg                 sync.WaitGroup
	pollInterval       int
	lpLock             sync.RWMutex
	withTimeStamps     bool
	running            bool
	pruned             bool
}

const (
	LogsCollectorName = "Logs"
	LogsInfo          = "log-line"
)

func (logs *LogsCollector) GetPollInterval() int {
	return logs.pollInterval
}

func (logs *LogsCollector) IsAnnouncer() bool {
	return false
}

func (logs *LogsCollector) SetLastPool() {
	if !logs.lpLock.TryLock() {
		return
	}
	defer logs.lpLock.Unlock()
	logs.lastPoll = time.Now()
}

// Start sets up the collector so it is ready to be polled
func (logs *LogsCollector) Start() error {
	go logs.writeToLogFile()
	logs.running = true
	return nil
}

func (logs *LogsCollector) consumeLine(line string, writer io.StringWriter) {
	splitLine := strings.SplitN(line, " ", 2) //nolint:gomnd // moving this to a var would make the code less clear
	if len(splitLine) < 2 {                   //nolint:gomnd // moving this to a var would make the code less clear
		return
	}
	timestampPart := splitLine[0]
	lineContent := splitLine[1]
	timestamp, err := time.Parse(time.RFC3339, timestampPart)
	if err != nil {
		// This is not a value line something went wrong
		return
	}

	if seen, ok := logs.seenLines[timestamp]; ok {
		for _, seenContent := range seen {
			if seenContent == lineContent {
				// Already have this line skip it
				return
			}
		}
	}
	logs.seenLines[timestamp] = append(logs.seenLines[timestamp], lineContent)
	logs.pruned = false
	if logs.withTimeStamps {
		_, err = writer.WriteString(line + "\n")
	} else {
		_, err = writer.WriteString(lineContent + "\n")
	}
	if err != nil {
		log.Error(fmt.Errorf("failed to write log output to file"))
	}
}

func (logs *LogsCollector) writeToLogFile() {
	logs.wg.Add(1)
	defer logs.wg.Done()

	fileHandle, err := os.OpenFile(logs.logsOutputFileName, os.O_CREATE|os.O_WRONLY, logFilePermissions)
	utils.IfErrorExitOrPanic(err)
	defer fileHandle.Close()

	for {
		select {
		case <-logs.quit:
			// Consume the rest of the lines so we don't miss lines
			for len(logs.lines) > 0 {
				line := <-logs.lines
				logs.consumeLine(line, fileHandle)
			}
			return
		case line := <-logs.lines:
			logs.consumeLine(line, fileHandle)
		default:
			if !logs.pruned {
				log.Info("pruning seen lines")
				for key := range logs.seenLines {
					if time.Since(key) > keepDuration {
						delete(logs.seenLines, key)
					}
				}
				logs.pruned = true
			} else {
				time.Sleep(time.Microsecond)
			}
		}
	}
}

func (logs *LogsCollector) processLines(line string) string {
	if strings.ContainsRune(line, lineDelim) {
		lines := strings.Split(line, "\n")
		for n := 0; n < len(lines)-2; n++ {
			logs.lines <- lines[n]
		}
		line = lines[len(lines)-1]
	}
	return line
}

func (logs *LogsCollector) poll() error {
	logs.lpLock.RLock()
	defer logs.lpLock.RUnlock()
	sinceSeconds := int64(time.Since(logs.lastPoll).Seconds())
	podLogOptions := v1.PodLogOptions{
		SinceSeconds: &sinceSeconds,
		Container:    contexts.PTPContainer,
		Follow:       true,
		Previous:     false,
		Timestamps:   true,
	}
	podName, err := logs.client.FindPodNameFromPrefix(contexts.PTPNamespace, contexts.PTPPodNamePrefix)
	if err != nil {
		return fmt.Errorf("failed to poll: %w", err)
	}
	logs.SetLastPool()
	podLogRequest := logs.client.K8sClient.CoreV1().
		Pods(contexts.PTPNamespace).
		GetLogs(podName, &podLogOptions).
		Timeout(followTimeout)
	stream, err := podLogRequest.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to poll: %w", err)
	}
	defer stream.Close()

	line := ""
	start := time.Now()
	buf := make([]byte, streamingBufferSize)
	for time.Since(start) <= followDuration {
		nBytes, err := stream.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to poll: %w", err)
		}
		line += string(buf[:nBytes])
		line = logs.processLines(line)
	}
	if len(line) > 0 {
		logs.lines <- line
	}
	return nil
}

// Poll collects information from the cluster then
// calls the callback.Call to allow that to persist it
func (logs *LogsCollector) Poll(resultsChan chan PollResult, wg *utils.WaitGroupCount) {
	defer func() {
		wg.Done()
	}()
	errorsToReturn := make([]error, 0)
	err := logs.poll()
	if err != nil {
		errorsToReturn = append(errorsToReturn, err)
	}
	resultsChan <- PollResult{
		CollectorName: LogsCollectorName,
		Errors:        errorsToReturn,
	}
}

// CleanUp stops a running collector
func (logs *LogsCollector) CleanUp() error {
	logs.running = false
	logs.quit <- os.Kill
	logs.wg.Wait()
	return nil
}

// Returns a new LogsCollector from the CollectionConstuctor Factory
func NewLogsCollector(constructor *CollectionConstructor) (Collector, error) {
	collector := LogsCollector{
		running:            false,
		client:             constructor.Clientset,
		quit:               make(chan os.Signal),
		pollInterval:       logPollInterval,
		pruned:             true,
		lines:              make(chan string, lineChanLength),
		seenLines:          make(map[time.Time][]string),
		lastPoll:           time.Now().Add(-time.Second), // Stop initial since seconds from being 0 as its invalid
		withTimeStamps:     false,
		logsOutputFileName: constructor.LogsOutputFile,
	}

	return &collector, nil
}

func init() {
	// Make log opt in as in may lose some data.
	RegisterCollector(LogsCollectorName, NewLogsCollector, optIn)
}
