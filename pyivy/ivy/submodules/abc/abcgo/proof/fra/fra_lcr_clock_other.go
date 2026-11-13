//go:build !linux && !aix && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !solaris && !windows

package fra

func fraLcrClock() int64 { panic("fra: timed latch correspondence is unsupported on this platform") }
func fraLcrPin() func()  { return func() {} }
