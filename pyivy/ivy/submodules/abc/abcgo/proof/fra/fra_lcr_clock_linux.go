//go:build linux

package fra

import (
	"runtime"
	"syscall"
	"unsafe"
)

// ABC's Linux Abc_Clock uses the calling thread's CPU time and returns
// CLOCKS_PER_SEC ticks (microseconds on Linux).
func fraLcrClock() int64 {
	var ts syscall.Timespec
	_, _, errno := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, 3, uintptr(unsafe.Pointer(&ts)), 0)
	if errno != 0 {
		return -1
	}
	return int64(ts.Sec)*1e6 + int64(ts.Nsec)/1e3
}
func fraLcrPin() func() { runtime.LockOSThread(); return runtime.UnlockOSThread }
