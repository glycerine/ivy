//go:build linux

package ssw

import (
	"testing"
	"time"
)

func TestSignalFilterClockExcludesSleep(t *testing.T) {
	defer sswPinClock()()
	start := sswCPUClock()
	if start < 0 {
		t.Fatal("thread clock unavailable")
	}
	time.Sleep(60 * time.Millisecond)
	elapsed := sswCPUClock() - start
	if elapsed < 0 || elapsed >= 30000 {
		t.Fatalf("sleep consumed %d microseconds of thread CPU", elapsed)
	}
}
