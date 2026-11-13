//go:build !linux && !cgo

package cec

import (
	"strings"
	"testing"
)

func TestSourceClockUnavailableWithoutCgo(t *testing.T) {
	now, release := cecSweepSourceClock()
	defer release()
	if now != nil {
		t.Fatal("unsupported sweep clock must be unavailable")
	}
	defer func() {
		r := recover()
		text, ok := r.(string)
		if !ok || !strings.Contains(text, "source C clock requires cgo") {
			t.Fatalf("unexpected unsupported-clock panic %v", r)
		}
	}()
	cecCPUClock()
}
