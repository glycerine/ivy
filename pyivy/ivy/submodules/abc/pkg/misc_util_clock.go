package abc

import "syscall"

// AbcClockTicksPerSecond is CLOCKS_PER_SEC on the platforms ABC supports (POSIX requires 1000000).
const AbcClockTicksPerSecond = 1000000

// AbcClock is Abc_Clock (misc/util/abc_global.h): the CPU time of the caller in clock
// ticks (abctime), or -1 on failure. The C code reads CLOCK_THREAD_CPUTIME_ID; a
// goroutine has no stable thread, so the CPU time of the process is used.
// C: static inline abctime Abc_Clock()
func AbcClock() int64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return -1
	}
	sec := int64(ru.Utime.Sec) + int64(ru.Stime.Sec)
	usec := int64(ru.Utime.Usec) + int64(ru.Stime.Usec)
	return sec*AbcClockTicksPerSecond + usec*AbcClockTicksPerSecond/1000000
}
