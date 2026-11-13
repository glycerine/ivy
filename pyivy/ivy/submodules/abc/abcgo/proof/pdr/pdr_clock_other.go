//go:build !linux && !aix && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !solaris && !windows

package pdr

func pdrCPUClock() int64  { panic("pdr: timed verification is unsupported on this platform") }
func pdrPinClock() func() { return func() {} }
