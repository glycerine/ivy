//go:build !linux

package cec

import "time"

// A source-compatible CPU clock is not yet implemented on this platform.
func cecSweepSourceClock() (func() time.Time, func()) { return nil, func() {} }
