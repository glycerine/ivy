//go:build linux

package abs

import (
	"runtime"
	"syscall"
	"unsafe"
)

// ABC's Linux Abc_Clock uses CLOCK_THREAD_CPUTIME_ID. The caller must keep
// the goroutine pinned for the entire interval between readings.
func absCPUClock() int64 {
	var ts syscall.Timespec
	_, _, errno := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, 3, uintptr(unsafe.Pointer(&ts)), 0)
	if errno != 0 {
		panic(errno)
	}
	return int64(ts.Sec)*1e9 + int64(ts.Nsec)
}

func absPinClockThread() func() {
	runtime.LockOSThread()
	return runtime.UnlockOSThread
}
