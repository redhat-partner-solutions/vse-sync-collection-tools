package vaildations

import "encoding/json"

type VersionWithError struct {
	Version string `json:"version"`
	Error   error  `json:"fetchError"`
}

func MarshalVersionAndError(ver *VersionWithError) ([]byte, error) {
	var err any
	if ver.Error != nil {
		err = ver.Error.Error()
	}
	return json.Marshal(&struct {
		Version string `json:"version"`
		Error   any    `json:"fetchError"`
	}{
		Version: ver.Version,
		Error:   err,
	})
}
