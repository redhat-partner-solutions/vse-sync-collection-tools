// SPDX-License-Identifier: GPL-2.0-or-later

package collectors //nolint:dupl // new collector

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/collectors/devices"
)

var (
	GPSCollectorName = "GNSS"
	gpsNavKey        = "gpsNav"
)

type GPSCollector struct {
	*baseCollector

	ctx           clients.ExecContext
	interfaceName string
}

func gpsNavPoller(gps *GPSCollector) func() (callbacks.OutputType, error) {
	return func() (callbacks.OutputType, error) {
		return devices.GetGPSNav(gps.ctx) //nolint:wrapcheck //no point wrapping this
	}
}

// Returns a new GPSCollector based on values in the CollectionConstructor
func NewGPSCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset, constructor.PTPNodeName)
	if err != nil {
		return &GPSCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}

	collector := &GPSCollector{
		baseCollector: newBaseCollector(
			constructor.PollInterval,
			false,
			constructor.Callback,
			GPSCollectorName,
			gpsNavKey,
		),
		ctx:           ctx,
		interfaceName: constructor.PTPInterface,
	}
	collector.poller = gpsNavPoller(collector)

	return collector, nil
}

func init() {
	RegisterCollector(GPSCollectorName, NewGPSCollector, optional)
}
