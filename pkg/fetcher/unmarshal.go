// SPDX-License-Identifier: GPL-2.0-or-later

package fetcher

import (
	"fmt"
	"reflect"
)

func setValueOnField(fieldVal reflect.Value, value any) error {
	if valueType := reflect.TypeOf(value); valueType != fieldVal.Type() {
		return fmt.Errorf(
			"incoming value %v with type %s not of expected type %s",
			value,
			valueType,
			fieldVal.Type().String(),
		)
	}
	fieldVal.Set(reflect.ValueOf(value))
	return nil
}

// unmarshal will populate the fields in `target` with the values from `result` according to the fields`fetcherKey` tag.
// fields with no `fetcherKey` tag will not be touched, and elements in `result` without a matched field will be ignored.
func unmarshal(result map[string]any, target any) error {
	val := reflect.ValueOf(target)
	typ := reflect.TypeOf(target)

	for i := 0; i < val.Elem().NumField(); i++ {
		field := typ.Elem().Field(i)
		resultName := field.Tag.Get("fetcherKey")
		if resultName != "" {
			fieldVal := val.Elem().FieldByIndex(field.Index)
			if res, ok := result[resultName]; ok {
				err := setValueOnField(fieldVal, res)
				if err != nil {
					return fmt.Errorf("failed to set value on field %s: %w", resultName, err)
				}
			}
		}
	}
	return nil
}
