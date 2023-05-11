// SPDX-License-Identifier: GPL-2.0-or-later

package runner

import (
	"fmt"
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

type CollectorRunner struct {
	quit               chan os.Signal
	collecterInstances []*collectors.Collector
	collectorNames     []string
}

func NewCollectorRunner() *CollectorRunner {
	collectorNames := make([]string, 0)
	collectorNames = append(collectorNames, "PTP")
	return &CollectorRunner{
		collecterInstances: make([]*collectors.Collector, 0),
		collectorNames:     collectorNames,
		quit:               getQuitChannel(),
	}
}

// initialise will call the constuctor for each
// value in collector name, it will panic if a collector name is not known.
func (runner *CollectorRunner) initialise(
	callback callbacks.Callback,
	ptpInterface string,
	clientset *clients.Clientset,
	pollRate float64,
) {
	constuctor := collectors.CollectionConstuctor{
		Callback:     callback,
		PTPInterface: ptpInterface,
		Clientset:    clientset,
		PollRate:     pollRate,
	}

	for _, constuctorName := range runner.collectorNames {
		var newCollector collectors.Collector
		switch constuctorName {
		case "PTP":
			NewPTPCollector, err := constuctor.NewPTPCollector() //nolint:govet // TODO clean this up
			utils.IfErrorPanic(err)
			newCollector = NewPTPCollector
			log.Debug("PTP Collector")
		default:
			newCollector = nil
			panic("Unknown collector")
		}
		if newCollector != nil {
			runner.collecterInstances = append(runner.collecterInstances, &newCollector)
			log.Debugf("Added collector %T, %v", newCollector, newCollector)
		}
	}
	log.Debugf("Collectors %v", runner.collecterInstances)
}

// start configures all collectors to start collecting all there data keys
func (runner *CollectorRunner) start() {
	for _, collector := range runner.collecterInstances {
		log.Debugf("start collector %v", collector)
		err := (*collector).Start(collectors.All)
		utils.IfErrorPanic(err)
	}
}

// poll iterates though each running collector and
// checks if it should be polled if so then will call poll
func (runner *CollectorRunner) poll() {
	for _, collector := range runner.collecterInstances {
		log.Debugf("Running collector: %v", collector)

		if (*collector).ShouldPoll() {
			log.Debugf("poll %v", collector)
			errors := (*collector).Poll()
			if len(errors) > 0 {
				// TODO: handle errors (better)
				log.Error(errors)
			}
		}
	}
}

// cleanup calls cleanup on each collector
func (runner *CollectorRunner) cleanUp() {
	for _, collector := range runner.collecterInstances {
		log.Debugf("cleanup %v", collector)
		errColletor := (*collector).CleanUp(collectors.All)
		utils.IfErrorPanic(errColletor)
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

	runner.initialise(callback, ptpInterface, clientset, pollRate)
	runner.start()

out:
	for numberOfPolls := 1; pollCount < 0 || numberOfPolls <= pollCount; numberOfPolls++ {
		select {
		case <-runner.quit:
			log.Info("Killed shuting down")
			break out
		default:
			runner.poll()
			time.Sleep(time.Duration(float64(time.Second.Nanoseconds()) / pollRate))
		}
	}
	runner.cleanUp()
	errCallback := callback.CleanUp()
	utils.IfErrorPanic(errCallback)
}
