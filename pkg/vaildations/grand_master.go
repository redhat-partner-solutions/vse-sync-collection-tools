// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

const (
	isConfiguredForGrandMaster = "Configured for grand master"
	gmFlag                     = "ts2phc.master 1"
)

type IsGrandMaster struct {
	client   *clients.Clientset `json:"-"`
	Profiles []PTPConfigProfile `json:"profiles"`
}

type PTPConfigProfile struct {
	TS2PhcConf string `json:"ts2phcConf"`
}

type PTPConfigSpec struct {
	Profiles []PTPConfigProfile `json:"profile"`
}
type PTPConfig struct {
	Spec PTPConfigSpec `json:"spec"`
}

type PTPConfigList struct {
	APIVersion string      `json:"apiVersion"`
	Items      []PTPConfig `json:"items"`
}

func FetchPTPConfigs(client *clients.Clientset) (PTPConfigList, error) {
	data, err := client.K8sRestClient.Get().
		AbsPath("/apis/ptp.openshift.io/v1").
		Namespace("openshift-ptp").
		Resource("ptpconfigs").
		DoRaw(context.TODO())

	if err != nil {
		return PTPConfigList{}, fmt.Errorf("failed to fetch ptpconfigs %w", err)
	}

	unpacked := &PTPConfigList{}
	err = json.Unmarshal(data, unpacked)
	if err != nil {
		return PTPConfigList{}, fmt.Errorf("failed to unmarshal ptpconfigs %w", err)
	}
	return *unpacked, nil
}

func (gm *IsGrandMaster) Verify() error {
	ptpConfigList, err := FetchPTPConfigs(gm.client)
	if err != nil {
		return err
	}
	for _, item := range ptpConfigList.Items {
		gm.Profiles = append(gm.Profiles, item.Spec.Profiles...)
		for _, profile := range item.Spec.Profiles {
			if strings.Contains(profile.TS2PhcConf, gmFlag) {
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
