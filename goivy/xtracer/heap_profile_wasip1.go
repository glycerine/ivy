//go:build !xtracer_off && (wasm || wasip1)

package xtracer

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
)

const defaultGCHeapLimit = uint64(512 << 20)
const heapProfileFD = uintptr(4)
const heapProfileHeaderMagic = "GOLDWEB_HEAP_PROFILE_V1"

var (
	profileConfigOnce sync.Once
	profileIndexes    map[int64]struct{}
	gcHeapLimit       uint64
)

func maybeWriteHeapProfile(traceIndex int64) {
	profileConfigOnce.Do(loadProfileConfig)
	if _, ok := profileIndexes[traceIndex]; !ok {
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
	writeHeapProfileBytes(uint32(traceIndex), data)
}

func writeHeapProfileBytes(traceIndex uint32, data []byte) {
	f := os.NewFile(heapProfileFD, "goldweb-heap-profile")
	if f == nil {
		fmt.Printf("[xtracer] heap profile fd %d unavailable\n", heapProfileFD)
		return
	}
	if _, err := fmt.Fprintf(f, "%s %d %d %d\n", heapProfileHeaderMagic, 1, traceIndex, len(data)); err != nil {
		fmt.Printf("[xtracer] write heap profile header: %v\n", err)
		return
	}
	if len(data) > 0 {
		if _, err := f.Write(data); err != nil {
			fmt.Printf("[xtracer] write heap profile body: %v\n", err)
		}
	}
}

func maybePaceGC(traceIndex int64, stats *runtime.MemStats) {
	profileConfigOnce.Do(loadProfileConfig)
	if gcHeapLimit == 0 || stats == nil || stats.HeapAlloc <= gcHeapLimit {
		return
	}

	beforeAlloc := stats.HeapAlloc
	beforeInuse := stats.HeapInuse
	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	fmt.Printf("[at trace %v] wasm.gc HeapAlloc = %0.3f MB -> %0.3f MB; HeapInuse = %0.3f MB -> %0.3f MB; limit = %0.3f MB\n",
		traceIndex,
		float64(beforeAlloc)/(1<<20),
		float64(after.HeapAlloc)/(1<<20),
		float64(beforeInuse)/(1<<20),
		float64(after.HeapInuse)/(1<<20),
		float64(gcHeapLimit)/(1<<20))
}

func loadProfileConfig() {
	profileIndexes = map[int64]struct{}{
		125_000: {},
		150_000: {},
		166_000: {},
		175_000: {},
	}
	if raw := strings.TrimSpace(os.Getenv("GOIVY_WASM_XTRACE_HEAPPROFILES")); raw != "" {
		profileIndexes = parseProfileIndexes(raw)
	}

	gcHeapLimit = defaultGCHeapLimit
	if raw := strings.TrimSpace(os.Getenv("GOIVY_WASM_XTRACE_GC_HEAP_LIMIT")); raw != "" {
		limit, err := parseByteLimit(raw)
		if err != nil {
			fmt.Printf("[xtracer] ignoring bad GOIVY_WASM_XTRACE_GC_HEAP_LIMIT=%q: %v\n", raw, err)
		} else {
			gcHeapLimit = limit
		}
	}
}

func parseProfileIndexes(raw string) map[int64]struct{} {
	indexes := make(map[int64]struct{})
	for _, field := range strings.Split(raw, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if field == "0" || strings.EqualFold(field, "off") || strings.EqualFold(field, "none") {
			return indexes
		}
		index, err := strconv.ParseInt(field, 10, 64)
		if err != nil || index < 0 {
			fmt.Printf("[xtracer] ignoring bad GOIVY_WASM_XTRACE_HEAPPROFILES entry %q\n", field)
			continue
		}
		indexes[index] = struct{}{}
	}
	return indexes
}

func parseByteLimit(raw string) (uint64, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return 0, fmt.Errorf("empty limit")
	}

	multiplier := uint64(1)
	for _, suffix := range []struct {
		text       string
		multiplier uint64
	}{
		{"gib", 1 << 30},
		{"gb", 1000 * 1000 * 1000},
		{"mib", 1 << 20},
		{"mb", 1000 * 1000},
		{"kib", 1 << 10},
		{"kb", 1000},
		{"b", 1},
	} {
		if strings.HasSuffix(s, suffix.text) {
			multiplier = suffix.multiplier
			s = strings.TrimSpace(strings.TrimSuffix(s, suffix.text))
			break
		}
	}

	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if value == 0 {
		return 0, nil
	}
	if value > ^uint64(0)/multiplier {
		return 0, fmt.Errorf("limit overflows uint64")
	}
	return value * multiplier, nil
}
