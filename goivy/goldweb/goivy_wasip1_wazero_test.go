package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type wazeroCheckMeta struct {
	Filename string            `json:"filename"`
	Params   map[string]string `json:"params"`
}

// TestGoIvyCheckWasip1UnderWazero is a manual crash-localization probe.
//
// It runs the same Big Go wasip1 browser artifact that goldweb serves, but
// under wazero instead of the browser's JavaScript WASI shim. This isolates
// Go-wasm/runtime failures from browser-host memory writes. Z3 imports are
// stubbed because the current crash under investigation happens before solver
// traffic; reaching a stubbed Z3 call means this probe has already passed that
// crash point.
//
// Example:
//
//	GOIVY_WAZERO_SPEC=/Users/jaten/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy \
//	go test ./goldweb -run TestGoIvyCheckWasip1UnderWazero -count=1 -v
func TestGoIvyCheckWasip1UnderWazero(t *testing.T) {
	specPath := os.Getenv("GOIVY_WAZERO_SPEC")
	if specPath == "" {
		t.Skip("set GOIVY_WAZERO_SPEC to run the wasip1 artifact under wazero")
	}

	root := filepath.Clean("..")
	wasmPath := os.Getenv("GOIVY_WAZERO_WASM")
	if wasmPath == "" {
		wasmPath = filepath.Join(root, "webvue", "static", "goivy-check-wasip1.wasm")
	}
	includeDir := os.Getenv("GOIVY_WAZERO_INCLUDE")
	if includeDir == "" {
		includeDir = defaultIncludeDir(root)
	}
	stdoutPath := os.Getenv("GOIVY_WAZERO_STDOUT")
	if stdoutPath == "" {
		stdoutPath = filepath.Join(root, "wazero.browser.xtrace.log")
	}
	stderrPath := os.Getenv("GOIVY_WAZERO_STDERR")
	if stderrPath == "" {
		stderrPath = filepath.Join(root, "wazero.stderr.log")
	}

	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}
	stdout, err := os.Create(stdoutPath)
	if err != nil {
		t.Fatalf("create stdout log: %v", err)
	}
	defer stdout.Close()
	stderr, err := os.Create(stderrPath)
	if err != nil {
		t.Fatalf("create stderr log: %v", err)
	}
	defer stderr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		t.Fatalf("compile wasm: %v", err)
	}
	defer compiled.Close(ctx)

	if err := instantiateZ3Stubs(ctx, rt, compiled); err != nil {
		t.Fatalf("instantiate smt_z3 stubs: %v", err)
	}
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	moduleConfig := wazero.NewModuleConfig().
		WithName("goivy_check_wasip1").
		WithArgs("goivy_check_wasip1").
		// Match goldweb's browser host: the include tree is preopened at the
		// same absolute path reported by GOIVY_INCLUDE, while the primary spec
		// itself is passed in memory through goivy_check_write_spec.
		WithEnv("GOIVY_INCLUDE", includeDir).
		WithEnv("GOIVY_WASM_HEAPPROFILE_INTERVAL", "0").
		WithFSConfig(wazero.NewFSConfig().WithReadOnlyDirMount(includeDir, includeDir)).
		WithStdout(stdout).
		WithStderr(stderr)
	if gcHeapLimit := os.Getenv("GOIVY_WAZERO_GC_HEAP_LIMIT"); gcHeapLimit != "" {
		moduleConfig = moduleConfig.WithEnv("GOIVY_WASM_XTRACE_GC_HEAP_LIMIT", gcHeapLimit)
	}
	mod, err := rt.InstantiateModule(ctx, compiled, moduleConfig)
	if err != nil {
		t.Fatalf("instantiate wasm: %v", err)
	}
	defer mod.Close(ctx)
	initialize := mod.ExportedFunction("_initialize")
	if initialize == nil {
		t.Fatalf("%s is not a wasip1 reactor: missing _initialize; rebuild with -buildmode=c-shared", wasmPath)
	}
	if _, err := initialize.Call(ctx); err != nil {
		t.Fatalf("_initialize trapped: %v", err)
	}

	meta, err := json.Marshal(wazeroCheckMeta{
		Filename: specPath,
		Params:   map[string]string{},
	})
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}

	if code := callExportU32(t, ctx, mod, "goivy_check_prepare", uint64(len(spec)), uint64(len(meta))); code != 0 {
		t.Fatalf("goivy_check_prepare returned %d", code)
	}
	writeBytesToWasm(t, ctx, mod, "goivy_check_write_spec", spec)
	writeBytesToWasm(t, ctx, mod, "goivy_check_write_meta", meta)

	start := time.Now()
	results, err := mod.ExportedFunction("goivy_check_run").Call(ctx)
	t.Logf("goivy_check_run elapsed=%s stdout=%s stderr=%s", time.Since(start), stdoutPath, stderrPath)
	if err != nil {
		t.Fatalf("goivy_check_run trapped: %v", err)
	}
	if got := uint32(results[0]); got != 0 {
		t.Fatalf("goivy_check_run returned %d", got)
	}
}

