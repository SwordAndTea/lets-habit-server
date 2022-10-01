package util

import (
	"fmt"
	"time"
)

// Timer to calculate time
type Timer struct {
	startTime time.Time
}

// NewTimer create new timer
func NewTimer() *Timer {
	return &Timer{startTime: time.Now()}
}

// Reset set start of timer to now
func (timer *Timer) Reset() {
	timer.startTime = time.Now()
}

// Elapsed get the elapsed since start in string
func (timer *Timer) Elapsed() string {
	return fmt.Sprintf("%s", time.Since(timer.startTime))
}

// ElapsedUInt64 get the elapsed since start in int64
func (timer *Timer) ElapsedUInt64() uint64 {
	return uint64(time.Since(timer.startTime) / time.Millisecond)
}

// Duration get the elapsed since start in duration
func (timer *Timer) Duration() time.Duration {
	return time.Since(timer.startTime)
}
