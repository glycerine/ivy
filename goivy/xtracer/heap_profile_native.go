//go:build !xtracer_off && !wasip1

package xtracer

import "runtime"

func maybeWriteHeapProfile(traceIndex int64) {
	_ = traceIndex
}

func maybePaceGC(traceIndex int64, stats *runtime.MemStats) {
	_ = traceIndex
	_ = stats
}
