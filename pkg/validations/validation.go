// SPDX-License-Identifier: GPL-2.0-or-later

package validations

type Validation interface {
	Verify() error
	GetID() string
	GetData() any
}
