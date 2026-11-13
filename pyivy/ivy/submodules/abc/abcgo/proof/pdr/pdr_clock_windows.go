//go:build windows

package pdr

import "time"

// Microsoft CRT clock() measures elapsed milliseconds since process start.
var pdrClockEpoch = time.Now()

func pdrCPUClock() int64  { return time.Since(pdrClockEpoch).Milliseconds() * 1000000 }
func pdrPinClock() func() { return func() {} }
