//go:build !xtracer_off && !wasip1

package xtracer

import "runtime"

func maybeWriteHeapProfile(traceIndex int64) {}

func maybePaceGC(traceIndex int64, stats *runtime.MemStats) {}
