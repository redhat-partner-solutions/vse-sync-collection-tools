// SPDX-License-Identifier: GPL-2.0-or-later

package contexts

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
)

const (
	PTPNamespace     = "openshift-ptp"
	PTPPodNamePrefix = "linuxptp-daemon-"
	PTPContainer     = "linuxptp-daemon-container"
	GPSContainer     = "gpsd"
)

func GetPTPDaemonContext(clientset *clients.Clientset) (clients.ContainerContext, error) {
	ctx, err := clients.NewContainerContext(clientset, PTPNamespace, PTPPodNamePrefix, PTPContainer)
	if err != nil {
		return ctx, fmt.Errorf("could not create container context %w", err)
	}
	return ctx, nil
}
