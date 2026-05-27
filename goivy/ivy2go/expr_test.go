package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// --- M3: expression emission tests ----------------------------------
//
// These tests build expressions in-process and call Generator.emitExpr
// directly, then assert the emitted Go source matches expectations.
// We don't need a full module for most cases — a minimal Generator
// with just enough Module wiring suffices.

// newExprGen returns a Generator suitable for emitExpr unit tests,
// with the given Ivy source compiled into its Module and every
// per-stream goWriter wired up against Ctx (so direct emitter calls
// in tests land in the same streams Generate() would use).
func newExprGen(t *testing.T, src string) *Generator {
	t.Helper()
	mod := compileIvySource(t, src)
	g := &Generator{
		Mod:                mod,
		Config:             Config{Target: "impl"},
		PackageName:        "p",
		StateTypeName:      "State",
		Ctx:                NewGoContext(),
		thunkMemo:          map[string]string{},
		exprAliases:        map[string]goivy.Expr{},
		extRel:             map[string]bool{},
		nativeOnceMemo:     map[string]bool{},
		encodedSorts:       map[string]bool{},
		importCallersCache: map[string]bool{},
	}
	g.Ctx.PackageName = g.PackageName
	g.types = newGoWriter(g.Ctx.Types)
	g.state = newGoWriter(g.Ctx.State)
	g.actions = newGoWriter(g.Ctx.Actions)
	g.init = newGoWriter(g.Ctx.Init)
	g.runtime = newGoWriter(g.Ctx.Runtime)
	g.nondet = newGoWriter(g.Ctx.Nondet)
	g.extensional = newGoWriter(g.Ctx.Extensional)
	g.definitions = newGoWriter(g.Ctx.Definitions)
	g.thunks = newGoWriter(g.Ctx.Thunks)
	g.native = newGoWriter(g.Ctx.Native)
	g.repl = newGoWriter(g.Ctx.Repl)
	g.main = newGoWriter(g.Ctx.Main)
	prepareModuleForGo(mod, g.Config)
	return g
}

