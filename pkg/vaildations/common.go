// SPDX-License-Identifier: GPL-2.0-or-later

package vaildations

import (
	"encoding/json"
	"fmt"

	"golang.org/x/mod/semver"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/utils"
)

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

func checkVersion(version, expected string) error {
	ver := fmt.Sprintf("v%s", version)
	if !semver.IsValid(ver) {
		return fmt.Errorf("could not parse version %s", ver)
	}
	if semver.Compare(ver, fmt.Sprintf("v%s", expected)) < 0 {
		return utils.NewInvalidEnvError(
			fmt.Errorf(
				"invalid version: %s < %s",
				version,
				expected,
			),
		)
	}
	return nil
}
