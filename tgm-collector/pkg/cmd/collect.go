// SPDX-License-Identifier: GPL-2.0-or-later

package cmd

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	fCmd "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/cmd"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/runner"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors"
)

const (
	defaultIncludeLogTimestamps bool   = false
	defaultTempDir              string = "."
	defaultKeepDebugFiles       bool   = false
	tempdirPerm                        = 0755
)

type CollectorArgFunc func() map[string]map[string]any
type CheckVarsFunc func()

var (
	logsOutputFile       string
	includeLogTimestamps bool
	tempDir              string
	keepDebugFiles       bool
)

func init() { //nolint:funlen // Allow this to get a little long
	AddInterfaceFlag(fCmd.CollectCmd)

	fCmd.CollectCmd.Flags().StringVarP(
		&logsOutputFile,
		"logs-output", "l", "",
		"Path to the logs output file. This is required when using the logs collector",
	)
	fCmd.CollectCmd.Flags().BoolVar(
		&includeLogTimestamps,
		"log-timestamps", defaultIncludeLogTimestamps,
		"Specifies if collected logs should include timestamps or not. (default is false)",
	)
	fCmd.CollectCmd.Flags().StringVarP(&tempDir, "tempdir", "t", defaultTempDir,
		"Directory for storing temp/debug files. Must exist.")
	fCmd.CollectCmd.Flags().BoolVar(&keepDebugFiles, "keep", defaultKeepDebugFiles, "Keep debug files")

	fCmd.SetCollecterArgsFunc(func() map[string]map[string]any {
		collectorArgs := make(map[string]map[string]any)
		collectorArgs["PTP"] = map[string]any{
			"ptpInterface": ptpInterface,
		}
		collectorArgs["Logs"] = map[string]any{
			"logsOutputFile":       logsOutputFile,
			"includeLogTimestamps": includeLogTimestamps,
			"tempDir":              tempDir,
			"keepDebugFiles":       keepDebugFiles,
		}
		return collectorArgs
	})

	fCmd.SetCheckVarsFunc(func(collectorNames []string) {
		for _, c := range collectorNames {
			if (c == collectors.LogsCollectorName || c == runner.All) && logsOutputFile == "" {
				utils.IfErrorExitOrPanic(utils.NewMissingInputError(
					errors.New("if Logs collector is selected you must also provide a log output file")),
				)
			}
		}

		if strings.Contains(tempDir, "~") {
			usr, err := user.Current()
			if err != nil {
				log.Fatal("Failed to fetch current user so could not resolve tempdir")
			}
			if tempDir == "~" {
				tempDir = usr.HomeDir
			} else if strings.HasPrefix(tempDir, "~/") {
				tempDir = filepath.Join(usr.HomeDir, tempDir[2:])
			}
		}

		if err := os.MkdirAll(tempDir, tempdirPerm); err != nil {
			log.Fatal(err)
		}
	})
}
