// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/utils"
)

const (
	configuredForGrandMaster            = TGMSyncEnvPath + "/ptp-operator/"
	configuredForGrandMasterDescription = "Configured for grand master"
	gmFlag                              = "ts2phc.master 1"
)

type GMProfiles struct {
	Error    error              `json:"fetchError"`
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

func fetchPTPConfigs(client *clients.Clientset) (PTPConfigList, error) {
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

func (gm *GMProfiles) Verify() error {
	if gm.Error != nil {
		return gm.Error
	}
	for _, profile := range gm.Profiles {
		if strings.Contains(profile.TS2PhcConf, gmFlag) {
			return nil
		}
	}

	return utils.NewInvalidEnvError(errors.New("no configuration for Grand Master clock"))
}

func (gm *GMProfiles) GetID() string {
	return configuredForGrandMaster
}

func (gm *GMProfiles) GetDescription() string {
	return configuredForGrandMasterDescription
}

func (gm *GMProfiles) GetData() any { //nolint:ireturn // data will vary for each validation
	return gm
}

func (gm *GMProfiles) GetOrder() int {
	return configuredForGrandMasterOrdering
}

func NewIsGrandMaster(client *clients.Clientset) *GMProfiles {
	ptpConfigList, err := fetchPTPConfigs(client)
	gmProfiles := &GMProfiles{
		Error: err,
	}
	if err != nil {
		return gmProfiles
	}
	for _, item := range ptpConfigList.Items {
		gmProfiles.Profiles = append(gmProfiles.Profiles, item.Spec.Profiles...)
	}
	return gmProfiles
}
