// SPDX-License-Identifier: GPL-2.0-or-later

package fetcher

import (
	"bytes"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/clients"
)

type PostProcessFuncType func(map[string]string) (map[string]any, error)

type Fetcher struct {
	cmdGrp        *clients.CmdGroup
	postProcessor PostProcessFuncType
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		cmdGrp: &clients.CmdGroup{},
	}
}

func TrimSpace(s string) (string, error) {
	return strings.TrimSpace(s), nil
}

func (inst *Fetcher) SetPostProcessor(ppFunc PostProcessFuncType) {
	inst.postProcessor = ppFunc
}

// AddNewCommand creates a new command from a string
// then adds it to the fetcher
func (inst *Fetcher) AddNewCommand(key, cmd string, trim bool) error {
	cmdInst, err := clients.NewCmd(key, cmd)
	if err != nil {
		return fmt.Errorf("add fetcher cmd failed %w", err)
	}
	if trim {
		cmdInst.SetOutputProcessor(TrimSpace)
	}
	inst.cmdGrp.AddCommand(cmdInst)
	return nil
}

// AddCommand takes a command instance and adds it the fetcher
func (inst *Fetcher) AddCommand(cmdInst *clients.Cmd) {
	inst.cmdGrp.AddCommand(cmdInst)
}

// Fetch executes the commands on the container passed as the ctx and
// use the results to populate pack
func (inst *Fetcher) Fetch(ctx clients.ExecContext, pack any) error {
	runResult, err := runCommands(ctx, inst.cmdGrp)
	if err != nil {
		return err
	}
	result := make(map[string]any)
	for key, value := range runResult {
		result[key] = value
	}
	if inst.postProcessor != nil {
		updatedResults, ppErr := inst.postProcessor(runResult)
		if ppErr != nil {
			return fmt.Errorf("feching failed post process the data %w", ppErr)
		}
		for key, value := range updatedResults {
			result[key] = value
		}
	}
	err = unmarshal(result, pack)
	if err != nil {
		return fmt.Errorf("feching failed to unpack data %w", err)
	}
	return nil
}

// runCommands executes the commands on the container passed as the ctx
// and extracts the results from the stdout
func runCommands(ctx clients.ExecContext, cmdGrp clients.Cmder) (result map[string]string, err error) { //nolint:lll // allow slightly long function definition
	cmd := cmdGrp.GetCommand()
	command := []string{"/usr/bin/sh"}
	var buffIn bytes.Buffer
	buffIn.WriteString(cmd)

	stdout, _, err := ctx.ExecCommandStdIn(command, buffIn)
	if err != nil {
		log.Debugf(
			"command in container failed unexpectedly:\n\tcontext: %v\n\tcommand: %v\n\terror: %v",
			ctx, command, err,
		)
		return result, fmt.Errorf("runCommands failed %w", err)
	}
	result, err = cmdGrp.ExtractResult(stdout)
	if err != nil {
		log.Debugf("extraction failed %s", err.Error())
		log.Debugf("output was %s", stdout)
		return result, fmt.Errorf("runCommands failed %w", err)
	}
	return result, nil
}
