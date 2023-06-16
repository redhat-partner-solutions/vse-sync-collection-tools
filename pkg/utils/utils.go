// SPDX-License-Identifier: GPL-2.0-or-later

package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	Epoch = time.Unix(0, 0)
)

func IfErrorPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// The standard sync.WaitGroup doesn't expose the
// count of members as this is considered internal state
// however this value is very useful.
type WaitGroupCount struct {
	sync.WaitGroup
	count int64
}

func (wg *WaitGroupCount) Add(delta int) {
	atomic.AddInt64(&wg.count, int64(delta))
	wg.WaitGroup.Add(delta)
}

func (wg *WaitGroupCount) Done() {
	atomic.AddInt64(&wg.count, -1)
	wg.WaitGroup.Done()
}

func (wg *WaitGroupCount) GetCount() int {
	return int(atomic.LoadInt64(&wg.count))
}

// ParseTimestamp converts an input number of seconds (including a decimal fraction) into a time.Time
func ParseTimestamp(timestamp string) (time.Time, error) {
	duration, err := time.ParseDuration(fmt.Sprintf("%ss", timestamp))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp as a duration %w", err)
	}
	return Epoch.Add(duration).UTC(), nil
}
