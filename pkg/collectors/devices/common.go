// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func failedToAddCommand(fetcherName, cmdKey string, err error) error {
	log.Errorf("failed to add command %s %s", cmdKey, err.Error())
	return fmt.Errorf("failed to add command %w to %s fetcher", err, fetcherName)
}
