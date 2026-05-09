//go:build wasip1

package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

// note that the following goldweb.heap_profile callback is
// implemented inside the embedded worker JS
// in goldweb.go (line 1787).

//go:wasmimport goldweb heap_profile
func goldwebHeapProfile(elapsedSeconds uint32, ptr *byte, len uint32)

const defaultHeapProfileInterval = 10 * time.Second

func startHeapProfiler() func() {
	interval := heapProfileInterval()
	if interval <= 0 {
		return func() {}
	}

	if raw := os.Getenv("GOIVY_WASM_MEMPROFILE_RATE"); raw != "" {
		rate, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[goivy wasm] ignoring bad GOIVY_WASM_MEMPROFILE_RATE=%q: %v\n", raw, err)
		} else {
			runtime.MemProfileRate = rate
		}
	}

	started := time.Now()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writeHeapProfile(uint32(time.Since(started) / time.Second))
			case <-stop:
				return
			}
		}
	}()

	fmt.Fprintf(os.Stderr, "[goivy wasm] heap profiles enabled interval=%s mem_profile_rate=%d\n",
		interval, runtime.MemProfileRate)

	return func() {
		close(stop)
		<-done
	}
}

func heapProfileInterval() time.Duration {
	raw := strings.TrimSpace(os.Getenv("GOIVY_WASM_HEAPPROFILE_INTERVAL"))
	if raw == "" {
		return defaultHeapProfileInterval
	}
	if raw == "0" {
		return 0
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[goivy wasm] ignoring bad GOIVY_WASM_HEAPPROFILE_INTERVAL=%q: %v\n", raw, err)
		return 0
	}
	return d
}

func writeHeapProfile(elapsedSeconds uint32) {
	runtime.GC()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	var buf bytes.Buffer
	if err := pprof.WriteHeapProfile(&buf); err != nil {
		fmt.Fprintf(os.Stderr, "[goivy wasm] write heap profile elapsed=%ds: %v\n", elapsedSeconds, err)
		return
	}

	data := buf.Bytes()
	fmt.Fprintf(os.Stderr, "[goivy wasm] heap_profile elapsed=%ds bytes=%d heap_alloc=%d heap_inuse=%d heap_sys=%d next_gc=%d num_gc=%d\n",
		elapsedSeconds, len(data), stats.HeapAlloc, stats.HeapInuse, stats.HeapSys, stats.NextGC, stats.NumGC)
	if len(data) == 0 {
		goldwebHeapProfile(elapsedSeconds, nil, 0)
		return
	}
	goldwebHeapProfile(elapsedSeconds, &data[0], uint32(len(data)))
	runtime.KeepAlive(data)
}
