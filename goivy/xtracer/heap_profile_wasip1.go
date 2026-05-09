//go:build !xtracer_off && wasip1

package xtracer

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/pprof"
	"unsafe"
)

const heapProfileXTraceIndex int64 = 150_000

//go:wasmimport goldweb xtrace_heap_profile
func goldwebXTraceHeapProfile(xtraceIndex uint32, ptr uint32, len uint32)

func maybeWriteHeapProfile(traceIndex int64) {
	if traceIndex != heapProfileXTraceIndex {
		return
	}

	runtime.GC()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	var buf bytes.Buffer
	if err := pprof.WriteHeapProfile(&buf); err != nil {
		fmt.Printf("[at trace %v] memprof error = %v\n", traceIndex, err)
		return
	}

	data := buf.Bytes()
	fmt.Printf("[at trace %v] memprof bytes = %d; HeapAlloc = %0.3f MB; HeapInuse = %0.3f MB\n",
		traceIndex, len(data), float64(stats.HeapAlloc)/(1<<20), float64(stats.HeapInuse)/(1<<20))
	if len(data) == 0 {
		goldwebXTraceHeapProfile(uint32(traceIndex), 0, 0)
		return
	}
	goldwebXTraceHeapProfile(uint32(traceIndex), uint32(uintptr(unsafe.Pointer(&data[0]))), uint32(len(data)))
	runtime.KeepAlive(data)
}
