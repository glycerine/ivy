package aig

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestXorLevelConventionsAgainstCompiledSource(t *testing.T) {
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
	binary := filepath.Join(t.TempDir(), "xor-levels")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src,
		filepath.Join(dir, "test_vectors/xor_levels_compare.c"), filepath.Join(src, "aig/aig/aigUtil.c"),
		filepath.Join(src, "misc/tim/timBox.c"), "-Wl,--gc-sections", "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile C oracle: %v: %s", err, out)
	}
	out, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run C oracle: %v: %s", err, out)
	}
	want := "structural 2 3 2\nchoice 2 1 2 0\narrival 6 7\n"
	if string(out) != want {
		t.Fatalf("source XOR level conventions:\n%swant:\n%s", out, want)
	}
}
