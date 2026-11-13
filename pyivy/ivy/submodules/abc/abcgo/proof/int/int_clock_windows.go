//go:build windows

package int

import "time"

// Microsoft CRT clock() measures elapsed milliseconds since process start.
var interClockEpoch = time.Now()

func interCPUClock() int64  { return time.Since(interClockEpoch).Milliseconds() * 1000000 }
func interPinClock() func() { return func() {} }
