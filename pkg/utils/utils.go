// SPDX-License-Identifier: GPL-2.0-or-later

package utils

func IfErrorPanic(err error) {
	if err != nil {
		panic(err)
	}
}
