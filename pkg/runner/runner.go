// SPDX-License-Identifier: GPL-2.0-or-later

package runner

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	maxRunningPolls           = 3
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

type CollectorRunner struct {
	quit                 chan os.Signal
	collectorQuitChannel map[string]chan os.Signal
	pollResults          chan collectors.PollResult
	erroredPolls         chan collectors.PollResult
	collectorInstances   map[string]collectors.Collector
	collectorNames       []string
	pollCount            int
	pollRate             float64
	devInfoAnnouceRate   float64
	runningCollectorsWG  utils.WaitGroupCount
	runningAnnouncersWG  utils.WaitGroupCount
	onlyAnnouncers       bool
}

func NewCollectorRunner(selectedCollectors []string) *CollectorRunner {
	return &CollectorRunner{
		collectorInstances:   make(map[string]collectors.Collector),
		collectorNames:       GetCollectorsToRun(selectedCollectors),
		quit:                 getQuitChannel(),
		pollResults:          make(chan collectors.PollResult, pollResultsQueueSize),
		erroredPolls:         make(chan collectors.PollResult, pollResultsQueueSize),
		collectorQuitChannel: make(map[string]chan os.Signal, 1),
		onlyAnnouncers:       false,
	}
}

// initialise will call theconstructor for each
// value in collector name, it will panic if a collector name is not known.
//
//nolint:funlen // this is going to be a long function
func (runner *CollectorRunner) initialise(
	callback callbacks.Callback,
	ptpInterface string,
	clientset *clients.Clientset,
	pollRate float64,
	pollCount int,
	devInfoAnnouceRate float64,
) {
	runner.pollRate = pollRate
	runner.pollCount = pollCount
	runner.devInfoAnnouceRate = devInfoAnnouceRate

	constructor := collectors.CollectionConstructor{
		Callback:           callback,
		PTPInterface:       ptpInterface,
		Clientset:          clientset,
		PollRate:           pollRate,
		DevInfoAnnouceRate: devInfoAnnouceRate,
	}

	for _, constructorName := range runner.collectorNames {
		var newCollector collectors.Collector
		newInstance := true
		switch constructorName {
		case collectors.DPLLCollectorName:
			NewDPLLCollector, err := constructor.NewDPLLCollector()
			utils.IfErrorPanic(err)
			newCollector = NewDPLLCollector
			log.Debug("DPLL Collector")

		case collectors.DevInfoCollectorName:
			NewDevInfCollector, err := constructor.NewDevInfoCollector(runner.erroredPolls)
			utils.IfErrorPanic(err)
			newCollector = NewDevInfCollector
			log.Debug("DPLL Collector")
		case collectors.GPSCollectorName:
			NewGPSCollector, err := constructor.NewGPSCollector()
			utils.IfErrorPanic(err)
			newCollector = NewGPSCollector
			log.Debug("GPS Collector")
		default:
			newInstance = false
			log.Errorf("Unknown collector %s", constructorName)
		}
		if newInstance {
			runner.collectorInstances[constructorName] = newCollector
			log.Debugf("Added collector %T, %v", newCollector, newCollector)
		}
	}
	log.Debugf("Collectors %v", runner.collectorInstances)
	runner.setOnlyAnnouncers()
}

func (runner *CollectorRunner) setOnlyAnnouncers() {
	onlyAnnouncers := true
	for _, collector := range runner.collectorInstances {
		if !collector.IsAnnouncer() {
			onlyAnnouncers = false
			break
		}
	}
	runner.onlyAnnouncers = onlyAnnouncers
}

func (runner *CollectorRunner) shouldKeepPolling(
	collector collectors.Collector,
	runningPolls *utils.WaitGroupCount,
) bool {
	if collector.IsAnnouncer() && !runner.onlyAnnouncers {
		return runner.runningCollectorsWG.GetCount() > 0
	} else {
		return runner.pollCount < 0 || (collector.GetPollCount()+runningPolls.GetCount()) <= runner.pollCount
	}
}

