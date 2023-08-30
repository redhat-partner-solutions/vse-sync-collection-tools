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
	logPollInterval     = 5
	logFilePermissions  = 0666
	keepGenerations     = 10
)

var (
	followDuration = logPollInterval * time.Second
	followTimeout  = 30 * followDuration
)

type ProcessedLine struct {
	Timestamp  time.Time
	Raw        string
	Content    string
	Generation uint32
}

type GenerationalLockedTime struct {
	time       time.Time
	lock       sync.RWMutex
	generation uint32
}

func (lt *GenerationalLockedTime) Time() time.Time {
	lt.lock.RLock()
	defer lt.lock.RUnlock()
	return lt.time
}

func (lt *GenerationalLockedTime) Generation() uint32 {
	lt.lock.RLock()
	defer lt.lock.RUnlock()
	return lt.generation
}

func (lt *GenerationalLockedTime) Update(update time.Time) {
	lt.lock.Lock()
	defer lt.lock.Unlock()
	lt.time = update
	lt.generation += 1
}

// LogsCollector collects logs from repeated calls to the kubeapi with overlapping query times,
// the lines are then fed into a channel, in another gorotine they are de-duplicated and written to an output file.
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
	quit               chan os.Signal
	lines              chan *ProcessedLine
	seenLines          map[time.Time][]string
	generations        map[uint32]time.Time
	client             *clients.Clientset
	logsOutputFileName string
	lastPoll           GenerationalLockedTime
	wg                 sync.WaitGroup
	pollInterval       int
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

func (logs *LogsCollector) SetLastPoll(pollTime time.Time) {
	logs.lastPoll.Update(pollTime)
}

// Start sets up the collector so it is ready to be polled
func (logs *LogsCollector) Start() error {
	go logs.writeToLogFile()
	logs.running = true
	return nil
}

func (logs *LogsCollector) consumeLine(line *ProcessedLine, writer io.StringWriter) {
	if genTimestamp, ok := logs.generations[line.Generation]; !ok || line.Timestamp.Sub(genTimestamp) < 0 {
		logs.generations[line.Generation] = line.Timestamp
	}

	if seen, ok := logs.seenLines[line.Timestamp]; ok {
		for _, seenContent := range seen {
			if seenContent == line.Content {
				// Already have this line skip it
				return
			}
		}
	}
	logs.seenLines[line.Timestamp] = append(logs.seenLines[line.Timestamp], line.Content)
	var err error
	if logs.withTimeStamps {
		_, err = writer.WriteString(line.Raw + "\n")
	} else {
		_, err = writer.WriteString(line.Content + "\n")
	}
	if err != nil {
		log.Error(fmt.Errorf("failed to write log output to file"))
	}
}

func (logs *LogsCollector) prune(currentGenSeen uint32) {
	log.Debug("logs: pruning seen lines")
	removeGen := currentGenSeen - keepGenerations
	keepTime := logs.generations[removeGen]
	for key := range logs.seenLines {
		if key.Sub(keepTime) < 0 {
			delete(logs.seenLines, key)
		}
	}
	delete(logs.generations, removeGen)
	logs.pruned = true
}

func (logs *LogsCollector) writeToLogFile() {
	logs.wg.Add(1)
	defer logs.wg.Done()

	fileHandle, err := os.OpenFile(logs.logsOutputFileName, os.O_CREATE|os.O_WRONLY, logFilePermissions)
	utils.IfErrorExitOrPanic(err)
	defer fileHandle.Close()
	var currentGenSeen uint32 = 0
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
			if line.Generation > currentGenSeen {
				currentGenSeen = line.Generation
				logs.pruned = false
			}
			logs.consumeLine(line, fileHandle)
		default:
			if !logs.pruned && currentGenSeen >= keepGenerations {
				logs.prune(currentGenSeen)
			} else {
				time.Sleep(time.Microsecond)
			}
		}
	}
}

func processLine(line string, generation uint32) (*ProcessedLine, error) {
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
	processed := &ProcessedLine{
		Timestamp:  timestamp,
		Content:    lineContent,
		Raw:        line,
		Generation: generation,
	}
	return processed, nil
}

func (logs *LogsCollector) processLines(line string, generation uint32) (string, time.Time) {
	var lastTimestamp time.Time
	if strings.ContainsRune(line, lineDelim) {
		lines := strings.Split(line, "\n")
		for index := 0; index < len(lines)-2; index++ {
			log.Debug("logs: lines: ", lines[index])
			processed, err := processLine(lines[index], generation)
			if err != nil {
				log.Warning("logs: error when processing lines: ", err)
			} else {
				logs.lines <- processed
				lastTimestamp = processed.Timestamp
			}
		}
		line = lines[len(lines)-1]
	}
	return line, lastTimestamp
}

func durationPassed(first, current time.Time, duration time.Duration) bool {
	if first.IsZero() {
		return false
	}
	if current.IsZero() {
		return false
	}
	return duration <= current.Sub(first)
}

//nolint:funlen // allow long function
func processStream(logs *LogsCollector, stream io.ReadCloser, sinceTime time.Duration) error {
	line := ""
	generation := logs.lastPoll.Generation()
	lastTimestamp := time.Time{}
	firstTimestamp := time.Time{}
	timestamp := time.Time{}
	buf := make([]byte, streamingBufferSize)
	expectedDuration := sinceTime + followDuration

	for !durationPassed(firstTimestamp, lastTimestamp, expectedDuration) {
		nBytes, err := stream.Read(buf)
		if err == io.EOF { //nolint:errorlint // No need for Is or As check as this should just be EOF
			log.Warning("log stream ended unexpectedly, possible log rotation detected at ", lastTimestamp)
			break
		}
		if err != nil {
			return fmt.Errorf("failed reading buffer: %w", err)
		}
		if nBytes == 0 {
			continue
		}
		line += string(buf[:nBytes])
		line, timestamp = logs.processLines(line, generation)
		// set First legitimate timestamp

		if !timestamp.IsZero() {
			if firstTimestamp.IsZero() {
				firstTimestamp = timestamp
			}
			lastTimestamp = timestamp
		}
	}
	if len(line) > 0 {
		processed, err := processLine(line, generation)
		if err == nil {
			logs.lines <- processed
		}
	}
	log.Debug("logs: Finish stream")
	return nil
}

func (logs *LogsCollector) poll() error {
	podName, err := logs.client.FindPodNameFromPrefix(contexts.PTPNamespace, contexts.PTPPodNamePrefix)
	if err != nil {
		return fmt.Errorf("failed to poll: %w", err)
	}
	sinceTime := time.Since(logs.lastPoll.Time())
	sinceSeconds := int64(sinceTime.Seconds())

	podLogOptions := v1.PodLogOptions{
		SinceSeconds: &sinceSeconds,
		Container:    contexts.PTPContainer,
		Follow:       true,
		Previous:     false,
		Timestamps:   true,
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
	err = processStream(logs, stream, sinceTime)
	if err != nil {
		return err
	}
	logs.SetLastPoll(start)
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
		lines:              make(chan *ProcessedLine, lineChanLength),
		seenLines:          make(map[time.Time][]string),
		generations:        make(map[uint32]time.Time),
		lastPoll:           GenerationalLockedTime{time: time.Now().Add(-time.Second)}, // Stop initial since seconds from being 0 as its invalid
		withTimeStamps:     constructor.IncludeLogTimestamps,
		logsOutputFileName: constructor.LogsOutputFile,
	}
	return &collector, nil
}

func init() {
	// Make log opt in as in may lose some data.
	RegisterCollector(LogsCollectorName, NewLogsCollector, includeByDefault)
}
