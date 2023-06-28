// SPDX-License-Identifier: GPL-2.0-or-later

package utils

import (
	"errors"
)

type exitCode int

const (
	// Exitcodes
	Success exitCode = iota
	InvalidEnv
	MissingInput
	NotHandled
)

type InvalidEnvError struct {
	err error
}

func (err InvalidEnvError) Error() string {
	return err.err.Error()
}
func (err InvalidEnvError) Unwrap() error {
	return err.err
}

func NewInvalidEnvError(err error) *InvalidEnvError {
	return &InvalidEnvError{err: err}
}

type MissingInputError struct {
	err error
}

func (err MissingInputError) Error() string {
	return err.err.Error()
}
func (err MissingInputError) Unwrap() error {
	return err.err
}

func NewMissingInputError(err error) *MissingInputError {
	return &MissingInputError{err: err}
}

func checkError(err error) (exitCode, bool) {
	var invalidEnv *InvalidEnvError
	if errors.As(err, &invalidEnv) {
		return InvalidEnv, true
	}

	var missingInput *MissingInputError
	if errors.As(err, &missingInput) {
		return MissingInput, true
	}

	return NotHandled, false
}
