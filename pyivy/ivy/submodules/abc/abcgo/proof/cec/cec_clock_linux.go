//go:build linux

package cec

import (
	"runtime"
	"syscall"
	"unsafe"
)

// ABC's Linux Abc_Clock measures CPU time of the calling OS thread.
func cecCPUClock() int64 {
	var ts syscall.Timespec
	_, _, errno := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, 3, uintptr(unsafe.Pointer(&ts)), 0)
	if errno != 0 {
		panic(errno)
	}
	return int64(ts.Sec)*1e9 + int64(ts.Nsec)
}

func cecPinClockThread() func() {
	runtime.LockOSThread()
	return runtime.UnlockOSThread
}
