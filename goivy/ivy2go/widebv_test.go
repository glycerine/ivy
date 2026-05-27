package ivy2go

import (
	"fmt"
	"os/exec"
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

func TestEmitBVNumeral_WideWidthsUseGoTypeBigInt(t *testing.T) {
	for _, tc := range []struct {
		name string
		bits int
	}{
		{"word96", 96},
		{"word128", 128},
		{"word256", 256},
	} {
		g := newExprGen(t, fmt.Sprintf("type %s\ninterpret %s -> bv[%d]", tc.name, tc.name, tc.bits))
		sort, _ := g.Mod.Sig.Sorts.Get2(tc.name)
		got, ok := g.emitBVNumeral(goivy.NewConst("300", sort))
		if !ok {
			t.Fatalf("emitBVNumeral(bv[%d]) = !ok", tc.bits)
		}
		for _, want := range []string{"func() *big.Int", `SetString("300", 0)`, fmt.Sprintf("bigIntMask(%d)", tc.bits)} {
			if !strings.Contains(got, want) {
				t.Fatalf("bv[%d] numeral missing %q:\n%s", tc.bits, want, got)
			}
		}
	}
}

func TestEmitActions_WideBVLiteralAssignmentUsesBigIntExpression(t *testing.T) {
	mod := compileIvySource(t, `
type word
interpret word -> bv[96]
individual x : word
action set = {
	x := 300
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	requireHasLineWithAllTerms(t, actions, "s.X", "=", "func() *big.Int")
}

func TestEmitInit_SolvedWideBVInitialStateUsesBigIntExpression(t *testing.T) {
	mod := wideBVInitCondMod(t, 96, "300")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	init := out.Files["init.go"]
	requireHasLineWithAllTerms(t, init, "s.X", "=", "func() *big.Int")
	if !strings.Contains(init, `SetString("300", 0)`) {
		t.Fatalf("solver-derived wide BV initial value should preserve the model value:\n%s", init)
	}
}

func TestParseBVModelValueTextFormats(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
		ok   bool
	}{
		{"300", "300", true},
		{"|300|", "300", true},
		{"#x00000000000000000000012c", "300", true},
		{"|#x0000012c|", "300", true},
		{"#b100101100", "300", true},
		{"(_ bv300 96)", "300", true},
		{"#xzz", "", false},
		{"(_ bv300)", "", false},
		{"-1", "", false},
	} {
		got, ok := parseBVModelValueText(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("parseBVModelValueText(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestEmitInit_SolvedSmallBVInitialStateUsesPrimitiveTypes(t *testing.T) {
	for _, tc := range []struct {
		bits  int
		value string
		want  string
	}{
		{8, "5", "uint32(5)"},
		{32, "300", "uint32(300)"},
		{64, "300", "uint64(300)"},
	} {
		mod := wideBVInitCondMod(t, tc.bits, tc.value)
		out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
		if err != nil {
			t.Fatalf("Generate bv[%d]: %v", tc.bits, err)
		}
		init := out.Files["init.go"]
		requireHasLineWithAllTerms(t, init, "s.X", "=", tc.want)
		if strings.Contains(init, "#x") || strings.Contains(init, "#b") {
			t.Fatalf("solver-derived small BV initial value leaked raw Z3 syntax:\n%s", init)
		}
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
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = pkgDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}

func TestSmoke_BuildEmittedWideBVLiteralAndInitialState(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := wideBVInitCondMod(t, 96, "300")
	set := goivy.NewSequence(goivy.NewAssignAction(goivy.NewConst("x", mustSort(t, mod, "word")), goivy.NewConst("511", mustSort(t, mod, "word"))))
	mod.Actions.Set("set", set)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = pkgDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}

func wideBVInitCondMod(t *testing.T, bits int, value string) *goivy.Module {
	t.Helper()
	mod := compileIvySource(t, fmt.Sprintf(`
type word
interpret word -> bv[%d]
individual x : word
`, bits))
	word := mustSort(t, mod, "word")
	eq, err := goivy.NewEq(goivy.NewConst("x", word), goivy.NewConst(value, word))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(eq, nil)
	return mod
}

func mustSort(t *testing.T, mod *goivy.Module, name string) goivy.Sort {
	t.Helper()
	sort, ok := mod.Sig.Sorts.Get2(name)
	if !ok {
		t.Fatalf("missing sort %q", name)
	}
	return sort
}
