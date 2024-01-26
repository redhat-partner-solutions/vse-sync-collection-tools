// SPDX-License-Identifier: GPL-2.0-or-later

package clients

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
)

type Cmder interface {
	GetCommand() string
	ExtractResult(string) (map[string]string, error)
}

type Cmd struct {
	key             string
	prefix          string
	suffix          string
	cmd             string
	outputProcessor func(string) (string, error)
	regex           *regexp.Regexp
	fullCmd         string
}

func NewCmd(key, cmd string) (*Cmd, error) {
	cmdInstance := Cmd{
		key:    key,
		cmd:    cmd,
		prefix: fmt.Sprintf("echo '<%s>'", key),
		suffix: fmt.Sprintf("echo '</%s>'", key),
	}

	cmdInstance.fullCmd = fmt.Sprintf("%s;", cmdInstance.prefix)
	cmdInstance.fullCmd += cmdInstance.cmd
	if string(cmd[len(cmd)-1]) != ";" {
		cmdInstance.fullCmd += ";"
	}
	cmdInstance.fullCmd += fmt.Sprintf("%s;", cmdInstance.suffix)

	compiledRegex, err := regexp.Compile(`(?s)<` + key + `>\n` + `(.*)` + `\n</` + key + `>`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex for key %s: %w", key, err)
	}
	cmdInstance.regex = compiledRegex
	return &cmdInstance, nil
}

func (c *Cmd) SetOutputProcessor(f func(string) (string, error)) {
	c.outputProcessor = f
}

func (c *Cmd) GetCommand() string {
	return c.fullCmd
}

func (c *Cmd) ExtractResult(s string) (map[string]string, error) {
	result := make(map[string]string)
	log.Debugf("extract %s from %s", c.key, s)
	match := c.regex.FindStringSubmatch(s)
	log.Debugf("match %#v", match)

	if len(match) > 0 {
		value := match[1]

		if c.outputProcessor != nil {
			cleanValue, err := c.outputProcessor(match[1])
			if err != nil {
				return result, fmt.Errorf("failed to cleanup value %s of key %s", match[1], c.key)
			}
			value = cleanValue
		}
		log.Debugf("r %s", value)
		result[c.key] = value
		return result, nil
	}
	return result, fmt.Errorf("failed to find result for key: %s", c.key)
}

type CmdGroup struct {
	cmds []*Cmd
}

func (cgrp *CmdGroup) AddCommand(c *Cmd) {
	cgrp.cmds = append(cgrp.cmds, c)
}

func (cgrp *CmdGroup) GetCommand() string {
	res := ""
	for _, c := range cgrp.cmds {
		res += c.GetCommand()
	}
	return res
}

func (cgrp *CmdGroup) ExtractResult(s string) (map[string]string, error) {
	results := make(map[string]string)
	for _, c := range cgrp.cmds {
		res, err := c.ExtractResult(s)
		if err != nil {
			return results, err
		}
		results[c.key] = res[c.key]
	}
	return results, nil
}
