package main

import (
	"runtime/debug"
	"testing"
)

func TestSetGCDefaultsRespectsEnvironment(t *testing.T) {
	old := debug.SetGCPercent(100)
	defer debug.SetGCPercent(old)
	defer debug.SetMemoryLimit(debug.SetMemoryLimit(-1))
	t.Setenv("GOGC", "")
	t.Setenv("GOMEMLIMIT", "")
	setGCDefaults()
	if got := debug.SetGCPercent(100); got != abcGCPercent {
		t.Fatalf("GC percent %d, want %d", got, abcGCPercent)
	}
	t.Setenv("GOGC", "150")
	setGCDefaults()
	if got := debug.SetGCPercent(100); got != 100 {
		t.Fatalf("GOGC set: GC percent changed to %d", got)
	}
	if physicalMemory() <= 0 {
		t.Skip("no /proc/meminfo")
	}
}
