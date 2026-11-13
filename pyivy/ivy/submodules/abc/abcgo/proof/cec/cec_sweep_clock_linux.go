//go:build linux

package cec

import (
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// ABC's Linux Abc_Clock uses CLOCK_THREAD_CPUTIME_ID. Keep the sweep on the
// same OS thread so its start and end readings refer to the same CPU clock.
func cecSweepSourceClock() (func() time.Time, func()) {
	runtime.LockOSThread()
	return func() time.Time {
		var ts syscall.Timespec
		_, _, err := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, 3, uintptr(unsafe.Pointer(&ts)), 0)
		if err != 0 {
			panic(err)
		}
		return time.Unix(ts.Sec, ts.Nsec)
	}, runtime.UnlockOSThread
}
