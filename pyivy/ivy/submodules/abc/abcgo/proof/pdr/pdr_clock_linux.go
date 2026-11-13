//go:build linux

package pdr

import (
	"runtime"
	"syscall"
	"unsafe"
)

// ABC's Linux Abc_Clock uses the calling thread's CPU time. Return
// nanoseconds quantized to its microsecond CLOCKS_PER_SEC ticks.
func pdrCPUClock() int64 {
	var ts syscall.Timespec
	_, _, errno := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, 3, uintptr(unsafe.Pointer(&ts)), 0)
	if errno != 0 {
		return -1
	}
	return (int64(ts.Sec)*1e6 + int64(ts.Nsec)/1e3) * 1000
}
func pdrPinClock() func() { runtime.LockOSThread(); return runtime.UnlockOSThread }
