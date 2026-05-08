package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func TestTinyGoProbeImportsZ3OverWazero(t *testing.T) {
	if _, err := exec.LookPath("tinygo"); err != nil {
		t.Skipf("tinygo not found: %v", err)
	}

	tmp := t.TempDir()
	wasmPath := filepath.Join(tmp, "ivy-tinygo-probe.wasm")
	cmd := exec.Command("tinygo", "build",
		"-tags", "xtracer_off",
		"-target", "wasip1",
		"-o", wasmPath,
		".",
	)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+filepath.Join(tmp, "gocache"),
		"HOME="+filepath.Join(tmp, "home"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tinygo build failed: %v\n%s", err, out)
	}

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm probe: %v", err)
	}

	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	const boolSortID uint32 = 0x471
	var boolSortCalls atomic.Uint32
	_, err = rt.NewHostModuleBuilder("goivy_z3").
		NewFunctionBuilder().
		WithFunc(func() uint32 {
			boolSortCalls.Add(1)
			return boolSortID
		}).
		Export("bool_sort").
		Instantiate(ctx)
	if err != nil {
		t.Fatalf("instantiate goivy_z3 host module: %v", err)
	}

	mod, err := rt.InstantiateWithConfig(ctx, wasmBytes, wazero.NewModuleConfig().
		WithName("ivy_tinygo_probe").
		WithStartFunctions(),
	)
	if err != nil {
		t.Fatalf("instantiate tinygo probe: %v", err)
	}

	probeInitFn := mod.ExportedFunction("ivy_probe_init")
	if probeInitFn == nil {
		t.Fatal("tinygo probe did not export ivy_probe_init")
	}
	results, err := probeInitFn.Call(ctx)
	if err != nil {
		t.Fatalf("call ivy_probe_init: %v", err)
	}
	const wantChecksum uint32 = 0x47110571
	if got := uint32(results[0]); got != wantChecksum {
		t.Fatalf("ivy_probe_init() = %#x, want %#x", got, wantChecksum)
	}
	if got := boolSortCalls.Load(); got != 1 {
		t.Fatalf("host bool_sort calls after ivy_probe_init = %d, want 1", got)
	}

	boolSortIDFn := mod.ExportedFunction("ivy_probe_z3_bool_sort_id")
	if boolSortIDFn == nil {
		t.Fatal("tinygo probe did not export ivy_probe_z3_bool_sort_id")
	}
	results, err = boolSortIDFn.Call(ctx)
	if err != nil {
		t.Fatalf("call ivy_probe_z3_bool_sort_id: %v", err)
	}
	if got := uint32(results[0]); got != boolSortID {
		t.Fatalf("ivy_probe_z3_bool_sort_id() = %#x, want %#x", got, boolSortID)
	}

	checksumFn := mod.ExportedFunction("ivy_probe_checksum")
	if checksumFn == nil {
		t.Fatal("tinygo probe did not export ivy_probe_checksum")
	}
	results, err = checksumFn.Call(ctx)
	if err != nil {
		t.Fatalf("call ivy_probe_checksum: %v", err)
	}
	if got := uint32(results[0]); got != wantChecksum {
		t.Fatalf("ivy_probe_checksum() = %#x, want %#x", got, wantChecksum)
	}
}
