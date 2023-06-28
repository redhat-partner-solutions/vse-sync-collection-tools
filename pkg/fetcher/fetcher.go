// SPDX-License-Identifier: GPL-2.0-or-later

package fetcher

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type Fetcher struct {
	cmdGrp        *clients.CmdGroup
	postProcesser func(map[string]string) (map[string]any, error)
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		cmdGrp: &clients.CmdGroup{},
	}
}

func TrimSpace(s string) (string, error) {
	return strings.TrimSpace(s), nil
}

func (inst *Fetcher) SetPostProcesser(ppFunc func(map[string]string) (map[string]any, error)) {
	inst.postProcesser = ppFunc
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

func setValue(field *reflect.StructField, fieldVal reflect.Value, value any) error {
	//nolint:exhaustive //we could extend this but its not needed yet
	switch field.Type.Kind() {
	case reflect.String:
		stringRes, ok := value.(string)
		if !ok {
			return fmt.Errorf("failed to convert %v into string", value)
		}
		fieldVal.SetString(stringRes)
	case reflect.Slice:
		resType := reflect.TypeOf(value)
		if field.Type != resType {
			return fmt.Errorf(
				"type of %v does not match field type %v",
				resType,
				field.Type,
			)
		}
		fieldVal.Set(reflect.ValueOf(value))
	default:
		return fmt.Errorf("fetcher unmarshal not implemented for type: %s", field.Type.Name())
	}
	return nil
}

// unmarshal will populate the fields in `target` with the values from `result` according to the fields`fetcherKey` tag.
// fields with no `fetcherKey` tag will not be touched, and elements in `result` without a matched field will be ignored.
func unmarshal(result map[string]any, target any) error {
	val := reflect.ValueOf(target)
	typ := reflect.TypeOf(target)

	for i := 0; i < val.Elem().NumField(); i++ {
		field := typ.Elem().Field(i)
		resultName := field.Tag.Get("fetcherKey")
		if resultName != "" {
			f := val.Elem().FieldByIndex(field.Index)
			if res, ok := result[resultName]; ok {
				err := setValue(&field, f, res)
				if err != nil {
					return fmt.Errorf("failed to set value on feild %s: %w", resultName, err)
				}
			}
		}
	}
	return nil
}

// Fetch executes the commands on the container passed as the ctx and
// use the results to populate pack
func (inst *Fetcher) Fetch(ctx clients.ContainerContext, pack any) error {
	runResult, err := runCommands(ctx, inst.cmdGrp)
	if err != nil {
		return err
	}
	result := make(map[string]any)
	for key, value := range runResult {
		result[key] = value
	}
	if inst.postProcesser != nil {
		updatedResults, ppErr := inst.postProcesser(runResult)
		if ppErr != nil {
			return fmt.Errorf("feching failed post process the data %w", err)
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
func runCommands(ctx clients.ContainerContext, cmdGrp clients.Cmder) (result map[string]string, err error) { //nolint:lll // allow slightly long function definition
	clientset, err := clients.GetClientset()
	if err != nil {
		return result, fmt.Errorf("failed to get clientset %w", err)
	}
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
