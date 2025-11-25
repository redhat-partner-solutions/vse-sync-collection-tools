// SPDX-License-Identifier: GPL-2.0-or-later

package detect

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type DetectedInterface struct {
	Name               string `json:"name"`
	PTPClockDevicePath string `json:"ptp_dev"` //nolint:tagliatelle // script assumes
	Primary            bool   `json:"primary"`
}

func Detect(kubeConfig, ptpNodeName string, outputAsJSON bool) {
	clientset, err := clients.GetClientset(kubeConfig)
	utils.IfErrorExitOrPanic(err)
	ctx, err := contexts.GetPTPDaemonContext(clientset, ptpNodeName)
	utils.IfErrorExitOrPanic(err)
	interfaces, err := checkTs2PhcConfig(ctx)
	utils.IfErrorExitOrPanic(err)
	output(os.Stdout, interfaces, outputAsJSON)
}

func output(outWriter io.Writer, interfaces []DetectedInterface, outputAsJSON bool) {
	if outputAsJSON {
		out, err := json.MarshalIndent(interfaces, "", "  ")
		utils.IfErrorExitOrPanic(err)
		_, err = outWriter.Write(out)
		utils.IfErrorExitOrPanic(err)
	} else {
		_, err := fmt.Fprintf(outWriter, "%T(%v)", interfaces, interfaces)
		utils.IfErrorExitOrPanic(err)
	}
}

func parseConfig(contents string) (map[string][]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(contents))
	var currentSection string
	result := make(map[string][]string)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		} else if currentSection != "" {
			result[currentSection] = append(result[currentSection], line)
		}
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("failed when parsing ts2phc config: %w", err)
	}
	return result, nil
}

var notMaster = regexp.MustCompile(`ts2phc.master\s+0`)

func getPTPClockDevice(ctx clients.ExecContext, interfaceName string) (string, error) {
	out, _, err := ctx.ExecCommand([]string{"ethtool", "-T", interfaceName})
	if err != nil {
		return "", fmt.Errorf("failed to get ptp clock number: %w", err)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "PTP Hardware Clock:") {
			clockNumber := strings.TrimSpace(strings.Split(line, ":")[1])
			return fmt.Sprintf("/dev/ptp%s", clockNumber), nil
		}
	}
	return "", errors.New("no PTP clock device found")
}

func getDetectedInterfaces(ctx clients.ExecContext, config map[string][]string) []DetectedInterface {
	detected := []DetectedInterface{}
	for section, lines := range config {
		if section == "global" || section == "nmea" { //nolint:goconst // only one time so const would obfuscate
			continue
		}

		isPrimary := true
		for _, l := range lines {
			if notMaster.MatchString(l) {
				isPrimary = false
			}
		}

		ptpDev, err := getPTPClockDevice(ctx, section)
		utils.IfErrorExitOrPanic(err)

		detected = append(detected, DetectedInterface{
			Name:               strings.TrimSpace(section),
			Primary:            isPrimary,
			PTPClockDevicePath: ptpDev,
		})
	}
	return detected
}

func checkTs2PhcConfig(ctx clients.ExecContext) ([]DetectedInterface, error) { //nolint:staticcheck //Suggestion looks bad
	errs := []error{}
	detected := []DetectedInterface{}
	files, _, err := ctx.ExecCommand([]string{"ls", "/var/run/"})
	utils.IfErrorExitOrPanic(err)

	ts2phcConfigFiles := make([]string, 0)
	for _, f := range strings.Fields(files) {
		if strings.HasPrefix(f, "ts2phc.") && strings.HasSuffix(f, ".config") {
			ts2phcConfigFiles = append(ts2phcConfigFiles, f)
		}
	}
	if len(ts2phcConfigFiles) == 0 {
		return nil, errors.New("failed to find ts2phc config file")
	} else if len(ts2phcConfigFiles) > 1 {
		log.Warnf("Multiple profiles found (%v)", ts2phcConfigFiles)
	}

	for _, ts2phcConfigPath := range ts2phcConfigFiles {
		ts2phcConfig, _, err := ctx.ExecCommand([]string{"cat", "/var/run/" + ts2phcConfigPath})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read ts2 config file: %w", err))
		}
		config, err := parseConfig(ts2phcConfig)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse ts2 config file: %w", err))
		}

		detected = append(detected, getDetectedInterfaces(ctx, config)...)
	}
	return detected, utils.MakeCompositeError("", errs) //nolint:wrapcheck //this just combines errors.
}
