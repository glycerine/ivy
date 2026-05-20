//go:build !xtrace_off && !js

package xtracer

import "runtime"

func maybeWriteHeapProfile(traceIndex int64) {}

func maybePaceGC(traceIndex int64, stats *runtime.MemStats) {}