// TestGoIvyCheckCLIWasip1UnderWazero runs the normal goivy_check CLI compiled
// to wasip1 under wazero. Unlike the browser exported-function artifact, the
// CLI naturally does all work during _start, so wazero's proc_exit semantics are
// useful instead of getting in the way.
//
// Example:
//
//	GOIVY_WAZERO_CLI_WASM=/private/tmp/goivy-check-cli-wasip1.wasm \
//	GOIVY_WAZERO_CLI_SPEC=/Users/jaten/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy \
//	go test ./goldweb -run TestGoIvyCheckCLIWasip1UnderWazero -count=1 -v
func TestGoIvyCheckCLIWasip1UnderWazero(t *testing.T) {
	specPath := os.Getenv("GOIVY_WAZERO_CLI_SPEC")
	wasmPath := os.Getenv("GOIVY_WAZERO_CLI_WASM")
	if specPath == "" || wasmPath == "" {
		t.Skip("set GOIVY_WAZERO_CLI_SPEC and GOIVY_WAZERO_CLI_WASM to run the CLI wasip1 artifact under wazero")
	}

	root := filepath.Clean("..")
	includeDir := os.Getenv("GOIVY_WAZERO_INCLUDE")
	if includeDir == "" {
		includeDir = defaultIncludeDir(root)
	}
	stdoutPath := os.Getenv("GOIVY_WAZERO_CLI_STDOUT")
	if stdoutPath == "" {
		stdoutPath = filepath.Join(root, "wazero.cli.xtrace.log")
	}
	stderrPath := os.Getenv("GOIVY_WAZERO_CLI_STDERR")
	if stderrPath == "" {
		stderrPath = filepath.Join(root, "wazero.cli.stderr.log")
	}

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}
	stdout, err := os.Create(stdoutPath)
	if err != nil {
		t.Fatalf("create stdout log: %v", err)
	}
	defer stdout.Close()
	stderr, err := os.Create(stderrPath)
	if err != nil {
		t.Fatalf("create stderr log: %v", err)
	}
	defer stderr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		t.Fatalf("compile wasm: %v", err)
	}
	defer compiled.Close(ctx)

	if err := instantiateZ3Stubs(ctx, rt, compiled); err != nil {
		t.Fatalf("instantiate smt_z3 stubs: %v", err)
	}
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	start := time.Now()
	moduleConfig := wazero.NewModuleConfig().
		WithName("goivy_check_cli_wasip1").
		WithArgs("goivy_check", specPath).
		WithEnv("GOIVY_INCLUDE", includeDir).
		WithEnv("GOIVY_WASM_HEAPPROFILE_INTERVAL", "0").
		WithFSConfig(wazero.NewFSConfig().
			WithReadOnlyDirMount("/Users/jaten/ivy", "/Users/jaten/ivy").
			WithReadOnlyDirMount(includeDir, includeDir)).
		WithStdout(stdout).
		WithStderr(stderr)
	if gcHeapLimit := os.Getenv("GOIVY_WAZERO_CLI_GC_HEAP_LIMIT"); gcHeapLimit != "" {
		moduleConfig = moduleConfig.WithEnv("GOIVY_WASM_XTRACE_GC_HEAP_LIMIT", gcHeapLimit)
	}
	_, err = rt.InstantiateModule(ctx, compiled, moduleConfig)
	t.Logf("cli _start elapsed=%s stdout=%s stderr=%s", time.Since(start), stdoutPath, stderrPath)
	if err != nil {
		t.Fatalf("instantiate/run CLI wasm: %v", err)
	}
}

func instantiateZ3Stubs(ctx context.Context, rt wazero.Runtime, compiled wazero.CompiledModule) error {
	builder := rt.NewHostModuleBuilder("smt_z3")
	seen := make(map[string]struct{})
	for _, fn := range compiled.ImportedFunctions() {
		moduleName, importName, ok := fn.Import()
		if !ok || moduleName != "smt_z3" {
			continue
		}
		if _, ok := seen[importName]; ok {
			continue
		}
		seen[importName] = struct{}{}
		params := append([]api.ValueType(nil), fn.ParamTypes()...)
		results := append([]api.ValueType(nil), fn.ResultTypes()...)
		builder.NewFunctionBuilder().WithGoFunction(api.GoFunc(func(ctx context.Context, stack []uint64) {
			for i := range results {
				stack[i] = 0
			}
		}), params, results).Export(importName)
	}
	_, err := builder.Instantiate(ctx)
	return err
}

func writeBytesToWasm(t *testing.T, ctx context.Context, mod api.Module, exportName string, data []byte) {
	t.Helper()
	fn := mod.ExportedFunction(exportName)
	if fn == nil {
		t.Fatalf("missing export %s", exportName)
	}
	for offset := 0; offset < len(data); offset += 4 {
		n := len(data) - offset
		if n > 4 {
			n = 4
		}
		var wordBytes [4]byte
		copy(wordBytes[:], data[offset:offset+n])
		word := binary.LittleEndian.Uint32(wordBytes[:])
		results, err := fn.Call(ctx, uint64(offset), uint64(word), uint64(n))
		if err != nil {
			t.Fatalf("%s offset=%d trapped: %v", exportName, offset, err)
		}
		if got := uint32(results[0]); got != 0 {
			t.Fatalf("%s offset=%d returned %d", exportName, offset, got)
		}
	}
}

func callExportU32(t *testing.T, ctx context.Context, mod api.Module, exportName string, args ...uint64) uint32 {
	t.Helper()
	fn := mod.ExportedFunction(exportName)
	if fn == nil {
		t.Fatalf("missing export %s", exportName)
	}
	results, err := fn.Call(ctx, args...)
	if err != nil {
		t.Fatalf("%s trapped: %v", exportName, err)
	}
	return uint32(results[0])
}
