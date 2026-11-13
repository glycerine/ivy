//go:build windows

package ssw

import "time"

// Microsoft CRT clock() measures elapsed milliseconds since process start.
var sswClockEpoch = time.Now()

func sswCPUClock() int64  { return time.Since(sswClockEpoch).Milliseconds() }
func sswPinClock() func() { return func() {} }
