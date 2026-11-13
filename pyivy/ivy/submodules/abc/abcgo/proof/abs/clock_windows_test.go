//go:build windows

package abs

import (
	"testing"
	"time"
)

func TestWindowsClockCountsElapsedSleep(t *testing.T) {
	start := absCPUClock()
	time.Sleep(20 * time.Millisecond)
	if elapsed := absCPUClock() - start; elapsed < int64(10*time.Millisecond) {
		t.Fatalf("elapsed clock advanced only %v during sleep", time.Duration(elapsed))
	}
}
