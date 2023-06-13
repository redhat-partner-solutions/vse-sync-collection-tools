// SPDX-License-Identifier: GPL-2.0-or-later

package runner

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	maxConnsecutivePollErrors = 1800
	pollResultsQueueSize      = 10
)

// getQuitChannel creates and returns a channel for notifying
// that a exit signal has been received
func getQuitChannel() chan os.Signal {
	// Allow ourselves to handle shut down gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	return quit
}

//nolint:ireturn // The point of this function is to return a callback but which one is dependant on the input
func selectCollectorCallback(outputFile string) (callbacks.Callback, error) {
	if outputFile != "" {
		callback, err := callbacks.NewFileCallback(outputFile)
		return callback, fmt.Errorf("failed to create callback %w", err)
	} else {
		return callbacks.StdoutCallBack{}, nil
	}
}

type PollResult struct {
	CollectorName string
	Errors        []error
}

type CollectorRunner struct {
	quit                  chan os.Signal
	collectorQuitChannel  map[string]chan os.Signal
	pollResults           chan PollResult
	collecterInstances    map[string]*collectors.Collector
	consecutivePollErrors map[string]int
	collectorNames        []string
	pollCount             int
	pollRate              float64
	runningCollectorsWG   WaitGroupCount
}

func NewCollectorRunner() *CollectorRunner {
	collectorNames := make([]string, 0)
	collectorNames = append(collectorNames, collectors.PTPCollectorName, collectors.GPSCollectorName)
	return &CollectorRunner{
		collecterInstances:    make(map[string]*collectors.Collector),
		collectorNames:        collectorNames,
		quit:                  getQuitChannel(),
		pollResults:           make(chan PollResult, pollResultsQueueSize),
		collectorQuitChannel:  make(map[string]chan os.Signal, 1),
		consecutivePollErrors: make(map[string]int),
	}
}

// initialise will call the constuctor for each
// value in collector name, it will panic if a collector name is not known.
func (runner *CollectorRunner) initialise(
	callback callbacks.Callback,
	ptpInterface string,
	clientset *clients.Clientset,
	pollRate float64,
	pollCount int,
) {
	runner.pollRate = pollRate
	runner.pollCount = pollCount

	constuctor := collectors.CollectionConstuctor{
		Callback:     callback,
		PTPInterface: ptpInterface,
		Clientset:    clientset,
		PollRate:     pollRate,
	}

	for _, constuctorName := range runner.collectorNames {
		var newCollector collectors.Collector
		switch constuctorName {
		case collectors.PTPCollectorName:
			NewPTPCollector, err := constuctor.NewPTPCollector()
			utils.IfErrorPanic(err)
			newCollector = NewPTPCollector
			log.Debug("PTP Collector")
		case collectors.GPSCollectorName:
			NewGPSCollector, err := constuctor.NewGPSCollector()
			utils.IfErrorPanic(err)
			newCollector = NewGPSCollector
			log.Debug("PTP Collector")
		default:
			newCollector = nil
			panic("Unknown collector")
		}
		if newCollector != nil {
			runner.collecterInstances[constuctorName] = &newCollector
			log.Debugf("Added collector %T, %v", newCollector, newCollector)
		}
	}
	log.Debugf("Collectors %v", runner.collecterInstances)
}

func (runner *CollectorRunner) poller(collectorName string, collector collectors.Collector, quit chan os.Signal) {
	defer runner.runningCollectorsWG.Done()
	for numberOfPolls := 1; runner.pollCount < 0 || numberOfPolls <= runner.pollCount; numberOfPolls++ {
		log.Debugf("Collector GoRoutine: %s", collectorName)
		select {
		case <-quit:
			log.Infof("Killed shutting down collector %s", collectorName)
			return
		default:
			if collector.ShouldPoll() {
				log.Debugf("poll %s", collectorName)
				errors := collector.Poll()
				runner.pollResults <- PollResult{
					CollectorName: collectorName,
					Errors:        errors,
				}
			}
			time.Sleep(time.Duration(float64(time.Second.Nanoseconds()) / runner.pollRate))
		}
	}
	log.Debugf("Collector finished %s", collectorName)
}

