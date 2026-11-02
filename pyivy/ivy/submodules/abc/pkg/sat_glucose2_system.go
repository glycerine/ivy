package abc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Port of src/sat/glucose2/System.h and System2.cpp (namespace Gluco2), Linux variant
// (the FreeBSD, Apple and generic branches are other platforms' definitions of the
// same two functions).

// Gluco2CpuTime is Gluco2::cpuTime (static inline; CPU-time in seconds).
// C: double Gluco2::cpuTime(void)
func Gluco2CpuTime() float64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	return float64(ru.Utime.Sec) + float64(ru.Utime.Usec)/1000000
}

// memReadStat_glucose2 is `static inline int memReadStat(int field)` (reads /proc/<pid>/statm).
// C: static inline int memReadStat(int field)
func memReadStat_glucose2(field int) int {
	in, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", os.Getpid()))
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(in))
	value := 0
	for k := 0; k <= field; k++ { // one fscanf(in, "%d", &value) per iteration
		var perr error
		if k < len(fields) {
			value, perr = strconv.Atoi(fields[k])
		} else {
			perr = fmt.Errorf("missing field")
		}
		if perr != nil {
			fmt.Printf("ERROR! Failed to parse memory statistics from \"/proc\".\n")
			os.Exit(1)
		}
	}
	return value
}

// memReadPeak_glucose2 is `static inline int memReadPeak(void)` (reads VmPeak from /proc/<pid>/status).
// C: static inline int memReadPeak(void)
func memReadPeak_glucose2() int {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", os.Getpid()))
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	peakKb := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "VmPeak:") {
			if _, err := fmt.Sscanf(line, "VmPeak: %d kB", &peakKb); err == nil {
				break
			}
		}
	}
	return peakKb
}

// Gluco2MemUsed is Gluco2::memUsed (extern, defined in System.cpp; memory in megabytes).
// C: double Gluco2::memUsed()
func Gluco2MemUsed() float64 {
	return float64(memReadStat_glucose2(0)) * float64(os.Getpagesize()) / (1024 * 1024)
}

// Gluco2MemUsedPeak is Gluco2::memUsedPeak (extern, defined in System.cpp; peak memory in megabytes).
// C: double Gluco2::memUsedPeak()
func Gluco2MemUsedPeak() float64 {
	peak := float64(memReadPeak_glucose2() / 1024)
	if peak == 0 {
		return Gluco2MemUsed()
	}
	return peak
}
