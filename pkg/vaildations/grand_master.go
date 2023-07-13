// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	isConfiguredForGrandMaster = "Card driver is valid"
	gmFlag                     = "ts2phc.master 1"
)

type IsGrandMaster struct {
	Profiles []PTPConfigProfile `json:"profiles"`
	client   *clients.Clientset `json:"-"`
}

type PTPConfigProfile struct {
	Ts2PhcConf string `json:"ts2phcConf"`
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

func FetchPTPConfigs(client *clients.Clientset) PTPConfigList {
	data, _ := client.K8sRestClient.Get().
		AbsPath("/apis/ptp.openshift.io/v1").
		Namespace("openshift-ptp").
		Resource("ptpconfigs").
		DoRaw(context.TODO())

	unpacked := &PTPConfigList{}
	json.Unmarshal(data, unpacked)
	return *unpacked
}

func (gm *IsGrandMaster) Verify() error {
	ptpConfigList := FetchPTPConfigs(gm.client)
	for _, item := range ptpConfigList.Items {
		gm.Profiles = append(gm.Profiles, item.Spec.Profiles...)
		for _, profile := range item.Spec.Profiles {
			if strings.Contains(profile.Ts2PhcConf, gmFlag) {
				return nil
			}
		}
	}
	return utils.NewInvalidEnvError(errors.New("no configuration for Grand Master clock"))
}

func (gm *IsGrandMaster) GetID() string {
	return isConfiguredForGrandMaster
}

func (gm *IsGrandMaster) GetData() any { //nolint:ireturn // data will very for each validation
	return gm
}

func NewIsGrandMaster(client *clients.Clientset) *IsGrandMaster {
	return &IsGrandMaster{
		client: client,
	}
}
