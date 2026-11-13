//go:build !linux && cgo

package cec

import "time"

// Unlike Linux's thread clock, the C clock on these targets is process-wide
// (or elapsed CRT time on Windows), so it does not require thread pinning.
func cecSweepSourceClock() (func() time.Time, func()) {
	return func() time.Time { return time.Unix(0, cecCPUClock()) }, func() {}
}
