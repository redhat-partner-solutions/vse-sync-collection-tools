// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"errors"
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/validations"
)

type resType int

const (
	resTypeUnknown resType = iota
	resTypeSuccess
	resTypeFailure
)

type ValidationResult struct {
	validation validations.Validation
	err        error
	resType    resType
}

func (res *ValidationResult) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	var result any
	msg := ""

	switch res.resType {
	case resTypeSuccess:
		result = true
	case resTypeFailure:
		result = false
		msg = res.err.Error()
	case resTypeUnknown:
		result = "error" //nolint:goconst // making it a constant would be an obfuscation
		msg = res.err.Error()
	}

	formatted := callbacks.AnalyserFormatType{
		ID: "environment-check",
		Data: map[string]any{
			"id":       res.validation.GetID(),
			"result":   result,
			"reason":   msg,
			"analysis": res.validation.GetData(),
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

func isInvalidEnv(err error) bool {
	if err != nil {
		var invalidEnv *utils.InvalidEnvError
		if errors.As(err, &invalidEnv) {
			return true
		}
	}
	return false
}

func (res *ValidationResult) GetPrefixedError() error {
	return fmt.Errorf("%s: %w", res.validation.GetDescription(), res.err)
}

func NewValidationResult(validation validations.Validation) *ValidationResult {
	result := resTypeUnknown
	err := validation.Verify()
	if err != nil {
		if isInvalidEnv(err) {
			result = resTypeFailure
		}
	} else {
		result = resTypeSuccess
	}

	return &ValidationResult{
		validation: validation,
		err:        err,
		resType:    result,
	}
}
