//go:build aix || darwin || dragonfly || freebsd || netbsd || openbsd || solaris

package int

import "syscall"

// ABC uses clock() on these platforms, which reports process CPU time.
func interCPUClock() int64 {
	var usage syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &usage) != nil {
		return -1
	}
	return ((int64(usage.Utime.Sec)+int64(usage.Stime.Sec))*1e6 + int64(usage.Utime.Usec) + int64(usage.Stime.Usec)) * 1000
}
func interPinClock() func() { return func() {} }
