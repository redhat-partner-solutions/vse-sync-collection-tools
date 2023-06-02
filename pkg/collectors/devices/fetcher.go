// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

type fetcher struct {
	cmdGrp *clients.CmdGroup
	Result map[string]string
}

func NewFetcher() *fetcher {
	return &fetcher{
		cmdGrp: &clients.CmdGroup{},
	}
}

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

func (inst *fetcher) AddCommand(cmdInst *clients.Cmd) {
	inst.cmdGrp.AddCommand(cmdInst)
}

func (inst *fetcher) Unmarshall(pack interface{}) error {
	if inst.Result == nil {
		return errors.New("result is not populated")
	}
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
				if res, ok := inst.Result[resultName]; ok {
					f.SetString(res)
				}
			default:
				panic(fmt.Sprintf("fetcher unmarshal not implmented for type: %s", field.Type.Name()))
			}
		}
	}
	return nil
}

func (inst *fetcher) Fetch(ctx clients.ContainerContext, pack interface{}) error {
	inst.Result = runCommands(ctx, inst.cmdGrp)
	err := inst.Unmarshall(pack)
	if err != nil {
		return fmt.Errorf("feching failed to unpack data %w", err)
	}
	return nil
}

func runCommands(ctx clients.ContainerContext, cmdGrp clients.Cmder) (result map[string]string) { //nolint:lll // allow slightly long function definition
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
		return result
	}
	result, err = cmdGrp.ExtractResult(stdout)
	if err != nil {
		log.Errorf("extraction failed %s", err.Error())
		log.Errorf("output was %s", stdout)
	}
	return
}
