//go:build !linux && !aix && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !solaris && !windows

package int

func interCPUClock() int64  { panic("int: timed interpolation is unsupported on this platform") }
func interPinClock() func() { return func() {} }