func TestEmitExpr_BooleanConstants(t *testing.T) {
	g := newExprGen(t, "")
	for in, want := range map[string]string{
		"true":  "true",
		"false": "false",
	} {
		c := &goivy.Const{Name: in, CSort: goivy.Boolean}
		got, err := g.emitExpr(c)
		if err != nil {
			t.Fatalf("emitExpr(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("emitExpr(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEmitExpr_LogicNot(t *testing.T) {
	g := newExprGen(t, "")
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, err := g.emitExpr(&goivy.LogicNot{Body: body})
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "!(true)" {
		t.Errorf("LogicNot = %q, want %q", got, "!(true)")
	}
}

func TestEmitExpr_LogicAndOr(t *testing.T) {
	g := newExprGen(t, "")
	t1 := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	t2 := &goivy.Const{Name: "false", CSort: goivy.Boolean}

	got, err := g.emitExpr(&goivy.LogicAnd{Terms: []goivy.Expr{t1, t2}})
	if err != nil {
		t.Fatalf("emitExpr And: %v", err)
	}
	if got != "(true) && (false)" {
		t.Errorf("And = %q, want %q", got, "(true) && (false)")
	}

	got, err = g.emitExpr(&goivy.LogicOr{Terms: []goivy.Expr{t1, t2}})
	if err != nil {
		t.Fatalf("emitExpr Or: %v", err)
	}
	if got != "(true) || (false)" {
		t.Errorf("Or = %q, want %q", got, "(true) || (false)")
	}

	// Empty And / Or use the identity.
	got, err = g.emitExpr(&goivy.LogicAnd{Terms: nil})
	if err != nil {
		t.Fatalf("empty And: %v", err)
	}
	if got != "true" {
		t.Errorf("empty And = %q, want true", got)
	}
	got, err = g.emitExpr(&goivy.LogicOr{Terms: nil})
	if err != nil {
		t.Fatalf("empty Or: %v", err)
	}
	if got != "false" {
		t.Errorf("empty Or = %q, want false", got)
	}
}

func TestEmitExpr_LogicImpliesIff(t *testing.T) {
	g := newExprGen(t, "")
	t1 := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	t2 := &goivy.Const{Name: "false", CSort: goivy.Boolean}

	got, err := g.emitExpr(&goivy.LogicImplies{T1: t1, T2: t2})
	if err != nil {
		t.Fatalf("Implies: %v", err)
	}
	if got != "(!(true) || (false))" {
		t.Errorf("Implies = %q", got)
	}

	got, err = g.emitExpr(&goivy.LogicIff{T1: t1, T2: t2})
	if err != nil {
		t.Fatalf("Iff: %v", err)
	}
	if got != "((true) == (false))" {
		t.Errorf("Iff = %q", got)
	}
}

func TestEmitExpr_Eq(t *testing.T) {
	g := newExprGen(t, "")
	t1 := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	t2 := &goivy.Const{Name: "false", CSort: goivy.Boolean}
	got, err := g.emitExpr(&goivy.Eq{T1: t1, T2: t2})
	if err != nil {
		t.Fatalf("Eq: %v", err)
	}
	if got != "(true == false)" {
		t.Errorf("Eq = %q, want %q", got, "(true == false)")
	}
}

func TestEmitExpr_IteUsesHelper(t *testing.T) {
	g := newExprGen(t, "")
	c := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	thn := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	els := &goivy.Const{Name: "false", CSort: goivy.Boolean}
	ite, err := goivy.NewIte(c, thn, els)
	if err != nil {
		t.Fatalf("NewIte: %v", err)
	}
	got, err := g.emitExpr(ite)
	if err != nil {
		t.Fatalf("Ite: %v", err)
	}
	if !strings.HasPrefix(got, "ite_bool(") {
		t.Errorf("Ite emission %q does not start with ite_bool(", got)
	}
}

func TestEmitExpr_NumericConst(t *testing.T) {
	g := newExprGen(t, "")
	// IsNumeral fires automatically when the name starts with a digit.
	c := &goivy.Const{Name: "42", CSort: goivy.Boolean}
	got, err := g.emitExpr(c)
	if err != nil {
		t.Fatalf("numeral: %v", err)
	}
	if got != "42" {
		t.Errorf("numeral = %q, want 42", got)
	}
}

func TestEmitExpr_CastToRangeClampIsTypedExpression(t *testing.T) {
	g := newExprGen(t, "")
	src := &goivy.UninterpretedSort{Name: "int"}
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "5"}, Ub: goivy.NumeralBound{Value: "7"}}
	castSort, err := goivy.NewFunctionSort(src, rng)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	app := goivy.MustApply(goivy.NewConst("cast", castSort), goivy.NewConst("v", src))
	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"func() Idx", "x := Idx(v)", "lo, hi := Idx(5), Idx(7)", "return x"} {
		if !strings.Contains(got, want) {
			t.Fatalf("range cast clamp missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "return if") || strings.Contains(got, "return if ") {
		t.Fatalf("range cast clamp should be expression-shaped Go, got:\n%s", got)
	}
	assertGoSourceValid(t, "range_cast.go", "package p\ntype Idx int\nvar v int\nvar _ Idx = "+got+"\n")
}

func TestEmitExpr_RangeArithmeticClampIsTypedExpression(t *testing.T) {
	g := newExprGen(t, "")
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "5"}, Ub: goivy.NumeralBound{Value: "7"}}
	fnSort, err := goivy.NewFunctionSort(rng, rng, rng)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	app := goivy.MustApply(goivy.NewConst("+", fnSort), goivy.NewConst("a", rng), goivy.NewConst("b", rng))
	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"func() Idx", "x := (a) + (b)", "lo, hi := Idx(5), Idx(7)", "return x"} {
		if !strings.Contains(got, want) {
			t.Fatalf("range arithmetic clamp missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "return if") {
		t.Fatalf("range arithmetic clamp should not emit statement text as expression:\n%s", got)
	}
	assertGoSourceValid(t, "range_arith.go", "package p\ntype Idx int\nvar a, b Idx\nvar _ Idx = "+got+"\n")
}

func TestEmitExpr_RangeNumeralClampIsTypedExpression(t *testing.T) {
	g := newExprGen(t, "")
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "5"}, Ub: goivy.NumeralBound{Value: "7"}}
	got, err := g.emitExpr(goivy.NewConst("10", rng))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"func() Idx", "x := Idx(10)", "lo, hi := Idx(5), Idx(7)", "return hi"} {
		if !strings.Contains(got, want) {
			t.Fatalf("range numeral clamp missing %q:\n%s", want, got)
		}
	}
	assertGoSourceValid(t, "range_numeral.go", "package p\ntype Idx int\nvar _ Idx = "+got+"\n")
}

