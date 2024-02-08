// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	collectorsBase "github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/collectors"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/collectors/contexts"
)

var LogsCollectorName = collectorsBase.LogsCollectorName

func NewPTPLogsCollector(constructor *collectorsBase.CollectionConstructor) (collectorsBase.Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset)
	if err != nil {
		return &collectorsBase.LogsCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}
	logCollector, err := collectorsBase.NewLogsCollector(constructor, ctx)
	if err != nil {
		err = fmt.Errorf("failed to create ptp log collector: %w", err)
	}
	return logCollector, err
}

func init() {
	// Make log opt in as in may lose some data.
	collectorsBase.RegisterCollector(LogsCollectorName, NewPTPLogsCollector, collectorsBase.Optional)
}
