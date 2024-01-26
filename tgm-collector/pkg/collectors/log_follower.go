// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/loglines"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/utils"
)

const (
	lineSliceChanLength = 100
	lineChanLength      = 1000
	lineDelim           = '\n'
	streamingBufferSize = 2000
	logPollInterval     = 2
	logFilePermissions  = 0666
	keepGenerations     = 5
)

var (
	followDuration = logPollInterval * time.Second
	followTimeout  = 30 * followDuration
)

// LogsCollector collects logs from repeated calls to the kubeapi with overlapping query times,
// the lines are then fed into a channel, in another goroutine they are de-duplicated and written to an output file.
//
// Overlap:
// cmd       followDuration
// |---------------|
// since          cmd        followDuration
// |---------------|---------------|
// .............. since           cmd        followDuration
// ..............  |---------------|---------------|
//
// This was done because oc logs and the kubeapi endpoint which it uses does not look back
// over a log rotation event, nor does it continue to follow.
//
// Log aggregators would be preferred over this method however this requires extra infra
// which might not be present in the environment.
type LogsCollector struct {
	*baseCollector
	generations        loglines.Generations
	writeQuit          chan os.Signal
	lines              chan *loglines.ProcessedLine
	slices             chan *loglines.LineSlice
	client             *clients.Clientset
	sliceQuit          chan os.Signal
	logsOutputFileName string
	lastPoll           loglines.GenerationalLockedTime
	wg                 sync.WaitGroup
	withTimeStamps     bool
	pruned             bool
}

const (
	LogsCollectorName = "Logs"
	LogsInfo          = "log-line"
)

func (logs *LogsCollector) SetLastPoll(pollTime time.Time) {
	logs.lastPoll.Update(pollTime)
}

// Start sets up the collector so it is ready to be polled
func (logs *LogsCollector) Start() error {
	go logs.processSlices()
	go logs.writeToLogFile()
	logs.generations.Dumper.Start()
	logs.running = true
	return nil
}

func (logs *LogsCollector) writeLine(line *loglines.ProcessedLine, writer io.StringWriter) {
	var err error
	if logs.withTimeStamps {
		_, err = writer.WriteString(line.Full + "\n")
	} else {
		_, err = writer.WriteString(line.Content + "\n")
	}
	if err != nil {
		log.Error("failed to write log output to file")
	}
}

//nolint:cyclop // allow this to be a little complicated
func (logs *LogsCollector) processSlices() {
	logs.wg.Add(1)
	defer logs.wg.Done()
	for {
		select {
		case sig := <-logs.sliceQuit:
			log.Debug("Clearing slices")
			for len(logs.slices) > 0 {
				lineSlice := <-logs.slices
				logs.generations.Add(lineSlice)
			}
			log.Debug("Flushing remaining generations")
			deduplicated := logs.generations.FlushAll()
			for _, line := range deduplicated.Lines {
				logs.lines <- line
			}
			log.Debug("Sending Signal to writer")
			logs.writeQuit <- sig
			return
		case lineSlice := <-logs.slices:
			logs.generations.Add(lineSlice)
		default:
			if logs.generations.ShouldFlush() {
				deduplicated := logs.generations.Flush()
				for _, line := range deduplicated.Lines {
					logs.lines <- line
				}
			}
			time.Sleep(time.Nanosecond)
		}
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
		case <-logs.writeQuit:
			// Consume the rest of the lines so we don't miss lines
			for len(logs.lines) > 0 {
				line := <-logs.lines
				logs.writeLine(line, fileHandle)
			}
			return
		case line := <-logs.lines:
			logs.writeLine(line, fileHandle)
		default:
			time.Sleep(time.Nanosecond)
		}
	}
}

