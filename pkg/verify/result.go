// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/vaildations"
)

type ValidationResult struct {
	valdation vaildations.Validation
	err       error
}

func (res *ValidationResult) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	msg := ""
	if res.err != nil {
		msg = res.err.Error()
	}
	formatted := callbacks.AnalyserFormatType{
		ID: "environment-check",
		Data: []any{
			res.valdation.GetID(),
			res.err == nil,
			msg,
			res.valdation.GetData(),
		},
	}
	return []*callbacks.AnalyserFormatType{&formatted}, nil
}