func (runner *CollectorRunner) poller(
	collectorName string,
	collector collectors.Collector,
	quit chan os.Signal,
	wg *utils.WaitGroupCount,
) {
	defer wg.Done()
	var lastPoll time.Time
	pollRate := collector.GetPollRate()
	inversePollRate := 1.0 / pollRate
	runningPolls := utils.WaitGroupCount{}
	log.Debugf("Collector with poll rate %f wait time %f", pollRate, inversePollRate)
	for runner.shouldKeepPolling(collector, &runningPolls) {
		// If pollResults were to block we do not want to keep spawning polls
		// so we shouldn't allow too many polls to be running simultaneously
		if runningPolls.GetCount() >= maxRunningPolls {
			runningPolls.Wait()
		}
		log.Debugf("Collector GoRoutine: %s", collectorName)
		select {
		case <-quit:
			log.Infof("Killed shutting down collector %s waiting for running polls to finish", collectorName)
			runningPolls.Wait()
			return
		default:
			if lastPoll.IsZero() || time.Since(lastPoll).Seconds() > inversePollRate {
				lastPoll = time.Now()
				log.Debugf("poll %s", collectorName)
				runningPolls.Add(1)
				go collector.Poll(runner.pollResults, &runningPolls)
			}
			time.Sleep(time.Microsecond)
		}
	}
	runningPolls.Wait()
	log.Debugf("Collector finished %s", collectorName)
}

// start configures all collectors to start collecting all their data keys
func (runner *CollectorRunner) start() {
	for collectorName, collector := range runner.collectorInstances {
		log.Debugf("start collector %v", collector)
		err := collector.Start()
		utils.IfErrorPanic(err)

		log.Debugf("Spawning  collector: %v", collector)
		collectorName := collectorName
		collector := collector
		quit := make(chan os.Signal, 1)
		if collector.IsAnnouncer() {
			runner.collectorQuitChannel[collectorName] = quit
			runner.runningAnnouncersWG.Add(1)
			go runner.poller(collectorName, collector, quit, &runner.runningAnnouncersWG)
		} else {
			runner.collectorQuitChannel[collectorName] = quit
			runner.runningCollectorsWG.Add(1)
			go runner.poller(collectorName, collector, quit, &runner.runningCollectorsWG)
		}
	}
}

// cleanup calls cleanup on each collector
func (runner *CollectorRunner) cleanUpAll() {
	for collectorName, collector := range runner.collectorInstances {
		log.Debugf("cleanup %s", collectorName)
		err := collector.CleanUp()
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
	devInfoAnnouceRate float64,
	ptpInterface string,
	useAnalyserJSON bool,
) {
	clientset := clients.GetClientset(kubeConfig)

	outputFormat := callbacks.Raw
	if useAnalyserJSON {
		outputFormat = callbacks.AnalyserJSON
	}

	callback, err := callbacks.SetupCallback(outputFile, outputFormat)
	utils.IfErrorPanic(err)
	runner.initialise(callback, ptpInterface, clientset, pollRate, pollCount, devInfoAnnouceRate)
	runner.start()

	// Use wg count to know if any collectors are running.
	for (runner.runningCollectorsWG.GetCount() + runner.runningAnnouncersWG.GetCount()) > 0 {
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
			runner.runningAnnouncersWG.Wait()
		case pollRes := <-runner.pollResults:
			log.Infof("Received %v", pollRes)
			if len(pollRes.Errors) > 0 {
				log.Warnf("Poll %s failed: %v", pollRes.CollectorName, pollRes.Errors)
				// If erroredPolls blocks it could cause pollResults to fill and
				// block the execution of the collectors.
				runner.erroredPolls <- pollRes
			}
		default:
			log.Debug("Sleeping main func")
			time.Sleep(time.Millisecond)
		}
	}
	log.Info("Doing Cleanup")
	runner.cleanUpAll()
	err = callback.CleanUp()
	utils.IfErrorPanic(err)
}
