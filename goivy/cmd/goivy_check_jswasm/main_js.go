//go:build js && wasm

package main

// This package is the js/wasm browser entrypoint for goldweb. It must be
// hosted with the exact Go-team wasm_exec.js matching the Go compiler version
// that built it; goldweb serves the vendored wasm_exec-go1.25.6.js copy.

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall/js"

	goivy "github.com/glycerine/ivy/goivy"
)

type checkMeta struct {
	Filename string            `json:"filename"`
	Params   map[string]string `json:"params"`
}

const checkMetaNULPrefix = "goivy-meta-v1\x00"

const (
	defaultWasmMemoryLimitBytes = int64(3 << 30)
	defaultWasmGCPercent        = 50
)

var jsCallbacks []js.Func

func main() {
	tuneRuntime()

	run := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: goivyCheckRun requires spec and optional metadata")
			return 1
		}
		spec := args[0].String()
		metaRaw := ""
		if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
			metaRaw = args[1].String()
		}
		return int(runGoivyCheck(spec, metaRaw))
	})
	jsCallbacks = append(jsCallbacks, run)

	global := js.Global()
	global.Set("goivyCheckRun", run)
	global.Set("goivyCheckReady", true)
	select {}
}

func runGoivyCheck(spec, metaRaw string) (code int32) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n", r)
			code = 2
		}
	}()

	meta := checkMeta{Filename: "browser_input.ivy"}
	if metaRaw != "" {
		if err := parseCheckMeta(metaRaw, &meta); err != nil {
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

func parseCheckMeta(raw string, meta *checkMeta) error {
	if strings.HasPrefix(raw, checkMetaNULPrefix) {
		parts := strings.Split(strings.TrimPrefix(raw, checkMetaNULPrefix), "\x00")
		for i := 0; i < len(parts); {
			switch parts[i] {
			case "":
				i++
			case "filename":
				if i+1 >= len(parts) {
					return fmt.Errorf("missing filename value")
				}
				meta.Filename = parts[i+1]
				i += 2
			case "param":
				if i+2 >= len(parts) {
					return fmt.Errorf("missing param key/value")
				}
				if meta.Params == nil {
					meta.Params = make(map[string]string)
				}
				meta.Params[parts[i+1]] = parts[i+2]
				i += 3
			default:
				return fmt.Errorf("unknown metadata field %q", parts[i])
			}
		}
		return nil
	}
	return json.Unmarshal([]byte(raw), meta)
}

func tuneRuntime() {
	limit := defaultWasmMemoryLimitBytes
	if raw := os.Getenv("GOIVY_WASM_MEMORY_LIMIT"); raw != "" {
		parsed, err := parseByteLimit(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[goivy js/wasm] ignoring bad GOIVY_WASM_MEMORY_LIMIT=%q: %v\n", raw, err)
		} else {
			limit = parsed
		}
	}
	previousLimit := debug.SetMemoryLimit(limit)

	gcPercent := defaultWasmGCPercent
	if raw := os.Getenv("GOIVY_WASM_GOGC"); raw != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[goivy js/wasm] ignoring bad GOIVY_WASM_GOGC=%q: %v\n", raw, err)
		} else {
			gcPercent = parsed
		}
	}
	previousGCPercent := debug.SetGCPercent(gcPercent)

	fmt.Fprintf(os.Stderr, "[goivy js/wasm] memory_limit=%d previous_memory_limit=%d gogc=%d previous_gogc=%d\n",
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
