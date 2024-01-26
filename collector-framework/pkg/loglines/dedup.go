// SPDX-License-Identifier: GPL-2.0-or-later

package loglines

import (
	"fmt"
	"os"
	"sort"

	log "github.com/sirupsen/logrus"
)

const (
	keepGenerations = 5
)

func dedupGeneration(lineSlices []*LineSlice) *LineSlice {
	ls1, ls2 := DedupLineSlices(lineSlices)
	output := MakeSliceFromLines(MakeNewCombinedSlice(ls1.Lines, ls2.Lines), ls2.Generation)
	return output
}

// findLineIndex will find the index of a line in a slice of lines
// and will return -1 if it is not found
func findLineIndex(needle *ProcessedLine, list []*ProcessedLine) int {
	checkLine := needle.Full
	for i, hay := range list {
		if hay.Full == checkLine {
			return i
		}
	}
	return -1
}

// looks for a line that exists in x without being present at the same index in y.
//
//nolint:varnamelen // x and y are just two sets of lines
func findFirstIssueIndex(x, y []*ProcessedLine) int {
	for index, line := range x {
		if index >= len(y) {
			// we've reached the end of y so no bad lines found
			break
		}
		if line.Full != y[index].Full {
			return index
		}
	}
	return -1
}

func findNextMatching(a, b []*ProcessedLine) (offset, index int) {
	index = findLineIndex(a[offset], b)
	if index == -1 {
		// Can't find X in Y, Search for next line which is in Y
		for offset = 1; offset < len(a); offset++ {
			index = findLineIndex(a[offset], b)
			if index != -1 {
				break
			}
		}
	}
	return offset, index
}

// fixLines will insert sequental lines missing in either x or y when stich them into the other.
// This is only done for the first issue found so you will need to call this iteratively.
// It will return the fixed slices and a flag which tells you if anything changed.
// If the lines at issueIndex are the same it will return the input slices.
//
//nolint:gocritic,varnamelen // don't want to name the return values as they should be built later, I think x, y are expressive enough names
func fixLines(x, y []*ProcessedLine, issueIndex int) ([]*ProcessedLine, []*ProcessedLine, bool) {
	// confirm x[i] != y[i]
	if x[issueIndex].Full == y[issueIndex].Full {
		// No issue here...
		return x, y, false
	}

	// Check if its the y value missing in x
	if findLineIndex(y[issueIndex], x[issueIndex:]) == -1 {
		// lets add missing y values into x
		yOffset, xIndex := findNextMatching(y[issueIndex:], x[issueIndex:])
		newX := make([]*ProcessedLine, 0, len(x)+yOffset)
		newX = append(newX, x[:issueIndex]...)
		newX = append(newX, y[issueIndex:issueIndex+yOffset]...)
		newX = append(newX, x[issueIndex+xIndex:]...)
		return newX, y, true
	}

	if findLineIndex(x[issueIndex], y[issueIndex:]) == -1 {
		xOffset, yIndex := findNextMatching(x[issueIndex:], y[issueIndex:])
		newY := make([]*ProcessedLine, 0, len(y)+xOffset)
		newY = append(newY, y[:issueIndex]...)
		newY = append(newY, x[issueIndex:issueIndex+xOffset]...)
		newY = append(newY, y[issueIndex+yIndex:]...)
		return x, newY, true
	}
	return x, y, false
}

// processOverlap checks for incompatible lines if they are found it attempts to fix the issue
// and then if it fixes it then it dedups the fixed sets
//
//nolint:gocritic,varnamelen // don't want to name the return values as they should be built later, I think x, y are expressive enough names
func processOverlap(x, y []*ProcessedLine) ([]*ProcessedLine, []*ProcessedLine, error) {
	newX := x
	newY := y
	issueIndex := findFirstIssueIndex(newX, newY)
	if issueIndex == -1 {
		return newX, newY, nil
	}
	var changed bool
	for issueIndex != -1 {
		newX, newY, changed = fixLines(newX, newY, issueIndex)
		issueIndex = findFirstIssueIndex(newX, newY)
		if !changed && issueIndex != -1 {
			return newX, newY, fmt.Errorf("failed to resolve overlap")
		}
	}
	// Now its been fixed lets just dedup it completely again.
	dx, dy := DedupAB(newX, newY)
	return dx, dy, nil
}

