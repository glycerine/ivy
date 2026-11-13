//go:build windows

package fra

import "time"

// Microsoft CRT clock() measures elapsed milliseconds since process start.
var fraLcrEpoch = time.Now()

func fraLcrClock() int64 { return time.Since(fraLcrEpoch).Milliseconds() }
func fraLcrPin() func()  { return func() {} }