func TestEmitExpr_EnumConstantUsesPascalCase(t *testing.T) {
	g := newExprGen(t, "type color = {red, green, blue}")
	// `red` should resolve as enum constant Red, not as ident red.
	colorSort, _ := g.Mod.Sig.Sorts.Get2("color")
	c := &goivy.Const{Name: "red", CSort: colorSort}
	got, err := g.emitExpr(c)
	if err != nil {
		t.Fatalf("enum: %v", err)
	}
	if got != "Red" {
		t.Errorf("enum const = %q, want Red", got)
	}
}

func TestEmitExpr_VariableUsesGoIdent(t *testing.T) {
	g := newExprGen(t, "")
	v := &goivy.LogicVariable{Name: "type"} // collides with Go keyword
	got, err := g.emitExpr(v)
	if err != nil {
		t.Fatalf("var: %v", err)
	}
	if got != "type_" {
		t.Errorf("reserved-name variable = %q, want type_", got)
	}
}

func TestEmitExpr_LogicNamedBinderRejected(t *testing.T) {
	g := newExprGen(t, "")
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	_, err := g.emitExpr(&goivy.LogicNamedBinder{Name: "test", Variables: nil, Body: body})
	if err == nil {
		t.Error("expected error for LogicNamedBinder, got nil")
	}
}

// --- M3: isInfix / iteHelperSuffix unit tests ------------------------

func TestIsInfix(t *testing.T) {
	for _, op := range []string{"+", "-", "*", "/", "%", "<", "<=", ">", ">="} {
		if !isInfix(op) {
			t.Errorf("isInfix(%q) = false, want true", op)
		}
	}
	for _, op := range []string{"foo", "bvadd", "and"} {
		if isInfix(op) {
			t.Errorf("isInfix(%q) = true, want false", op)
		}
	}
}

func TestIteHelperSuffix(t *testing.T) {
	cases := map[string]string{
		"uint32":       "uint32",
		"bool":         "bool",
		"[3]uint32":    "arr3_uint32",
		"map[int]bool": "maparrint_bool", // simple round-trip; loses fidelity for compound types
	}
	for in, want := range cases {
		if got := iteHelperSuffix(in); got != want {
			t.Errorf("iteHelperSuffix(%q) = %q, want %q", in, got, want)
		}
	}
}

// --- M3: bv_expr basic tests ----------------------------------------

func TestBVMaskBitsBands(t *testing.T) {
	cases := map[int]string{
		1:   "1",
		8:   "255",
		32:  "4294967295",
		64:  "0xFFFFFFFFFFFFFFFF",
		128: "Uint128Mask(128)",
		256: "bigIntMask(256)",
	}
	for bits, want := range cases {
		if got := bvMask(bits); got != want {
			t.Errorf("bvMask(%d) = %q, want %q", bits, got, want)
		}
	}
}

func TestEmitBVNumeral_Narrow(t *testing.T) {
	g := newExprGen(t, `
type word
interpret word -> bv[8]
`)
	wordSort, _ := g.Mod.Sig.Sorts.Get2("word")
	c := &goivy.Const{Name: "5", CSort: wordSort} // digit-prefix → IsNumeral
	got, ok := g.emitBVNumeral(c)
	if !ok {
		t.Fatal("emitBVNumeral !ok")
	}
	if got != "(uint32(5) & 255)" {
		t.Errorf("emitBVNumeral = %q", got)
	}
}

