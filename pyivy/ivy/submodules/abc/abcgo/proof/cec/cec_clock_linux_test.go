//go:build linux

package cec

import (
	"testing"
	"time"
)

func TestCecThreadClockDoesNotChargeSleep(t *testing.T) {
	unpin := cecPinClockThread()
	defer unpin()
	start := cecCPUClock()
	time.Sleep(40 * time.Millisecond)
	if elapsed := cecCPUClock() - start; elapsed < 0 || elapsed >= int64(30*time.Millisecond) {
		t.Fatalf("thread CPU clock advanced %s during sleep", time.Duration(elapsed))
	}
}
