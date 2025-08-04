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
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/constants"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
)

type DetectedInterface struct {
	Name               string `json:"name"`
	PTPClockDevicePath string `json:"ptp_dev"` //nolint:tagliatelle // script assumes
	Primary            bool   `json:"primary"`
}

// sortAndDeduplicateInterfaces sorts interfaces by Primary (primary first) then by Name (alphabetically),
// and deduplicates based on PTP device path, keeping the first occurrence
func sortAndDeduplicateInterfaces(interfaces []DetectedInterface) []DetectedInterface {
	if len(interfaces) == 0 {
		return interfaces
	}

	// First, deduplicate based on PTP device path
	// Use a map to track seen PTP devices and keep only the first occurrence
	seen := make(map[string]bool)
	deduplicated := make([]DetectedInterface, 0, len(interfaces))

	// Sort by Primary (primary first) then by Name (alphabetically)
	sort.Slice(interfaces, func(i, j int) bool {
		// Primary interfaces come first
		if interfaces[i].Primary != interfaces[j].Primary {
			return interfaces[i].Primary // true comes before false
		}
		// If Primary status is the same, sort by Name alphabetically
		return interfaces[i].Name < interfaces[j].Name
	})

	for _, iface := range interfaces {
		if !seen[iface.PTPClockDevicePath] {
			seen[iface.PTPClockDevicePath] = true
			deduplicated = append(deduplicated, iface)
		} else {
			log.Infof("Deduplicating interface %s with PTP device %s (already seen)",
				iface.Name, iface.PTPClockDevicePath)
		}
	}

	return deduplicated
}

func Detect(kubeConfig, ptpNodeName string, outputAsJSON bool, clockType string) {
	clientset, err := clients.GetClientset(kubeConfig)
	utils.IfErrorExitOrPanic(err)
	ctx, err := contexts.GetPTPDaemonContext(clientset, ptpNodeName)
	utils.IfErrorExitOrPanic(err)
	interfaces, err := checkPTPConfig(ctx, clockType)
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
		return result, fmt.Errorf("failed when parsing config: %w", err)
	}
	return result, nil
}

var ts2phcNotMaster = regexp.MustCompile(`ts2phc.master\s+0`)
var ptp4lMasterOnly = regexp.MustCompile(`masterOnly\s+1`)
var ptp4lServerOnly = regexp.MustCompile(`serverOnly\s+1`)

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
			if ts2phcNotMaster.MatchString(l) {
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

func checkPTPConfig(ctx clients.ExecContext, clockType string) ([]DetectedInterface, error) {
	if clockType == constants.ClockTypeBC {
		// For BC clocks, try ptp4l config first
		interfaces, err := checkPtp4lConfig(ctx)
		if err != nil {
			log.Info("ptp4l config not found, falling back to ts2phc config for BC clock")
			return checkTs2PhcConfig(ctx)
		}
		return interfaces, nil
	} else {
		// For GM clocks, try ts2phc config first
		interfaces, err := checkTs2PhcConfig(ctx)
		if err != nil {
			log.Info("ts2phc config not found, falling back to ptp4l config for GM clock")
			return checkPtp4lConfig(ctx)
		}
		return interfaces, nil
	}
}

func checkPtp4lConfig(ctx clients.ExecContext) ([]DetectedInterface, error) {
	errs := []error{}
	detected := []DetectedInterface{}
	files, _, err := ctx.ExecCommand([]string{"ls", "/var/run/"})
	if err != nil {
		return nil, fmt.Errorf("failed to list /var/run/ directory: %w", err)
	}

	ptp4lConfigFiles := make([]string, 0)
	for _, f := range strings.Fields(files) {
		if strings.HasPrefix(f, "ptp4l.") && strings.HasSuffix(f, ".config") {
			ptp4lConfigFiles = append(ptp4lConfigFiles, f)
		}
	}
	if len(ptp4lConfigFiles) == 0 {
		return nil, errors.New("failed to find ptp4l config file")
	} else if len(ptp4lConfigFiles) > 1 {
		log.Warnf("Multiple ptp4l profiles found (%v)", ptp4lConfigFiles)
	}

	for _, ptp4lConfigPath := range ptp4lConfigFiles {
		ptp4lConfig, _, err := ctx.ExecCommand([]string{"cat", "/var/run/" + ptp4lConfigPath})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read ptp4l config file: %w", err))
			continue
		}
		config, err := parseConfig(ptp4lConfig)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse ptp4l config file: %w", err))
			continue
		}

		detected = append(detected, getDetectedInterfacesFromPtp4l(ctx, config)...)
	}
	return detected, utils.MakeCompositeError("", errs) //nolint:wrapcheck //this just combines errors.
}

func getDetectedInterfacesFromPtp4l(ctx clients.ExecContext, config map[string][]string) []DetectedInterface {
	detected := []DetectedInterface{}
	for section, lines := range config {
		if section == "global" || section == "nmea" { //nolint:goconst // only one time so const would obfuscate
			continue
		}

		// For BC clocks, all interfaces are primary (they participate in PTP sync)
		isPrimary := true
		for _, l := range lines {
			if ptp4lMasterOnly.MatchString(l) || ptp4lServerOnly.MatchString(l) {
				isPrimary = false
			}
		}

		ptpDev, err := getPTPClockDevice(ctx, section)
		if err != nil {
			log.Warnf("Failed to get PTP clock device for interface %s: %v", section, err)
			continue
		}

		detected = append(detected, DetectedInterface{
			Name:               strings.TrimSpace(section),
			Primary:            isPrimary,
			PTPClockDevicePath: ptpDev,
		})
	}
	return sortAndDeduplicateInterfaces(detected)
}

func checkTs2PhcConfig(ctx clients.ExecContext) ([]DetectedInterface, error) { //nolint:stylecheck //Suggestion looks bad
	errs := []error{}
	detected := []DetectedInterface{}
	files, _, err := ctx.ExecCommand([]string{"ls", "/var/run/"})
	if err != nil {
		return nil, fmt.Errorf("failed to list /var/run/ directory: %w", err)
	}

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
