//go:build windows

package abs

import "time"

// Microsoft CRT clock() measures elapsed time since process startup, not CPU
// time. Differences of samples from this monotonic clock have that behavior.
// See https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/clock
var absClockEpoch = time.Now()

func absCPUClock() int64        { return time.Since(absClockEpoch).Nanoseconds() }
func absPinClockThread() func() { return func() {} }
