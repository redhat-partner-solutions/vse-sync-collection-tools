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
	AnalyserJSON
)

type AnalyserFormatType struct {
	Data any    `json:"data"`
	ID   string `json:"id"`
}

type OutputType interface {
	GetAnalyserFormat() ([]*AnalyserFormatType, error)
}

// getFormattedOutput returns the output in the format configured by on the callback
func getFormattedOutput(c Callback, output OutputType, tag string) ([]byte, error) {
	switch c.getFormat() {
	case Raw:
		line, err := json.Marshal(output)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to marshal %T %w", output, err)
		}
		return []byte(fmt.Sprintf("%T:%s, %s", output, tag, line)), nil
	case AnalyserJSON:
		outputs, err := output.GetAnalyserFormat()
		if err != nil {
			return []byte{}, fmt.Errorf("failed to get AnalyserFormat %w", err)
		}
		newline := []byte("\n")
		lines := make([]byte, 0)
		for count, obj := range outputs {
			line, err := json.Marshal(obj)
			if err != nil {
				return []byte{}, fmt.Errorf("failed to marshal AnalyserFormat for %s %w", tag, err)
			}
			lines = append(lines, line...)
			// Append a new line between the entries but not a trailing one
			if count < len(outputs)-1 {
				lines = append(lines, newline...)
			}
		}
		return lines, nil
	default:
		return []byte{}, errors.New("unknown format")
	}
}

// Returns the filehandle for callback
// if filename is empty or "-" it will output to stdout otherwise it will
// write to a file of the given name
func GetFileHandle(filename string) (io.WriteCloser, error) {
	var (
		fileHandle io.WriteCloser
		err        error
	)
	if filename == "-" || filename == "" {
		fileHandle = os.Stdout
	} else {
		fileHandle, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, logFilePermissions)
		if err != nil {
			return fileHandle, fmt.Errorf("failed to open file: %w", err)
		}
	}
	return fileHandle, nil
}

func NewFileCallback(fileHandle io.WriteCloser, format OutputFormat) FileCallBack {
	return FileCallBack{fileHandle: fileHandle, format: format}
}

// SetupCallback returns a FileCallback
// if filename is empty or "-" it will output to stdout otherwise it will
// write to a file of the given name
func SetupCallback(filename string, format OutputFormat) (FileCallBack, error) {
	fileHandle, err := GetFileHandle(filename)
	if err != nil {
		return FileCallBack{}, err
	}
	return NewFileCallback(fileHandle, format), nil
}

type FileCallBack struct {
	fileHandle io.WriteCloser
	format     OutputFormat
}

func (c FileCallBack) Call(output OutputType, tag string) error {
	formattedOutput, err := getFormattedOutput(c, output, tag)
	if err != nil {
		return err
	}
	formattedOutput = append(formattedOutput, []byte("\n")...)
	_, err = c.fileHandle.Write(formattedOutput)
	if err != nil {
		return fmt.Errorf("failed to write to file in callback: %w", err)
	}
	return nil
}

func (c FileCallBack) getFormat() OutputFormat {
	return c.format
}

func (c FileCallBack) CleanUp() error {
	err := c.fileHandle.Close()
	if err != nil {
		return fmt.Errorf("failed to close file handle in callback: %w", err)
	}
	return nil
}
