package aig

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestReferenceAccessorsAgainstCompiledSource(t *testing.T) {
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "reference-accessors")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-I", src,
		filepath.Join(dir, "test_vectors/reference_accessors_compare.c"), "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile original inline accessors: %v: %s", err, out)
	}
	out, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run C accessors: %v: %s", err, out)
	}
	want := "refs 0 1 0 0\nhead 2\nempty 0\n"
	if string(out) != want {
		t.Fatalf("original reference/accessor states:\n%swant:\n%s", out, want)
	}
}
