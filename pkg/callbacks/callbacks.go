// SPDX-License-Identifier: GPL-2.0-or-later

package callbacks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	logFilePermissions = 0666
)

type Callback interface {
	Call(OutputType, string) error
	CleanUp() error
	getFormat() OutputFormat
}

type OutputFormat int

const (
	Raw OutputFormat = iota
	// AnalyserJSON
)

type OutputType interface {
}

func getLine(c Callback, output OutputType, tag string) (string, error) {
	switch c.getFormat() {
	case Raw:
		line, err := json.Marshal(output)
		if err != nil {
			return "", fmt.Errorf("failed to mashall %T %w", output, err)
		}
		return fmt.Sprintf("%T:%s, %s", output, tag, line), nil
	default:
		return "", errors.New("unknown format")
	}
}

// SetupCallback if filename is empty or "-" it will output to stdout otherwise it will
// write to a file of the given name
func SetupCallback(filename string, format OutputFormat) (FileCallBack, error) {
	var (
		fileHandle io.WriteCloser
		err        error
	)

	if filename == "-" || filename == "" {
		fileHandle = os.Stdout
	} else {
		fileHandle, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, logFilePermissions)
		if err != nil {
			return FileCallBack{}, fmt.Errorf("failed to open file: %w", err)
		}
	}
	return FileCallBack{FileHandle: fileHandle, format: format}, nil
}

type FileCallBack struct {
	FileHandle io.WriteCloser
	format     OutputFormat
}

func (c FileCallBack) Call(output OutputType, tag string) error {
	line, err := getLine(c, output, tag)
	if err != nil {
		return err
	}
	_, err = c.FileHandle.Write([]byte(line))
	if err != nil {
		return fmt.Errorf("failed to write to file in callback: %w", err)
	}
	return nil
}

func (c FileCallBack) getFormat() OutputFormat {
	return c.format
}

func (c FileCallBack) CleanUp() error {
	err := c.FileHandle.Close()
	if err != nil {
		return fmt.Errorf("failed to close file handle in callback: %w", err)
	}
	return nil
}
