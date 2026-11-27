package main

import (
	"bufio"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

// abcGCPercent is the collector's target unless GOGC is set. ABC commands
// rebuild whole networks of pointer-rich objects, and with GOMAXPROCS=1 the
// collector runs on the same thread: a larger heap target cut restrash,
// lutpack and resyn2 times by 10-20% on mul64.
const abcGCPercent = 400

// setGCDefaults applies abcGCPercent and, unless GOMEMLIMIT is set, a soft
// memory limit of half the physical memory, so large designs still collect
// early rather than grow the heap fourfold.
func setGCDefaults() {
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(abcGCPercent)
	}
	if os.Getenv("GOMEMLIMIT") == "" {
		if total := physicalMemory(); total > 0 {
			debug.SetMemoryLimit(total / 2)
		}
	}
}

// physicalMemory reads MemTotal from /proc/meminfo (0 when unavailable).
func physicalMemory() int64 {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) >= 2 && fields[0] == "MemTotal:" {
			kb, err := strconv.ParseInt(fields[1], 10, 64)
			if err == nil {
				return kb * 1024
			}
		}
	}
	return 0
}
