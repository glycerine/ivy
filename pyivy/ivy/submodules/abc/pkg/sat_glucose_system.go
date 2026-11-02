package abc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Port of src/sat/glucose/System.h and System.cpp (namespace Gluco), Linux variant
// (the FreeBSD, Apple and generic branches are other platforms' definitions of the
// same two functions).

// GlucoCpuTime is Gluco::cpuTime (static inline; CPU-time in seconds).
// C: double Gluco::cpuTime(void)
func GlucoCpuTime() float64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	return float64(ru.Utime.Sec) + float64(ru.Utime.Usec)/1000000
}

// memReadStat_glucose is `static inline int memReadStat(int field)` (reads /proc/<pid>/statm).
// C: static inline int memReadStat(int field)
func memReadStat_glucose(field int) int {
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

// memReadPeak_glucose is `static inline int memReadPeak(void)` (reads VmPeak from /proc/<pid>/status).
// C: static inline int memReadPeak(void)
func memReadPeak_glucose() int {
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

// GlucoMemUsed is Gluco::memUsed (extern, defined in System.cpp; memory in megabytes).
// C: double Gluco::memUsed()
func GlucoMemUsed() float64 {
	return float64(memReadStat_glucose(0)) * float64(os.Getpagesize()) / (1024 * 1024)
}

// GlucoMemUsedPeak is Gluco::memUsedPeak (extern, defined in System.cpp; peak memory in megabytes).
// C: double Gluco::memUsedPeak()
func GlucoMemUsedPeak() float64 {
	peak := float64(memReadPeak_glucose() / 1024)
	if peak == 0 {
		return GlucoMemUsed()
	}
	return peak
}
