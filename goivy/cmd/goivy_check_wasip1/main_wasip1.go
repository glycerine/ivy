//go:build wasip1

package main

// This is not a second goivy_check CLI and not an additional Go Ivy runtime.
// Big Go needs a package main as the build root for an exporting wasip1 module;
// this tiny adapter is that root. The produced artifact is the single browser
// Go Ivy wasm blob that goldweb loads, containing goivy plus these exported
// in-memory command functions.

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	goivy "github.com/glycerine/ivy/goivy"
)

type checkMeta struct {
	Filename string            `json:"filename"`
	Params   map[string]string `json:"params"`
}

type checkInput struct {
	spec []byte
	meta []byte
}

var pendingInput checkInput
var tuneRuntimeOnce sync.Once

const (
	defaultWasmMemoryLimitBytes = int64(3 << 30)
	defaultWasmGCPercent        = 50
	maxCheckInputBytes          = uint32(256 << 20)
)

func main() {
	tuneRuntimeOnce.Do(tuneRuntime)
}

//go:wasmexport goivy_check_prepare
func goivyCheckPrepare(specLen, metaLen uint32) int32 {
	if specLen > maxCheckInputBytes || metaLen > maxCheckInputBytes {
		return 1
	}
	pendingInput = checkInput{
		spec: make([]byte, specLen),
		meta: make([]byte, metaLen),
	}
	return 0
}

//go:wasmexport goivy_check_write_spec
func goivyCheckWriteSpec(offset, word, n uint32) int32 {
	return writePackedBytes(pendingInput.spec, offset, word, n)
}

//go:wasmexport goivy_check_write_meta
func goivyCheckWriteMeta(offset, word, n uint32) int32 {
	return writePackedBytes(pendingInput.meta, offset, word, n)
}

//go:wasmexport goivy_check_run
func goivyCheckRun() (code int32) {
	tuneRuntimeOnce.Do(tuneRuntime)
	stopHeapProfiler := startHeapProfiler()
	defer stopHeapProfiler()

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n", r)
			code = 2
		}
	}()

	spec := string(pendingInput.spec)
	metaRaw := string(pendingInput.meta)
	pendingInput = checkInput{}

	meta := checkMeta{Filename: "browser_input.ivy"}
	if metaRaw != "" {
		if err := json.Unmarshal([]byte(metaRaw), &meta); err != nil {
			fmt.Fprintf(os.Stderr, "error: bad goivy_check metadata: %v\n", err)
			return 1
		}
	}
	if meta.Filename == "" {
		meta.Filename = "browser_input.ivy"
	}
	if meta.Params == nil {
		meta.Params = make(map[string]string)
	}

	cfg := goivy.NewConfig()
	if err := goivy.ApplyIvyCheckParams(cfg, meta.Params); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if err := goivy.StartSourceWithConfig(meta.Filename, spec, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func writePackedBytes(dst []byte, offset, word, n uint32) int32 {
	if n > 4 {
		return 1
	}
	if offset > uint32(len(dst)) || n > uint32(len(dst))-offset {
		return 1
	}
	for i := uint32(0); i < n; i++ {
		dst[offset+i] = byte(word >> (8 * i))
	}
	return 0
}

func tuneRuntime() {
	limit := defaultWasmMemoryLimitBytes
	if raw := os.Getenv("GOIVY_WASM_MEMORY_LIMIT"); raw != "" {
		parsed, err := parseByteLimit(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[goivy wasm] ignoring bad GOIVY_WASM_MEMORY_LIMIT=%q: %v\n", raw, err)
		} else {
			limit = parsed
		}
	}
	previousLimit := debug.SetMemoryLimit(limit)

	gcPercent := defaultWasmGCPercent
	if raw := os.Getenv("GOIVY_WASM_GOGC"); raw != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[goivy wasm] ignoring bad GOIVY_WASM_GOGC=%q: %v\n", raw, err)
		} else {
			gcPercent = parsed
		}
	}
	previousGCPercent := debug.SetGCPercent(gcPercent)

	fmt.Fprintf(os.Stderr, "[goivy wasm] memory_limit=%d previous_memory_limit=%d gogc=%d previous_gogc=%d\n",
		limit, previousLimit, gcPercent, previousGCPercent)
}

func parseByteLimit(raw string) (int64, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return 0, fmt.Errorf("empty limit")
	}

	multiplier := int64(1)
	for _, suffix := range []struct {
		text       string
		multiplier int64
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

	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("limit must be positive")
	}
	if value > (1<<63-1)/multiplier {
		return 0, fmt.Errorf("limit overflows int64")
	}
	return value * multiplier, nil
}
