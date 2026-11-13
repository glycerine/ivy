//go:build !linux

package cec

import "time"

// Other platforms retain the existing elapsed-time policy pending a
// platform-specific source clock implementation.
func cecCPUClock() int64 { return time.Now().UnixNano() }
func cecPinClockThread() func() { return func() {} }
