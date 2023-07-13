// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"errors"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/vaildations"
)

type ValidationResult struct {
	valdation vaildations.Validation
	err       error
}

func (res *ValidationResult) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	msg := ""
	result := "unknown"
	if res.err != nil {
		msg = res.err.Error()
		invalidEnv := &utils.InvalidEnvError{}
		if errors.As(res.err, invalidEnv) {
			result = "false"
		}
	} else {
		result = "true"
	}

	formatted := callbacks.AnalyserFormatType{
		ID: "environment-check",
		Data: []any{
			res.valdation.GetID(),
			result,
			msg,
			res.valdation.GetData(),
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}

func (res *ValidationResult) IsInvalidEnv() bool {
	if res.err != nil {
		invalidEnv := &utils.InvalidEnvError{}
		if errors.As(res.err, invalidEnv) {
			return true
		}
	}
	return false
}
