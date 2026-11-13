//go:build aix || darwin || dragonfly || freebsd || netbsd || openbsd || solaris

package abs

import (
	"testing"
	"time"
)

func TestProcessCPUClockAdvancesWithWork(t *testing.T) {
	start := absCPUClock()
	// Bound wall time in case the OS cannot report a working CPU clock.
	deadline := time.Now().Add(3 * time.Second)
	for absCPUClock()-start < int64(5*time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("process CPU clock did not advance")
		}
	}
	if now := absCPUClock(); now < start {
		t.Fatalf("process CPU clock moved backward: %d to %d", start, now)
	}
}
