package aig

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestIncrementalTimingAgainstCompiledSource(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture uses GNU linker section garbage collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "incremental-timing")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src,
		filepath.Join(dir, "test_vectors/incremental_timing_compare.c"),
		filepath.Join(src, "aig/aig/aigFanout.c"), filepath.Join(src, "aig/aig/aigUtil.c"),
		"-Wl,--gc-sections", "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile C oracle: %v: %s", err, out)
	}
	out, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run C oracle: %v: %s", err, out)
	}
	// The matching Go regressions check these same seeded/increase/decrease
	// fixtures and the reverse queue's dynamically extended old-level bucket.
	want := "forward0 5 2 17 1 2\nforward1 5 6 1 0 2\nforward2 1 2 1 0 6\nreverse 1 2 3 13 000\n"
	if string(out) != want {
		t.Fatalf("original C timing states:\n%swant:\n%s", out, want)
	}
}
