//go:build !linux && cgo

package cec

/*
#include <time.h>
static long long cec_clock_ticks(void) { return (long long)clock(); }
static long long cec_clock_ticks_per_second(void) { return (long long)CLOCKS_PER_SEC; }
*/
import "C"

// On non-Linux ABC uses the C runtime's clock(). POSIX runtimes report
// process CPU time; Windows CRTs report elapsed time from CRT initialization.
// Values are converted to nanoseconds for the Go timing and deadline code.
// Exact compatibility requires the same C runtime as the ABC build.
func cecCPUClock() int64 {
	ticks := int64(C.cec_clock_ticks())
	frequency := int64(C.cec_clock_ticks_per_second())
	return (ticks/frequency)*1e9 + (ticks%frequency)*1e9/frequency
}
func cecPinClockThread() func() { return func() {} }
