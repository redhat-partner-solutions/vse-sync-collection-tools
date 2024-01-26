// SPDX-License-Identifier: GPL-2.0-or-later

package logging

import (
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/utils"
)

// SetupLogging will configure the output stream and the level
// of the logrus logger
func SetupLogging(logLevel string, out io.Writer) {
	log.SetOutput(out)
	level, err := log.ParseLevel(logLevel)
	utils.IfErrorExitOrPanic(err)
	log.SetLevel(level)
}