// start configures all collectors to start collecting all their data keys
func (runner *CollectorRunner) start() {
	for collectorName, collector := range runner.collecterInstances {
		log.Debugf("start collector %v", collector)
		err := (*collector).Start(collectors.All)
		utils.IfErrorPanic(err)

		log.Debugf("Spawning  collector: %v", collector)
		collectorName := collectorName
		collector := collector
		quit := make(chan os.Signal, 1)
		runner.collectorQuitChannel[collectorName] = quit
		runner.runningCollectorsWG.Add(1)
		go runner.poller(collectorName, (*collector), quit)
	}
}

func (runner *CollectorRunner) stop(collectorName string) {
	log.Errorf("Stopping collector %s for to manny consecutive polling errors", collectorName)
	collector := runner.collecterInstances[collectorName]
	quit := runner.collectorQuitChannel[collectorName]
	quit <- os.Kill
	err := (*collector).CleanUp(collectors.All)
	utils.IfErrorPanic(err)
}

// cleanup calls cleanup on each collector
func (runner *CollectorRunner) cleanUpAll() {
	for collectorName, collector := range runner.collecterInstances {
		log.Debugf("cleanup %s", collectorName)
		err := (*collector).CleanUp(collectors.All)
		utils.IfErrorPanic(err)
	}
}

// Run manages set of collectors.
// It first initialises them,
// then polls them on the correct cadence and
// finally cleans up the collectors when exiting
func (runner *CollectorRunner) Run(
	kubeConfig string,
	outputFile string,
	pollCount int,
	pollRate float64,
	ptpInterface string,
) {
	clientset := clients.GetClientset(kubeConfig)
	callback, err := selectCollectorCallback(outputFile)
	utils.IfErrorPanic(err)
	runner.initialise(callback, ptpInterface, clientset, pollRate, pollCount)
	runner.start()

	// Use wg count to know if any collectors are running.
	for runner.runningCollectorsWG.GetCount() > 0 {
		log.Debugf("Main Loop ")
		select {
		case sig := <-runner.quit:
			log.Info("Killed shutting down")
			// Forward signal to collector QuitChannels
			for collectorName, quit := range runner.collectorQuitChannel {
				log.Infof("Killed shutting down: %s", collectorName)
				quit <- sig
			}
			runner.runningCollectorsWG.Wait()
		case pollRes := <-runner.pollResults:
			log.Infof("Received %v", pollRes)
			if len(pollRes.Errors) > 0 {
				runner.consecutivePollErrors[pollRes.CollectorName] += 1
				log.Warnf(
					"Increment failures for %s. Number of consecutive errors %d",
					pollRes.CollectorName,
					runner.consecutivePollErrors[pollRes.CollectorName],
				)
			} else {
				runner.consecutivePollErrors[pollRes.CollectorName] = 0
				log.Debugf("Reset failures for %s", pollRes.CollectorName)
			}
			if runner.consecutivePollErrors[pollRes.CollectorName] >= maxConnsecutivePollErrors {
				runner.stop(pollRes.CollectorName)
			}
		default:
			log.Debug("Sleeping main func")
			time.Sleep(time.Duration(float64(time.Second.Nanoseconds()) / pollRate))
		}
	}
	log.Info("Doing Cleanup")
	runner.cleanUpAll()
	err = callback.CleanUp()
	utils.IfErrorPanic(err)
}

type WaitGroupCount struct {
	sync.WaitGroup
	count int64
}

func (wg *WaitGroupCount) Add(delta int) {
	atomic.AddInt64(&wg.count, int64(delta))
	wg.WaitGroup.Add(delta)
}

func (wg *WaitGroupCount) Done() {
	atomic.AddInt64(&wg.count, -1)
	wg.WaitGroup.Done()
}

func (wg *WaitGroupCount) GetCount() int {
	return int(atomic.LoadInt64(&wg.count))
}