//nolint:gocritic,varnamelen // don't want to name the return values as they should be built later, I think a, b are expressive enough names
func handleIncompleteOverlap(a, b []*ProcessedLine) ([]*ProcessedLine, []*ProcessedLine) {
	newA, newB, err := processOverlap(a, b)
	if err != nil {
		issueIndex := findFirstIssueIndex(newA, newB)
		log.Warning("Failed to fix issues gonna just split at the issue and retry this might lose some data")
		return DedupAB(newA[:issueIndex], newB[issueIndex:])
	}
	return newA, newB
}

//nolint:gocritic,varnamelen // don't want to name the return values as they should be built later, I think a, b are expressive enough names
func DedupAB(a, b []*ProcessedLine) ([]*ProcessedLine, []*ProcessedLine) {
	bFirstLineIndex := findLineIndex(b[0], a)
	log.Debug("line index: ", bFirstLineIndex)
	if bFirstLineIndex == -1 {
		log.Debug("didn't to find first line of b")
		lastLineIndex := findLineIndex(a[len(a)-1], b)
		if lastLineIndex == -1 {
			log.Debug("didn't to find last line of a; assuming no overlap")
			return a, b
		}

		// we have the index of the last line of a in b
		// so b[:lastLineIndex+1] contains some of a
		// lets try
		return handleIncompleteOverlap(a, b)
	}

	index := findFirstIssueIndex(a[bFirstLineIndex:], b)
	if index >= 0 {
		return handleIncompleteOverlap(a, b)
	}
	return a[:bFirstLineIndex], b
}

func MakeNewCombinedSlice(x, y []*ProcessedLine) []*ProcessedLine {
	r := make([]*ProcessedLine, 0, len(x)+len(y))
	r = append(r, x...)
	r = append(r, y...)
	return r
}

//nolint:gocritic // don't want to name the return values as they should be built later
func DedupLineSlices(lineSlices []*LineSlice) (*LineSlice, *LineSlice) {
	sort.Slice(lineSlices, func(i, j int) bool {
		sdif := lineSlices[i].start.Sub(lineSlices[j].start)
		if sdif == 0 {
			// If start is the same then lets put the earlist end time first
			return lineSlices[i].end.Sub(lineSlices[j].end) < 0
		}
		return sdif < 0
	})

	if len(lineSlices) == 1 {
		return &LineSlice{}, lineSlices[0]
	}

	lastLineSlice := lineSlices[len(lineSlices)-1]
	lastButOneLineSlice := lineSlices[len(lineSlices)-2]

	// work backwards thought the slices
	// dedupling the earlier one along the way
	dedupedLines, lastLines := DedupAB(lastButOneLineSlice.Lines, lastLineSlice.Lines)

	if len(lineSlices) == 2 { //nolint:gomnd // it would obscure that its just a length of 2
		if len(dedupedLines) == 0 {
			return &LineSlice{Generation: lastButOneLineSlice.Generation},
				MakeSliceFromLines(lastLines, lastLineSlice.Generation)
		}
		return MakeSliceFromLines(dedupedLines, lastButOneLineSlice.Generation),
			MakeSliceFromLines(lastLines, lastLineSlice.Generation)
	}

	resLines := dedupedLines
	reference := MakeNewCombinedSlice(dedupedLines, lastLines)
	for index := len(lineSlices) - 3; index >= 0; index-- {
		aLines, bLines := DedupAB(lineSlices[index].Lines, reference)
		resLines = MakeNewCombinedSlice(aLines, resLines)
		reference = MakeNewCombinedSlice(aLines, bLines)
	}
	return MakeSliceFromLines(resLines, lastButOneLineSlice.Generation),
		MakeSliceFromLines(lastLines, lastLineSlice.Generation)
}

func WriteOverlap(lines []*ProcessedLine, name string) error {
	logFile, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed %w", err)
	}
	defer logFile.Close()

	for _, line := range lines {
		_, err := logFile.WriteString(line.Full + "\n")
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}
