// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/vaildations"
)

//nolint:ireturn // this needs to be an interface
func getDevInfoValidations(
	clientset *clients.Clientset,
	interfaceName string,
) []vaildations.Validation {
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)
	devInfo, err := devices.GetPTPDeviceInfo(interfaceName, ctx)
	utils.IfErrorExitOrPanic(err)
	devDetails := vaildations.NewDeviceDetails(&devInfo)
	devFirmware := vaildations.NewDeviceFirmware(&devInfo)
	devDriver := vaildations.NewDeviceDriver(&devInfo)
	return []vaildations.Validation{devDetails, devFirmware, devDriver}
}

func getGNSSValidation(
	clientset *clients.Clientset,
) vaildations.Validation {
	ctx, err := contexts.GetPTPgpsdContext(clientset)
	utils.IfErrorExitOrPanic(err)
	gnssInfo, err := devices.GetGPSNav(ctx)
	utils.IfErrorExitOrPanic(err)
	gnssVersion := vaildations.NewGNSS(&gnssInfo)
	return gnssVersion
}

func getValidations(interfaceName, kubeConfig string) []vaildations.Validation {
	checks := make([]vaildations.Validation, 0)
	clientset, err := clients.GetClientset(kubeConfig)
	utils.IfErrorExitOrPanic(err)
	checks = append(checks, getDevInfoValidations(clientset, interfaceName)...)
	checks = append(checks, vaildations.NewIsGrandMaster(clientset))
	checks = append(checks, vaildations.NewOperatorVersion(clientset))
	checks = append(checks, getGNSSValidation(clientset))
	// OCP Version 4.13 >=
	// an interesting one is the GPSD version minimum acceptable is 3.25
	return checks
}

func reportAnalysertJSON(failures, successes, unknown []*ValidationResult) {
	callback, err := callbacks.SetupCallback("-", callbacks.AnalyserJSON)
	utils.IfErrorExitOrPanic(err)

	for _, unknownCheck := range unknown {
		err := callback.Call(unknownCheck, "env-check-unkown")
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
		reportAnalysertJSON(failures, successes, unknown)
		if len(failures) > 0 {
			os.Exit(int(utils.InvalidEnv))
		}
	} else {
		for _, res := range unknown {
			log.Warn(res.err.Error())
		}
		if len(failures) == 0 {
			fmt.Println("No issues found.") //nolint:forbidigo // This to print out to the user
		} else {
			validationsErrors := make([]error, 0)
			for _, res := range failures {
				validationsErrors = append(validationsErrors, res.err)
			}
			err := utils.MakeCompositeInvalidEnvError(validationsErrors)
			utils.IfErrorExitOrPanic(err)
		}
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
			err:       err,
			valdation: check,
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
