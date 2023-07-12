// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"context"
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

const isConfiguredForGrandMaster = "Card driver is valid"

type IsGrandMaster struct {
	client *clients.Clientset
}

func (dev *IsGrandMaster) Verify() error {
	data, err := dev.client.K8sRestClient.Get().
		AbsPath("/apis/ptp.openshift.io/v1").
		Namespace("openshift-ptp").
		Resource("ptpconfigs").
		DoRaw(context.TODO())
	fmt.Println(data, err)
	return nil
}

func (dev *IsGrandMaster) GetID() string {
	return isConfiguredForGrandMaster
}

func (dev *IsGrandMaster) GetData() any { //nolint:ireturn // data will very for each validation
	return dev
}

func NewIsGrandMaster(client *clients.Clientset) *IsGrandMaster {
	return &IsGrandMaster{
		client: client,
	}
}
