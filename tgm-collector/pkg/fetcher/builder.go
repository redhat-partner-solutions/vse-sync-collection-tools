// SPDX-License-Identifier: GPL-2.0-or-later

package fetcher

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
)

type AddCommandArgs struct {
	Key     string
	Command string
	Trim    bool
}

func FetcherFactory(commands []*clients.Cmd, addCommands []AddCommandArgs) (*Fetcher, error) {
	fetcherInst := NewFetcher()
	for _, cmdInst := range commands {
		fetcherInst.AddCommand(cmdInst)
	}

	for _, addCmd := range addCommands {
		err := fetcherInst.AddNewCommand(addCmd.Key, addCmd.Command, addCmd.Trim)
		if err != nil {
			return fetcherInst, fmt.Errorf("failed to add command %s: %w", addCmd.Key, err)
		}
	}

	return fetcherInst, nil
}
