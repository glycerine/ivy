//go:build !linux && !aix && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !solaris && !windows

package ssw

func sswCPUClock() int64  { panic("ssw: timed verification is unsupported on this platform") }
func sswPinClock() func() { return func() {} }
