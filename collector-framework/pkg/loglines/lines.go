// SPDX-License-Identifier: GPL-2.0-or-later

package loglines

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/collector-framework/pkg/utils"
)

const (
	dumpChannelSize = 100
)

type ProcessedLine struct {
	Timestamp  time.Time
	Full       string
	Content    string
	Generation uint32
}

func ProcessLine(line string) (*ProcessedLine, error) {
	splitLine := strings.SplitN(line, " ", 2) //nolint:gomnd // moving this to a var would make the code less clear
	if len(splitLine) < 2 {                   //nolint:gomnd // moving this to a var would make the code less clear
		return nil, fmt.Errorf("failed to split line %s", line)
	}
	timestampPart := splitLine[0]
	lineContent := splitLine[1]
	timestamp, err := time.Parse(time.RFC3339, timestampPart)
	if err != nil {
		// This is not a value line something went wrong
		return nil, fmt.Errorf("failed to process timestamp from line: '%s'", line)
	}
	processed := &ProcessedLine{
		Timestamp: timestamp,
		Content:   strings.TrimRightFunc(lineContent, unicode.IsSpace),
		Full:      strings.TrimRightFunc(line, unicode.IsSpace),
	}
	return processed, nil
}

func NewGenerationalLockedTime(initialTime time.Time) GenerationalLockedTime {
	return GenerationalLockedTime{time: initialTime}
}

type GenerationalLockedTime struct {
	time       time.Time
	lock       sync.RWMutex
	generation uint32
}

func (lt *GenerationalLockedTime) Time() time.Time {
	lt.lock.RLock()
	defer lt.lock.RUnlock()
	return lt.time
}

func (lt *GenerationalLockedTime) Generation() uint32 {
	lt.lock.RLock()
	defer lt.lock.RUnlock()
	return lt.generation
}

func (lt *GenerationalLockedTime) Update(update time.Time) {
	lt.lock.Lock()
	defer lt.lock.Unlock()
	if update.Sub(lt.time) > 0 {
		lt.time = update
		lt.generation += 1
	}
}

type LineSlice struct {
	start      time.Time
	end        time.Time
	Lines      []*ProcessedLine
	Generation uint32
}

type Generations struct {
	Store  map[uint32][]*LineSlice
	Dumper *GenerationDumper
	Latest uint32
	Oldest uint32
}

type Dump struct {
	slice       *LineSlice
	numberInGen int
}

func NewGenerationDumper(dir string, keepLogs bool) *GenerationDumper {
	return &GenerationDumper{
		dir:       dir,
		toDump:    make(chan *Dump, dumpChannelSize),
		quit:      make(chan *os.Signal),
		filenames: make([]string, 0),
		keepLogs:  keepLogs,
	}
}

type GenerationDumper struct {
	dir       string
	toDump    chan *Dump
	quit      chan *os.Signal
	filenames []string
	keepLogs  bool
	wg        sync.WaitGroup
}

func (dump *GenerationDumper) Start() {
	dump.wg.Add(1)
	go dump.dumpProcessor()
}

func (dump *GenerationDumper) DumpLines(ls *LineSlice, numberInGen int) {
	dump.toDump <- &Dump{slice: ls, numberInGen: numberInGen}
}

func (dump *GenerationDumper) writeToFile(toDump *Dump) {
	fname := filepath.Join(
		dump.dir,
		fmt.Sprintf("generation-%d-%d.log", toDump.slice.Generation, toDump.numberInGen),
	)
	dump.filenames = append(dump.filenames, fname)
	err := WriteOverlap(toDump.slice.Lines, fname)
	if err != nil {
		log.Errorf("failed to write generation dump file: %s", err.Error())
	}
}

func (dump *GenerationDumper) dumpProcessor() {
	defer dump.wg.Done()
	for {
		select {
		case <-dump.quit:
			log.Info("Dumping slices")
			for len(dump.toDump) > 0 {
				toDump := <-dump.toDump
				dump.writeToFile(toDump)
			}
			return
		case toDump := <-dump.toDump:
			dump.writeToFile(toDump)
		default:
			time.Sleep(time.Nanosecond)
		}
	}
}

func (dump *GenerationDumper) Stop() {
	dump.quit <- &os.Kill
	log.Debug("waiting for generation dumping to complete")
	dump.wg.Wait()

	if !dump.keepLogs {
		log.Debug("removing generation dump files")
		utils.RemoveTempFiles(dump.dir, dump.filenames)
	}
}

func (gens *Generations) Add(lineSlice *LineSlice) {
	genSlice, ok := gens.Store[lineSlice.Generation]
	if !ok {
		genSlice = make([]*LineSlice, 0)
	}
	numberInGen := len(genSlice)
	gens.Store[lineSlice.Generation] = append(genSlice, lineSlice)
	gens.Dumper.DumpLines(lineSlice, numberInGen)

	log.Debug("Logs: all generations: ", gens.Store)

	if gens.Latest < lineSlice.Generation {
		gens.Latest = lineSlice.Generation
		log.Debug("Logs: lastest updated ", gens.Latest)
		log.Debug("Logs: should flush ", gens.ShouldFlush())
	}
}

func (gens *Generations) removeOlderThan(keepGen uint32) {
	log.Debug("Removing geners <", keepGen)
	for g := range gens.Store {
		if g < keepGen {
			delete(gens.Store, g)
		}
	}
	gens.Oldest = keepGen
}

func (gens *Generations) ShouldFlush() bool {
	return (gens.Latest-gens.Oldest > keepGenerations &&
		len(gens.Store) > keepGenerations)
}

func (gens *Generations) Flush() *LineSlice {
	lastGen := gens.Oldest + keepGenerations
	log.Debug("Flushing generations <=", lastGen)

	gensToFlush := make([][]*LineSlice, 0)
	for index, value := range gens.Store {
		if index <= lastGen {
			gensToFlush = append(gensToFlush, value)
		}
	}
	result, lastSlice := gens.flush(gensToFlush)
	gens.removeOlderThan(lastSlice.Generation)
	gens.Store[lastSlice.Generation] = []*LineSlice{lastSlice}
	return result
}

func (gens *Generations) FlushAll() *LineSlice {
	log.Debug("Flushing all generations")
	gensToFlush := make([][]*LineSlice, 0)
	for _, value := range gens.Store {
		gensToFlush = append(gensToFlush, value)
	}
	if len(gensToFlush) == 0 {
		return &LineSlice{}
	}
	result, lastSlice := gens.flush(gensToFlush)
	return MakeSliceFromLines(MakeNewCombinedSlice(result.Lines, lastSlice.Lines), lastSlice.Generation)
}

//nolint:gocritic // don't want to name the return values as they should be built later
func (gens *Generations) flush(generations [][]*LineSlice) (*LineSlice, *LineSlice) {
	log.Debug("genrations: ", generations)
	sort.Slice(generations, func(i, j int) bool {
		return generations[i][0].Generation < generations[j][0].Generation
	})
	dedupGen := make([]*LineSlice, len(generations))
	for index, gen := range generations {
		dedupGen[index] = dedupGeneration(gen)
	}
	return DedupLineSlices(dedupGen)
}

func MakeSliceFromLines(lines []*ProcessedLine, generation uint32) *LineSlice {
	if len(lines) == 0 {
		return &LineSlice{Generation: generation}
	}
	return &LineSlice{
		Lines:      lines,
		start:      lines[0].Timestamp,
		end:        lines[len(lines)-1].Timestamp,
		Generation: generation,
	}
}
