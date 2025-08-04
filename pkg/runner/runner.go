// SPDX-License-Identifier: GPL-2.0-or-later

package runner

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/constants"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

const (
	maxRunningPolls      = 3
	pollResultsQueueSize = 10
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
	endTime                time.Time
	quit                   chan os.Signal
	collectorQuitChannel   map[string]chan os.Signal
	pollResults            chan collectors.PollResult
	erroredPolls           chan collectors.PollResult
	collectorInstances     map[string]collectors.Collector
	collectorNames         []string
	runningCollectorsWG    utils.WaitGroupCount
	runningAnnouncersWG    utils.WaitGroupCount
	pollInterval           int
	devInfoAnnouceInterval int
	onlyAnnouncers         bool
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
func (runner *CollectorRunner) initialise( //nolint:funlen // allow a slightly long function
	constructor *collectors.CollectionConstructor,
	requestedDuration time.Duration,
) {
	runner.pollInterval = constructor.PollInterval
	runner.endTime = time.Now().Add(requestedDuration)
	runner.devInfoAnnouceInterval = constructor.DevInfoAnnouceInterval

	registry := collectors.GetRegistry()

	for _, collectorName := range runner.collectorNames {
		// Skip GPS/GNSS collectors for Boundary Clock
		if constructor.ClockType == constants.ClockTypeBC && collectorName == collectors.GPSCollectorName {
			log.Infof("Skipping GPS collector '%s' for Boundary Clock configuration", collectorName)
			continue
		}

		builderFunc, err := registry.GetBuilderFunc(collectorName)
		if err != nil {
			log.Error(err)
			continue
		}

		newCollector, err := builderFunc(constructor)
		var missingRequirements *utils.RequirementsNotMetError
		if errors.As(err, &missingRequirements) {
			// Requirements are missing so don't add the collector to collectorInstance
			// so that it doesn't get ran
			log.Warning(err.Error())
		} else {
			utils.IfErrorExitOrPanic(err)
			runner.collectorInstances[collectorName] = newCollector
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
) bool {
	if collector.IsAnnouncer() && !runner.onlyAnnouncers {
		return runner.runningCollectorsWG.GetCount() > 0
	} else {
		return time.Since(runner.endTime) <= 0
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
	pollInterval := collector.GetPollInterval()
	runningPolls := utils.WaitGroupCount{}
	log.Debugf("Collector with poll interval %f ", pollInterval.Seconds())
	for runner.shouldKeepPolling(collector) {
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
			log.Debug(
				"Collector GoRoutine:",
				collectorName,
				lastPoll, pollInterval,
				lastPoll.IsZero(),
				time.Since(lastPoll),
				time.Since(lastPoll) > pollInterval,
				lastPoll.IsZero() || time.Since(lastPoll) > pollInterval,
			)
			// Not using a ticker as the idea was to make
			// pollInterval dynamic to be able to respond
			// to events triggered the action tool
			if lastPoll.IsZero() || time.Since(lastPoll) > pollInterval {
				lastPoll = time.Now()
				log.Debugf("poll %s", collectorName)
				runningPolls.Add(1)
				go collector.Poll(runner.pollResults, &runningPolls)
			}
			time.Sleep(10 * time.Nanosecond) //nolint:gomnd,mnd // no point in making this its own var
		}
	}
	runningPolls.Wait()
	log.Debugf("Collector finished %s", collectorName)
}

// start configures all collectors to start collecting all their data keys
func (runner *CollectorRunner) start() {
	collectorsNames := make([]string, 0)
	announcersNames := make([]string, 0)

	for collectorName, collector := range runner.collectorInstances {
		if collector.IsAnnouncer() {
			announcersNames = append(announcersNames, collectorName)
		} else {
			collectorsNames = append(collectorsNames, collectorName)
		}
	}

	for _, collectorName := range append(collectorsNames, announcersNames...) {
		collectorName := collectorName

		collector := runner.collectorInstances[collectorName]
		log.Debugf("start collector %v", collector)
		err := collector.Start()
		utils.IfErrorExitOrPanic(err)

		log.Debugf("Spawning  collector: %v", collector)
		quit := make(chan os.Signal, 1)
		runner.collectorQuitChannel[collectorName] = quit
		var pollerWaitGroup *utils.WaitGroupCount

		if collector.IsAnnouncer() {
			pollerWaitGroup = &runner.runningAnnouncersWG
		} else {
			pollerWaitGroup = &runner.runningCollectorsWG
		}
		pollerWaitGroup.Add(1)
		go runner.poller(collectorName, collector, quit, pollerWaitGroup)
	}
}

// cleanup calls cleanup on each collector
func (runner *CollectorRunner) cleanUpAll() {
	for collectorName, collector := range runner.collectorInstances {
		log.Debugf("cleanup %s", collectorName)
		err := collector.CleanUp()
		utils.IfErrorExitOrPanic(err)
	}
}

// Run manages set of collectors.
// It first initialises them,
// then polls them on the correct cadence and
// finally cleans up the collectors when exiting
func (runner *CollectorRunner) Run( //nolint:funlen // allow a slightly long function
	requestedDuration time.Duration,
	constuctor *collectors.CollectionConstructor,
) {
	runner.initialise(constuctor, requestedDuration)
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
				log.Warnf("Poll %s had issues: %v. Will retry next poll", pollRes.CollectorName, pollRes.Errors)
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
	err := constuctor.Callback.CleanUp()
	utils.IfErrorExitOrPanic(err)
}
