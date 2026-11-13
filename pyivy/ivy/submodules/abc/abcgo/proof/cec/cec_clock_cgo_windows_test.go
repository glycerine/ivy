//go:build windows && cgo

package cec

import (
	"testing"
	"time"
)

func TestSourceCRTClockChargesElapsedSleep(t *testing.T) {
	before := cecCPUClock()
	time.Sleep(50 * time.Millisecond)
	elapsed := cecCPUClock() - before
	if elapsed < int64(30*time.Millisecond) {
		t.Fatalf("CRT elapsed clock advanced only %s during sleep", time.Duration(elapsed))
	}
}
