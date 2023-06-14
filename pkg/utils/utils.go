// SPDX-License-Identifier: GPL-2.0-or-later

package utils

import (
	log "github.com/sirupsen/logrus"
)

func IfErrorPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}
