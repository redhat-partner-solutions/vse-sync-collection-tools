// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

type Validation interface {
	Verify() error
	GetID() string
	GetData() any
}
