package abc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Port of src/sat/bsat2/System.h and System.cpp (namespace Minisat), Linux variant
// (the FreeBSD, Apple and generic branches are other platforms' definitions of the
// same two functions).

// MinisatCpuTime is Minisat::cpuTime (static inline; CPU-time in seconds).
// C: double Minisat::cpuTime(void)
func MinisatCpuTime() float64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	return float64(ru.Utime.Sec) + float64(ru.Utime.Usec)/1000000
}

// memReadStat_bsat2 is `static inline int memReadStat(int field)` (reads /proc/<pid>/statm).
// C: static inline int memReadStat(int field)
func memReadStat_bsat2(field int) int {
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

// memReadPeak_bsat2 is `static inline int memReadPeak(void)` (reads VmPeak from /proc/<pid>/status).
// C: static inline int memReadPeak(void)
func memReadPeak_bsat2() int {
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

// MinisatMemUsed is Minisat::memUsed (extern, defined in System.cpp; memory in megabytes).
// C: double Minisat::memUsed()
func MinisatMemUsed() float64 {
	return float64(memReadStat_bsat2(0)) * float64(os.Getpagesize()) / (1024 * 1024)
}

// MinisatMemUsedPeak is Minisat::memUsedPeak (extern, defined in System.cpp; peak memory in megabytes).
// C: double Minisat::memUsedPeak()
func MinisatMemUsedPeak() float64 {
	peak := float64(memReadPeak_bsat2() / 1024)
	if peak == 0 {
		return MinisatMemUsed()
	}
	return peak
}
