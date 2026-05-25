package ivy2go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// --- OPEN 054: wide BV (>64) operator emission tests -----------------

func TestEmitBV_Wide96_RoutesThroughBigInt(t *testing.T) {
	g := newExprGen(t, `
type word
interpret word -> bv[96]
`)
	wordSort, _ := g.Mod.Sig.Sorts.Get2("word")
	bvaddSort, _ := goivy.NewFunctionSort(wordSort, wordSort, wordSort)
	fn := goivy.NewConst("bvadd", bvaddSort)
	x := &goivy.Const{Name: "x", CSort: wordSort}
	y := &goivy.Const{Name: "y", CSort: wordSort}
	apply, err := goivy.NewApply(fn, x, y)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	got, handled, err := g.emitBVApply("bvadd", apply)
	if err != nil {
		t.Fatalf("emitBVApply: %v", err)
	}
	if !handled {
		t.Fatal("bvadd not handled for wide BV")
	}
	if !strings.Contains(got, "wideBVAdd(") {
		t.Errorf("wide bvadd should call wideBVAdd, got %q", got)
	}
	if !strings.Contains(got, ", 96)") {
		t.Errorf("wide bvadd should pass bit width 96, got %q", got)
	}
}

func TestEmitBV_Wide256_RequestsBigIntRuntime(t *testing.T) {
	g := newExprGen(t, `
type huge
interpret huge -> bv[256]
`)
	hugeSort, _ := g.Mod.Sig.Sorts.Get2("huge")
	bvxorSort, _ := goivy.NewFunctionSort(hugeSort, hugeSort, hugeSort)
	fn := goivy.NewConst("bvxor", bvxorSort)
	x := &goivy.Const{Name: "x", CSort: hugeSort}
	y := &goivy.Const{Name: "y", CSort: hugeSort}
	apply, _ := goivy.NewApply(fn, x, y)
	_, handled, err := g.emitBVApply("bvxor", apply)
	if err != nil || !handled {
		t.Fatalf("bvxor on bv[256] should be handled, got err=%v handled=%v", err, handled)
	}
	if !g.Ctx.OnceGlobals["__need_bigint"] {
		t.Error("wide BV op should request bigint runtime")
	}
}

func TestEmitBV_Wide_BvShiftLowersThroughBigInt(t *testing.T) {
	g := newExprGen(t, `
type w128
interpret w128 -> bv[128]
`)
	wSort, _ := g.Mod.Sig.Sorts.Get2("w128")
	bvshlSort, _ := goivy.NewFunctionSort(wSort, wSort, wSort)
	fn := goivy.NewConst("bvshl", bvshlSort)
	x := &goivy.Const{Name: "x", CSort: wSort}
	n := &goivy.Const{Name: "n", CSort: wSort}
	apply, _ := goivy.NewApply(fn, x, n)
	got, handled, err := g.emitBVApply("bvshl", apply)
	if err != nil || !handled {
		t.Fatalf("bvshl on bv[128] should be handled, got err=%v handled=%v", err, handled)
	}
	if !strings.Contains(got, "wideBVShl(") {
		t.Errorf("wide bvshl should call wideBVShl, got %q", got)
	}
}

// Tier 2 smoke: build a wide-BV-using package.
func TestSmoke_BuildEmittedWideBV(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type word
interpret word -> bv[96]
action add(x: word, y: word) returns (r: word) = {
	r := x + y
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := t.TempDir()
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"),
		[]byte("module ivygo_widebv_smoke\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = pkgDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}
