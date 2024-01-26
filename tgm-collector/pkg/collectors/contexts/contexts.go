// SPDX-License-Identifier: GPL-2.0-or-later

package contexts

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/clients"
)

const (
	PTPNamespace               = "openshift-ptp"
	PTPPodNamePrefix           = "linuxptp-daemon-"
	PTPContainer               = "linuxptp-daemon-container"
	GPSContainer               = "gpsd"
	NetlinkDebugPod            = "ptp-dpll-netlink-debug-pod"
	NetlinkDebugContainer      = "ptp-dpll-netlink-debug-container"
	NetlinkDebugContainerImage = "quay.io/redhat-partner-solutions/dpll-debug:0.1"
)

func GetPTPDaemonContext(clientset *clients.Clientset) (*clients.ContainerExecContext, error) {
	ctx, err := clients.NewContainerContext(clientset, PTPNamespace, PTPPodNamePrefix, PTPContainer)
	if err != nil {
		return ctx, fmt.Errorf("could not create container context %w", err)
	}
	return ctx, nil
}

func GetNetlinkContext(clientset *clients.Clientset) (*clients.ContainerCreationExecContext, error) {
	hpt := corev1.HostPathDirectory
	ctx := clients.NewContainerCreationExecContext(
		clientset,
		PTPNamespace,
		NetlinkDebugPod,
		NetlinkDebugContainer,
		NetlinkDebugContainerImage,
		map[string]string{},
		[]string{"sleep", "inf"},
		&corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				// Requires NET_ADMIN: having (NET_RAW + NET_BIND_SERVICE + NET_BROADCAST) does not work
				//     Without NET_ADMIN it will not connect to the netlink interface
				// Requires SYS_AMDIN: having every other permission does not work.
				//     Without SYS_ADMIN lspci does not include the Serial number in the comments thefore can not calculate the clockID
				Add: []corev1.Capability{
					"SYS_ADMIN",
					"NET_ADMIN",
				},
			},
		},
		true,
		[]*clients.Volume{
			{
				Name:         "modules",
				MountPath:    "/lib/modules",
				VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/lib/modules", Type: &hpt}},
			},
		},
	)
	return ctx, nil
}
