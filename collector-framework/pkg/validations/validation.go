// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
)

type Validation interface {
	Verify() error
	GetID() string
	GetDescription() string
	GetData() any
	GetOrder() int
}

//nolint:varnamelen // id is fine this seems like a bug in varnamelen
func NewVersionCheck(
	id string,
	version string,
	checkVersion string,
	minVersion string,
	description string,
	order int,
) *VersionCheck {
	return &VersionCheck{
		id:           id,
		Version:      version,
		checkVersion: checkVersion,
		MinVersion:   minVersion,
		description:  description,
		order:        order,
	}
}

type VersionCheck struct {
	id           string `json:"-"`
	Version      string `json:"version"`
	checkVersion string `json:"-"`
	MinVersion   string `json:"expected"`
	description  string `json:"-"`
	order        int    `json:"-"`
}

func (verCheck *VersionCheck) Verify() error {
	ver := fmt.Sprintf("v%s", strings.ReplaceAll(verCheck.checkVersion, "_", "-"))
	if !semver.IsValid(ver) {
		return fmt.Errorf("could not parse version %s", ver)
	}
	if semver.Compare(ver, fmt.Sprintf("v%s", verCheck.MinVersion)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf("unexpected version: %s < %s", verCheck.checkVersion, verCheck.MinVersion),
		)
	}
	return nil
}

func (verCheck *VersionCheck) GetID() string {
	return verCheck.id
}

func (verCheck *VersionCheck) GetDescription() string {
	return verCheck.description
}

func (verCheck *VersionCheck) GetData() any { //nolint:ireturn // data will vary for each validation
	return verCheck
}

func (verCheck *VersionCheck) GetOrder() int {
	return verCheck.order
}

//nolint:varnamelen // id is fine this seems like a bug in varnamelen
func NewVersionCheckWithError(
	id string,
	version string,
	checkVersion string,
	minVersion string,
	description string,
	order int,
	err error,
) *VersionWithErrorCheck {
	return &VersionWithErrorCheck{
		VersionCheck: &VersionCheck{
			id:           id,
			Version:      version,
			checkVersion: checkVersion,
			MinVersion:   minVersion,
			description:  description,
			order:        order,
		},
		Error: err,
	}
}

type VersionWithError struct {
	Error   error  `json:"fetchError"`
	Version string `json:"version"`
}

func MarshalVersionAndError(ver *VersionWithError) ([]byte, error) {
	var err any
	if ver.Error != nil {
		err = ver.Error.Error()
	}
	marsh, marshalErr := json.Marshal(&struct {
		Error   any    `json:"fetchError"`
		Version string `json:"version"`
	}{
		Version: ver.Version,
		Error:   err,
	})
	return marsh, fmt.Errorf("failed to marshal VersionWithError %w", marshalErr)
}

type VersionWithErrorCheck struct {
	*VersionCheck
	Error error
}

func (verCheck *VersionWithErrorCheck) MarshalJSON() ([]byte, error) {
	return MarshalVersionAndError(&VersionWithError{
		Version: verCheck.Version,
		Error:   verCheck.Error,
	})
}

func (verCheck *VersionWithErrorCheck) Verify() error {
	if verCheck.Error != nil {
		return verCheck.Error
	}
	return verCheck.VersionCheck.Verify()
}
