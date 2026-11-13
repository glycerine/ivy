//go:build !linux && !aix && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !solaris && !windows

package abs

// No suitable process CPU clock is available for this target through the
// standard library. Fail explicitly if a timed ABS run is requested.
func absCPUClock() int64        { panic("abs: CPU-time deadlines are unsupported on this platform") }
func absPinClockThread() func() { return func() {} }
