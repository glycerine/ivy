//go:build !linux && !cgo

package cec

import "time"

// The source C clock is unavailable without cgo on non-Linux targets.
func cecSweepSourceClock() (func() time.Time, func()) { return nil, func() {} }
