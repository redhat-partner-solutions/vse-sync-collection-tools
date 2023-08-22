// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/validations"
)

const (
	unknownMsgPrefix = "The following error occurred when trying to gather environment data for the following validations"
	antPowerRetries  = 3
)

//nolint:ireturn // this needs to be an interface
func getDevInfoValidations(
	clientset *clients.Clientset,
	interfaceName string,
) []validations.Validation {
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)
	devInfo, err := devices.GetPTPDeviceInfo(interfaceName, ctx)
	utils.IfErrorExitOrPanic(err)
	devDetails := validations.NewDeviceDetails(&devInfo)
	devFirmware := validations.NewDeviceFirmware(&devInfo)
	devDriver := validations.NewDeviceDriver(&devInfo)
	return []validations.Validation{devDetails, devFirmware, devDriver}
}

func getGPSVersionValidations(
	clientset *clients.Clientset,
) []validations.Validation {
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)
	gnssVersions, err := devices.GetGPSVersions(ctx)
	utils.IfErrorExitOrPanic(err)
	return []validations.Validation{
		validations.NewGNSS(&gnssVersions),
		validations.NewGPSDVersion(&gnssVersions),
		validations.NewGNSDevices(&gnssVersions),
		validations.NewGNSSModule(&gnssVersions),
		validations.NewGNSSProtocol(&gnssVersions),
	}
}

func getGPSStatusValidation(
	clientset *clients.Clientset,
) []validations.Validation {
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)

	// If we need to do this for more validations then consider a generic
	var antCheck *validations.GNSSAntStatus
	var gpsDetails devices.GPSDetails
	for i := 0; i < antPowerRetries; i++ {
		gpsDetails, err = devices.GetGPSNav(ctx)
		utils.IfErrorExitOrPanic(err)
		if check := validations.NewGNSSAntStatus(&gpsDetails); check.Verify() == nil {
			antCheck = check
			break
		}
		time.Sleep(time.Second)
	}
	return []validations.Validation{
		antCheck,
		validations.NewGNSSNavStatus(&gpsDetails),
	}
}

func getValidations(interfaceName, kubeConfig string) []validations.Validation {
	checks := make([]validations.Validation, 0)
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

func reportAnalyserJSON(failures, successes, unknown []*ValidationResult) {
	callback, err := callbacks.SetupCallback("-", callbacks.AnalyserJSON)
	utils.IfErrorExitOrPanic(err)

	for _, unknownCheck := range unknown {
		err := callback.Call(unknownCheck, "env-check-unknown")
		if err != nil {
			log.Errorf("callback failed during validation %s", err.Error())
		}
	}
	for _, failure := range failures {
		err := callback.Call(failure, "env-check-failure")
		if err != nil {
			log.Errorf("callback failed during validation %s", err.Error())
		}
	}
	for _, success := range successes {
		err := callback.Call(success, "env-check-success")
		if err != nil {
			log.Errorf("callback failed during validation %s", err.Error())
		}
	}
}

func report(failures, successes, unknown []*ValidationResult, useAnalyserJSON bool) {
	if useAnalyserJSON {
		reportAnalyserJSON(failures, successes, unknown)
		if len(failures) > 0 {
			os.Exit(int(utils.InvalidEnv))
		}
		return
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
		fmt.Println("Some checks failed, it is likely something is not correct in the environment") //nolint:forbidigo // This to print out to the user
	default:
		fmt.Println("No issues found.") //nolint:forbidigo // This to print out to the user
	}
}

func Verify(interfaceName, kubeConfig string, useAnalyserJSON bool) {
	checks := getValidations(interfaceName, kubeConfig)

	failures := make([]*ValidationResult, 0)
	successes := make([]*ValidationResult, 0)
	unknown := make([]*ValidationResult, 0)

	for _, check := range checks {
		err := check.Verify()
		res := &ValidationResult{
			err:        err,
			validation: check,
		}
		if err != nil {
			if res.IsInvalidEnv() {
				failures = append(failures, res)
			} else {
				unknown = append(unknown, res)
			}
		} else {
			successes = append(successes, res)
		}
	}

	report(failures, successes, unknown, useAnalyserJSON)
}
