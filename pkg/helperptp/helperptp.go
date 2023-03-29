// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helperptp

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	writeFileCap   = 0666
	ptpMETRICS     = 4
	ptpGROUPLENGTH = 4

	ptpGROUP1 = "ptp4l"
	ptpGROUP3 = " master"
)

var (
	ptpFileRegExp = regexp.MustCompile(`(\w+)\[(\d+\.\d{3})\]:\s\[\S+\](.*)`)
	ptpLineRegExp = regexp.MustCompile(`-?\d+`)
)

func Readln(reader *bufio.Reader) (string, error) {
	var (
		isPrefix               = true
		err              error = nil
		line, lineOutput []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = reader.ReadLine()
		lineOutput = append(lineOutput, line...)
	}
	return string(lineOutput), err
}

func decodePtp4line(tstamp, syncMetrics string, processedFile *os.File) error {
	ptp4lSampleData := ptpLineRegExp.FindAllString(syncMetrics, -1)
	if len(ptp4lSampleData) == ptpMETRICS {
		// write timestamp, offset with master, state, frequency, and path delay
		ptp4lSampleData = append([]string{tstamp}, ptp4lSampleData...)
		join := strings.Join(ptp4lSampleData, ",")
		_, err := processedFile.WriteString(join + "\n")
		if err != nil {
			return fmt.Errorf("failure updating ptp processed file: %w", err)
		}
	}
	return nil
}

func ParsePtpLogs(filename, filenameDst string) error {
	ptpFile, err := os.Open(filename)
	if err != nil {
		log.Panic("error problem reading file")
	}
	ptpReader := bufio.NewReader(ptpFile)
	processedFile, err := os.OpenFile(filenameDst, os.O_CREATE|os.O_WRONLY, writeFileCap)
	if err != nil {
		return fmt.Errorf("failure opening ptp data file: %w", err)
	}
	_, err = processedFile.WriteString("tstamp,phase,state,freq,delay,event" + "\n")
	if err != nil {
		return fmt.Errorf("failure writing ptp data file: %w", err)
	}
	ptpSample, err := Readln(ptpReader)
	defer processedFile.Close()
	for err == nil {
		sampleGroups := ptpFileRegExp.FindStringSubmatch(ptpSample)
		if len(sampleGroups) == ptpGROUPLENGTH {
			// only support ptp4l samples
			if sampleGroups[1] == ptpGROUP1 && (strings.HasPrefix(sampleGroups[3], ptpGROUP3)) {
				err = decodePtp4line(sampleGroups[2], sampleGroups[3], processedFile)
				if err != nil {
					return fmt.Errorf("failure decoding ptp line: %w", err)
				}
			}
		}
		ptpSample, err = Readln(ptpReader)
	}
	return nil
}