func processLine(line string) (*loglines.ProcessedLine, error) {
	splitLine := strings.SplitN(line, " ", 2) //nolint:gomnd // moving this to a var would make the code less clear
	if len(splitLine) < 2 {                   //nolint:gomnd // moving this to a var would make the code less clear
		return nil, fmt.Errorf("failed to split line %s", line)
	}
	timestampPart := splitLine[0]
	lineContent := splitLine[1]
	timestamp, err := time.Parse(time.RFC3339, timestampPart)
	if err != nil {
		// This is not a value line something went wrong
		return nil, fmt.Errorf("failed to process timestamp from line: '%s'", line)
	}
	processed := &loglines.ProcessedLine{
		Timestamp: timestamp,
		Content:   strings.TrimRightFunc(lineContent, unicode.IsSpace),
		Full:      strings.TrimRightFunc(line, unicode.IsSpace),
	}
	return processed, nil
}

//nolint:funlen // allow long function
func processStream(stream io.ReadCloser, expectedEndtime time.Time) ([]*loglines.ProcessedLine, error) {
	scanner := bufio.NewScanner(stream)
	segment := make([]*loglines.ProcessedLine, 0)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return segment, fmt.Errorf("error while reading logs stream %w", err)
		}
		pline, err := processLine(scanner.Text())
		if err != nil {
			log.Warning("failed to process line: ", err)
			continue
		}
		segment = append(segment, pline)
		if expectedEndtime.Sub(pline.Timestamp) < 0 {
			// Were past our expected end time lets finish there
			break
		}
	}
	return segment, nil
}

func (logs *LogsCollector) poll() error {
	podName, err := logs.client.FindPodNameFromPrefix(contexts.PTPNamespace, contexts.PTPPodNamePrefix)
	if err != nil {
		return fmt.Errorf("failed to poll: %w", err)
	}
	podLogOptions := v1.PodLogOptions{
		SinceTime:  &metav1.Time{Time: logs.lastPoll.Time()},
		Container:  contexts.PTPContainer,
		Follow:     true,
		Previous:   false,
		Timestamps: true,
	}
	podLogRequest := logs.client.K8sClient.CoreV1().
		Pods(contexts.PTPNamespace).
		GetLogs(podName, &podLogOptions).
		Timeout(followTimeout)
	stream, err := podLogRequest.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to poll when r: %w", err)
	}
	defer stream.Close()

	start := time.Now()
	generation := logs.lastPoll.Generation()
	lines, err := processStream(stream, time.Now().Add(logs.GetPollInterval()))
	if err != nil {
		return err
	}
	if len(lines) > 0 {
		lineSlice := loglines.MakeSliceFromLines(lines, generation)
		logs.slices <- lineSlice
		logs.SetLastPoll(start)
	}
	return nil
}

// Poll collects log lines
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
	logs.sliceQuit <- os.Kill
	log.Debug("waiting for logs to complete")
	logs.wg.Wait()
	logs.generations.Dumper.Stop()
	return nil
}

// Returns a new LogsCollector from the CollectionConstuctor Factory
func NewLogsCollector(constructor *CollectionConstructor) (Collector, error) {
	collector := LogsCollector{
		baseCollector: newBaseCollector(
			logPollInterval,
			false,
			constructor.Callback,
		),
		client:             constructor.Clientset,
		sliceQuit:          make(chan os.Signal),
		writeQuit:          make(chan os.Signal),
		pruned:             true,
		slices:             make(chan *loglines.LineSlice, lineSliceChanLength),
		lines:              make(chan *loglines.ProcessedLine, lineChanLength),
		lastPoll:           loglines.NewGenerationalLockedTime(time.Now().Add(-time.Second)), // Stop initial since seconds from being 0 as its invalid
		withTimeStamps:     constructor.IncludeLogTimestamps,
		logsOutputFileName: constructor.LogsOutputFile,
		generations: loglines.Generations{
			Store:  make(map[uint32][]*loglines.LineSlice),
			Dumper: loglines.NewGenerationDumper(constructor.TempDir, constructor.KeepDebugFiles),
		},
	}
	return &collector, nil
}

func init() {
	// Make log opt in as in may lose some data.
	RegisterCollector(LogsCollectorName, NewLogsCollector, optional)
}
