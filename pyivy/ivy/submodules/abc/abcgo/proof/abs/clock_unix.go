//go:build aix || darwin || dragonfly || freebsd || netbsd || openbsd || solaris

package abs

import "syscall"

// ABC falls back to clock(), which measures process CPU time on POSIX.
// RUSAGE_SELF includes the user and system CPU time of all process threads.
func absCPUClock() int64 {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		panic(err)
	}
	return (int64(usage.Utime.Sec)+int64(usage.Stime.Sec))*1e9 +
		(int64(usage.Utime.Usec)+int64(usage.Stime.Usec))*1e3
}

func absPinClockThread() func() { return func() {} }
