//go:build linux

package fra

import (
	"testing"
	"time"
)

func TestLatchCorrespondenceClockExcludesSleep(t *testing.T) {
	defer fraLcrPin()()
	start := fraLcrClock()
	if start < 0 {
		t.Fatal("thread clock unavailable")
	}
	time.Sleep(60 * time.Millisecond)
	elapsed := fraLcrClock() - start
	if elapsed < 0 || elapsed >= 30000 {
		t.Fatalf("sleep consumed %d microseconds of thread CPU", elapsed)
	}
	// A deadline earlier than the current clock must be observed strictly.
	if now := fraLcrClock(); !(now > now-1) {
		t.Fatal("expired deadline not ordered")
	}
}
