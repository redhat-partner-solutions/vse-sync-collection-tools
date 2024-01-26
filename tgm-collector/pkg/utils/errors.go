// SPDX-License-Identifier: GPL-2.0-or-later

package utils

import (
	"errors"
	"fmt"
	"strings"
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

type RequirementsNotMetError struct {
	err error
}

func (err RequirementsNotMetError) Error() string {
	return err.err.Error()
}
func (err RequirementsNotMetError) Unwrap() error {
	return err.err
}

func NewRequirementsNotMetError(err error) *RequirementsNotMetError {
	return &RequirementsNotMetError{err: err}
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

func MakeCompositeError(prefix string, errSlice []error) error {
	pattern := strings.Repeat("\t%s\n", len(errSlice))

	values := make([]any, 0)
	for _, err := range errSlice {
		values = append(values, err.Error())
	}
	if len(prefix) > 0 {
		return fmt.Errorf(prefix+":\n"+pattern, values...)
	}
	return fmt.Errorf(pattern, values...)
}

func MakeCompositeInvalidEnvError(errSlice []error) error {
	return NewInvalidEnvError(
		MakeCompositeError("The following issues where found", errSlice),
	)
}
