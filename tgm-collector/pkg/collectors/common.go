// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"errors"

	collectorsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/collectors"
)

//nolint:varnamelen // ok is the idomatic name for this var
func getPTPInterfaceName(constructor *collectorsBase.CollectionConstructor) (string, error) {
	ptpArgs, ok := constructor.CollectorArgs["PTP"]
	if !ok {
		return "", errors.New("no PTP args in collector args")
	}
	ptpInterfaceRaw, ok := ptpArgs["ptpInterface"]
	if !ok {
		return "", errors.New("no ptpInterface in PTP collector args")
	}

	ptpInterface, ok := ptpInterfaceRaw.(string)
	if !ok {
		return "", errors.New("PTP interface is not a string")
	}
	return ptpInterface, nil
}

func LinkMe() {}
