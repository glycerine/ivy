//go:build !xtracer_off && !wasip1

package xtracer

func maybeWriteHeapProfile(traceIndex int64) {
	_ = traceIndex
}
