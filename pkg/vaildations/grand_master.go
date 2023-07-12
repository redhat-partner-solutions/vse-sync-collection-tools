// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

const isConfiguredForGrandMaster = "Card driver is valid"

type IsGrandMaster struct {
	client *clients.Clientset
}

type PTPConfigProfile struct {
	Ptp4lConf string `json:"ptp4lConf"`
}

type PTPConfigSpec struct {
	Profiles []PTPConfigProfile `json:"profile"`
}
type PTPConfig struct {
	Spec PTPConfigSpec `json:"spec"`
}

type PTPConfigList struct {
	ApiVersion string      `json:"apiVersion"`
	Items      []PTPConfig `json:"items"`
}

func FetchPTPConfigs(client *clients.Clientset) {

}

func (dev *IsGrandMaster) Verify() error {
	data, _ := dev.client.K8sRestClient.Get().
		AbsPath("/apis/ptp.openshift.io/v1").
		Namespace("openshift-ptp").
		Resource("ptpconfigs").
		DoRaw(context.TODO())

	unpacked := &PTPConfigList{}
	json.Unmarshal(data, unpacked)
	fmt.Println(unpacked)
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
