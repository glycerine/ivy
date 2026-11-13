//go:build !linux && !aix && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !solaris && !windows

package abs

import "testing"

func TestUnsupportedCPUClockIsExplicit(t *testing.T) {
	defer func() {
		if got := recover(); got != "abs: CPU-time deadlines are unsupported on this platform" {
			t.Fatalf("unexpected panic: %v", got)
		}
	}()
	absCPUClock()
}
