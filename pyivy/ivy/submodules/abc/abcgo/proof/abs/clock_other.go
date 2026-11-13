//go:build !linux

package abs

import "time"

// Fidelity gap: ABC uses process CPU time on these platforms. Until a
// platform-specific CPU-clock implementation is available, use wall time.
func absCPUClock() int64        { return time.Now().UnixNano() }
func absPinClockThread() func() { return func() {} }
