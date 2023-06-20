// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type fetcher struct {
	cmdGrp        *clients.CmdGroup
	postProcesser func(map[string]string) (map[string]string, error)
}

func NewFetcher() *fetcher {
	return &fetcher{
		cmdGrp: &clients.CmdGroup{},
	}
}

func (inst *fetcher) SetPostProcesser(ppFunc func(map[string]string) (map[string]string, error)) {
	inst.postProcesser = ppFunc
}

// AddNewCommand creates a new command from a string
// then adds it to the fetcher
func (inst *fetcher) AddNewCommand(key, cmd string, trim bool) error {
	cmdInst, err := clients.NewCmd(key, cmd)
	if err != nil {
		return fmt.Errorf("add fetcher cmd failed %w", err)
	}
	if trim {
		cmdInst.SetCleanupFunc(strings.TrimSpace)
	}
	inst.cmdGrp.AddCommand(cmdInst)
	return nil
}

// AddCommand takes a command instance and adds it the fetcher
func (inst *fetcher) AddCommand(cmdInst *clients.Cmd) {
	inst.cmdGrp.AddCommand(cmdInst)
}

// unmarshall will populate the fields in `pack` with the values from `result` according to the fields`fetcherKey` tag.
// fields with no `fetcherKey` tag will not be touched, and elements in `result` without a matched field will be ignored.
func unmarshall(result map[string]string, pack interface{}) error {
	val := reflect.ValueOf(pack)
	typ := reflect.TypeOf(pack)

	for i := 0; i < val.Elem().NumField(); i++ {
		field := typ.Elem().Field(i)
		resultName := field.Tag.Get("fetcherKey")
		if resultName != "" {
			f := val.Elem().FieldByIndex(field.Index)
			//nolint:exhaustive //we could extend this but its not needed yet
			switch field.Type.Kind() {
			case reflect.String:
				if res, ok := result[resultName]; ok {
					f.SetString(res)
				}
			default:
				return fmt.Errorf("fetcher unmarshal not implmented for type: %s", field.Type.Name())
			}
		}
	}
	return nil
}

// Fetch executes the commands on the container passed as the ctx and
// use the results to populate pack
func (inst *fetcher) Fetch(ctx clients.ContainerContext, pack interface{}) error {
	result, err := runCommands(ctx, inst.cmdGrp)
	if err != nil {
		return err
	}
	if inst.postProcesser != nil {
		result, err = inst.postProcesser(result)
		if err != nil {
			return fmt.Errorf("feching failed post process the data %w", err)
		}
	}
	err = unmarshall(result, pack)
	if err != nil {
		return fmt.Errorf("feching failed to unpack data %w", err)
	}
	return nil
}

// runCommands executes the commands on the container passed as the ctx
// and extracts the results from the stdout
func runCommands(ctx clients.ContainerContext, cmdGrp clients.Cmder) (result map[string]string, err error) { //nolint:lll // allow slightly long function definition
	clientset := clients.GetClientset()
	cmd := cmdGrp.GetCommand()
	command := []string{"/usr/bin/sh"}
	var buffIn bytes.Buffer
	buffIn.WriteString(cmd)

	stdout, _, err := clientset.ExecCommandContainerStdIn(ctx, command, buffIn)
	if err != nil {
		log.Errorf("command in container failed unexpectedly. context: %v", ctx)
		log.Errorf("command in container failed unexpectedly. command: %v", command)
		log.Errorf("command in container failed unexpectedly. error: %v", err)
		return result, fmt.Errorf("runCommands failed %w", err)
	}
	result, err = cmdGrp.ExtractResult(stdout)
	if err != nil {
		log.Errorf("extraction failed %s", err.Error())
		log.Errorf("output was %s", stdout)
		return result, fmt.Errorf("runCommands failed %w", err)
	}
	return result, nil
}
