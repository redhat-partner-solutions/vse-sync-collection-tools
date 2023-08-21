// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"errors"
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/validations"
)

const (
	resUnk   = "error"
	resTrue  = true
	resFalse = false
)

type ValidationResult struct {
	validation validations.Validation
	err        error
}

func (res *ValidationResult) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	var result any
	msg := ""
	result = resUnk
	if res.err != nil {
		msg = res.err.Error()
		if res.IsInvalidEnv() {
			result = resFalse
		}
	} else {
		result = resTrue
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

func (res *ValidationResult) IsInvalidEnv() bool {
	if res.err != nil {
		var invalidEnv *utils.InvalidEnvError
		if errors.As(res.err, &invalidEnv) {
			return true
		}
	}
	return false
}

func (res *ValidationResult) GetPrefixedError() error {
	return fmt.Errorf("%s: %w", res.validation.GetDescription(), res.err)
}
