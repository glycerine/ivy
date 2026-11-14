package aig

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestSupportFrontierAgainstCompiledSource(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture uses GNU linker section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "factor-frontier")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, filepath.Join(dir, "test_vectors/factor_frontier_compare.c"), "-Wl,--gc-sections", "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile source frontier: %v: %s", err, out)
	}
	out, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run source frontier: %v: %s", err, out)
	}
	if want := "support: 2 3\nmarks: 0 7 7 7 7 7\n"; string(out) != want {
		t.Fatalf("source frontier = %q, want %q", out, want)
	}
}
