// SPDX-License-Identifier: GPL-2.0-or-later

package verify

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/contexts"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/collectors/devices"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/vaildations"
)

func Verify(interfaceName, kubeConfig string) {
	clientset, err := clients.GetClientset(kubeConfig)
	utils.IfErrorExitOrPanic(err)
	ctx, err := contexts.GetPTPDaemonContext(clientset)
	utils.IfErrorExitOrPanic(err)
	devInfo, err := devices.GetPTPDeviceInfo(interfaceName, ctx)
	utils.IfErrorExitOrPanic(err)
	err = vaildations.VerifyDeviceInfo(&devInfo)
	utils.IfErrorExitOrPanic(err)
	fmt.Println("No issues found.") //nolint:forbidigo // This to print out to the user
}
