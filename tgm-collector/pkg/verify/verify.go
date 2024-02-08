// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"fmt"
	"os"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
	validationsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/validations"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/validations"
)

const (
	unknownMsgPrefix = "The following error occurred when trying to gather environment data for the following validations"
	antPowerRetries  = 3
)

//nolint:ireturn // this needs to be an interface
func getDevInfoValidations(
	clientset *clients.Clientset,
	interfaceName string,
) []validationsBase.Validation {
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)
	devInfo, err := devices.GetPTPDeviceInfo(interfaceName, ctx)
	utils.IfErrorExitOrPanic(err)
	devDetails := validations.NewDeviceDetails(&devInfo)
	devFirmware := validations.NewDeviceFirmware(&devInfo)
	devDriver := validations.NewDeviceDriver(&devInfo)
	return []validationsBase.Validation{devDetails, devFirmware, devDriver}
}

func getGPSVersionValidations(
	clientset *clients.Clientset,
) []validationsBase.Validation {
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)
	gnssVersions, err := devices.GetGPSVersions(ctx)
	utils.IfErrorExitOrPanic(err)
	return []validationsBase.Validation{
		validations.NewGNSS(&gnssVersions),
		validations.NewGPSDVersion(&gnssVersions),
		validations.NewGNSDevices(&gnssVersions),
		validations.NewGNSSModule(&gnssVersions),
		validations.NewGNSSProtocol(&gnssVersions),
	}
}

func getGPSStatusValidation(
	clientset *clients.Clientset,
) []validationsBase.Validation {
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)

	// If we need to do this for more validations then consider a generic
	var antCheck *validations.GNSSAntStatus
	var gpsDetails devices.GPSDetails
	for i := 0; i < antPowerRetries; i++ {
		gpsDetails, err = devices.GetGPSNav(ctx)
		if err != nil {
			continue
		}
		if antCheck = validations.NewGNSSAntStatus(&gpsDetails); antCheck.Verify() == nil {
			break
		}
		time.Sleep(time.Second)
	}
	utils.IfErrorExitOrPanic(err)
	return []validationsBase.Validation{
		antCheck,
		validations.NewGNSSNavStatus(&gpsDetails),
	}
}

func getValidations(interfaceName, kubeConfig string) []validationsBase.Validation {
	checks := make([]validationsBase.Validation, 0)
	clientset, err := clients.GetClientset(kubeConfig)
	utils.IfErrorExitOrPanic(err)
	checks = append(checks, getDevInfoValidations(clientset, interfaceName)...)
	checks = append(checks, getGPSVersionValidations(clientset)...)
	checks = append(checks, getGPSStatusValidation(clientset)...)
	checks = append(
		checks,
		validations.NewIsGrandMaster(clientset),
		validations.NewOperatorVersion(clientset),
		validations.NewClusterVersion(clientset),
	)
	return checks
}

func reportAnalyserJSON(results []*ValidationResult) {
	callback, err := callbacks.SetupCallback("-", callbacks.AnalyserJSON)
	utils.IfErrorExitOrPanic(err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].validation.GetOrder() < results[j].validation.GetOrder()
	})

	anyHasFailed := false
	for _, res := range results {
		if res.resType == resTypeFailure {
			anyHasFailed = true
		}
		err := callback.Call(res, "env-check")
		if err != nil {
			log.Errorf("callback failed during validation %s", err.Error())
		}
	}

	if anyHasFailed {
		os.Exit(int(utils.InvalidEnv))
	}
}

//nolint:funlen,cyclop // allow slightly long function
func report(results []*ValidationResult, useAnalyserJSON bool) {
	if useAnalyserJSON {
		reportAnalyserJSON(results)
		return
	}

	failures := make([]*ValidationResult, 0)
	unknown := make([]*ValidationResult, 0)

	for _, res := range results {
		//nolint:exhaustive // Not reporting successes so no need to gather them
		switch res.resType {
		case resTypeFailure:
			failures = append(failures, res)
		case resTypeUnknown:
			unknown = append(unknown, res)
		}
	}

	// Report unknowns along side failures
	if len(unknown) > 0 {
		dataErrors := make([]error, 0)
		for _, res := range unknown {
			dataErrors = append(dataErrors, res.GetPrefixedError())
		}
		log.Error(utils.MakeCompositeError(unknownMsgPrefix, dataErrors))
	}

	switch {
	case len(failures) > 0:
		validationsErrors := make([]error, 0)
		for _, res := range failures {
			validationsErrors = append(validationsErrors, res.GetPrefixedError())
		}
		err := utils.MakeCompositeInvalidEnvError(validationsErrors)
		utils.IfErrorExitOrPanic(err)
	case len(unknown) > 0:
		// If only unknowns print this message
		fmt.Println("Some checks did not complete, it is likely something is not correct in the environment") //nolint:forbidigo // This to print out to the user
	default:
		fmt.Println("No issues found.") //nolint:forbidigo // This to print out to the user
	}
}

func Verify(interfaceName, kubeConfig string, useAnalyserJSON bool) {
	checks := getValidations(interfaceName, kubeConfig)

	results := make([]*ValidationResult, 0)
	for _, check := range checks {
		results = append(results, NewValidationResult(check))
	}

	report(results, useAnalyserJSON)
}
