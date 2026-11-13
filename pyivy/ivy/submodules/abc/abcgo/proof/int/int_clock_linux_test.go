//go:build linux

package int

import (
	"testing"
	"time"
)

func TestInterpolationClockExcludesSleep(t *testing.T) {
	defer interPinClock()()
	start := interCPUClock()
	if start < 0 {
		t.Fatal("thread CPU clock unavailable")
	}
	time.Sleep(60 * time.Millisecond)
	elapsed := interCPUClock() - start
	if elapsed < 0 || elapsed >= int64(30*time.Millisecond) {
		t.Fatalf("sleep consumed %s of thread CPU", time.Duration(elapsed))
	}
}
