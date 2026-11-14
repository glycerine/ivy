package aig

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func runAigSourceFixture(t *testing.T, fixture string, sources ...string) string {
	t.Helper()
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
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, filepath.Join(dir, "test_vectors", fixture)}
	for _, name := range sources {
		args = append(args, filepath.Join(src, "aig/aig", name))
	}
	args = append(args, "-Wl,--gc-sections", "-o", binary)
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile source frontier: %v: %s", err, out)
	}
	out, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run source frontier: %v: %s", err, out)
	}
	return string(out)
}

func TestSupportFrontierAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "factor_frontier_compare.c")
	if want := "support: 2 3\nmarks: 0 7 7 7 7 7\n"; out != want {
		t.Fatalf("source frontier = %q, want %q", out, want)
	}
}

func TestFactoringCopiesAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "factor_copies_compare.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c")
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	x := p.AND(a, b)
	root := p.AND(x, c)
	var got strings.Builder
	fmt.Fprintf(&got, "recursive %d", Aig_ManDeriveNewCone_rec(p, root))
	for _, e := range []Edge{a, b, c, x, root} {
		fmt.Fprintf(&got, " %d@%d", Aig_ObjCopy(p, e), p.node(e).TravID)
	}
	got.WriteByte('\n')
	result := Aig_ManDeriveNewCone(p, []Edge{a.Not()}, root)
	fmt.Fprintf(&got, "derive %d generation %d\n", result, p.travID)
	result = Aig_SuppMinReconstruct(p, []Edge{a}, []Edge{x, root}, []Edge{a, b, c})
	fmt.Fprintf(&got, "reconstruct %d", result)
	for _, e := range []Edge{a, b, c, x, root} {
		fmt.Fprintf(&got, " %d", Aig_ObjCopy(p, e))
	}
	got.WriteByte('\n')
	if got.String() != out {
		t.Fatalf("Go copy states:\n%sC copy states:\n%s", got.String(), out)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