func TestEmitBVNumeral_WideRequestsUint128(t *testing.T) {
	g := newExprGen(t, `
type wide
interpret wide -> bv[128]
`)
	wideSort, _ := g.Mod.Sig.Sorts.Get2("wide")
	c := &goivy.Const{Name: "65", CSort: wideSort}
	_, ok := g.emitBVNumeral(c)
	if !ok {
		t.Fatal("emitBVNumeral wide !ok")
	}
	if !g.Ctx.OnceGlobals["__need_uint128"] {
		t.Error("wide BV numeral should trigger Uint128 runtime emission")
	}
}

func TestEmitBVApply_BvAdd(t *testing.T) {
	g := newExprGen(t, `
type word
interpret word -> bv[8]
`)
	wordSort, _ := g.Mod.Sig.Sorts.Get2("word")
	// Build `bvadd : (word, word) -> word` as a FunctionSort so
	// NewApply infers the result sort correctly.
	bvaddSort, err := goivy.NewFunctionSort(wordSort, wordSort, wordSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	a := &goivy.Const{Name: "x", CSort: wordSort}
	b := &goivy.Const{Name: "y", CSort: wordSort}
	fn := goivy.NewConst("bvadd", bvaddSort)
	apply, err := goivy.NewApply(fn, a, b)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	got, handled, err := g.emitBVApply("bvadd", apply)
	if err != nil {
		t.Fatalf("emitBVApply: %v", err)
	}
	if !handled {
		t.Fatal("bvadd not handled")
	}
	if !strings.Contains(got, "+") || !strings.Contains(got, "& 255") {
		t.Errorf("bvadd emission missing + or & 255: %q", got)
	}
}

func TestParseBFEParams(t *testing.T) {
	if lo, hi, ok := parseBFEParams("bfe[0][7]"); !ok || lo != 0 || hi != 7 {
		t.Errorf("parseBFEParams = (%d, %d, %v)", lo, hi, ok)
	}
	if lo, hi, ok := parseBFEParams("bfe[3:5]"); !ok || lo != 3 || hi != 5 {
		t.Errorf("parseBFEParams[3:5] = (%d, %d, %v)", lo, hi, ok)
	}
	if _, _, ok := parseBFEParams("bvadd"); ok {
		t.Error("parseBFEParams(bvadd) should be !ok")
	}
}

// --- M3: runtime helper emission test --------------------------------

func TestRuntimeHelpers_AlwaysOn(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["runtime.go"]
	for _, want := range []string{
		"func ivyAssert",
		"func ivyAssume",
		"func ivyChoose",
		"var ivyRand",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("runtime.go missing %q:\n%s", want, text)
		}
	}
}

func TestRuntimeHelpers_Uint128EmittedOnDemand(t *testing.T) {
	// Build a module that triggers emitBVNumeral on a 128-bit BV
	// during action/definitions emission; for now we invoke
	// requireUint128 directly to validate runtime emission.
	mod := compileIvySource(t, "")
	g := &Generator{
		Mod:           mod,
		Config:        Config{Target: "impl"},
		PackageName:   "p",
		StateTypeName: "State",
		Ctx:           NewGoContext(),
	}
	g.Ctx.PackageName = "p"
	g.runtime = newGoWriter(g.Ctx.Runtime)
	g.requireUint128()
	g.emitRuntimeHelpers(&g.runtime)
	g.emitRuntimeHelpersLate(&g.runtime)
	text := g.Ctx.Runtime.GetFile()
	for _, want := range []string{"type Uint128 struct", "func Uint128Mask", "func Uint128FromString"} {
		if !strings.Contains(text, want) {
			t.Errorf("runtime missing %q:\n%s", want, text)
		}
	}
}

func TestRuntimeFileIsGofmtClean(t *testing.T) {
	// End-to-end: empty module emission includes runtime.go and it
	// must still be gofmt-clean.
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertGoSourceGofmt(t, "runtime.go", out.Files["runtime.go"])
}
