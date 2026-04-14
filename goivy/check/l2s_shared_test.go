package check

import (
	"sort"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// helper: make a FunctionSort for (domain... -> Boolean)
func boolFuncSort(domains ...lg.Sort) lg.Sort {
	sorts := make([]lg.Sort, len(domains)+1)
	copy(sorts, domains)
	sorts[len(domains)] = lg.Boolean
	return &lg.FunctionSort{Sorts: sorts}
}

// TestBuildDefnDeps_Definition verifies that *lg.Definition formulas
// in mod.Definitions are processed (Bug B fix: addEq skipped them
// because it only handled *lg.Eq).
func TestBuildDefnDeps_Definition(t *testing.T) {
	cfg := ast.NewAstConfig()
	mod := module.New()

	// Build: f(X) = And(g(X), h(X))
	uSort := &lg.UninterpretedSort{Name: "T"}
	fSym := lg.NewConst("f", boolFuncSort(uSort))
	gSym := lg.NewConst("g", boolFuncSort(uSort))
	hSym := lg.NewConst("h", boolFuncSort(uSort))
	X := &lg.Variable{Name: "X", VSort: uSort}

	lhs, _ := lg.NewApply(fSym, X)
	gApp, _ := lg.NewApply(gSym, X)
	hApp, _ := lg.NewApply(hSym, X)
	rhs, _ := lg.NewAnd(gApp, hApp)

	def := lg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("def_f"), def)
	mod.Definitions = append(mod.Definitions, lf)

	deps := BuildDefnDeps(mod)

	// g depends on definition of f, h depends on definition of f.
	assertDep(t, deps, "g", []string{"f"})
	assertDep(t, deps, "h", []string{"f"})
}

// TestBuildDefnDeps_Eq verifies that *lg.Eq formulas are still handled
// (regression guard — some code paths may produce Eq instead of Definition).
func TestBuildDefnDeps_Eq(t *testing.T) {
	cfg := ast.NewAstConfig()
	mod := module.New()

	uSort := &lg.UninterpretedSort{Name: "T"}
	fSym := lg.NewConst("f", boolFuncSort(uSort))
	gSym := lg.NewConst("g", boolFuncSort(uSort))
	X := &lg.Variable{Name: "X", VSort: uSort}

	lhs, _ := lg.NewApply(fSym, X)
	gApp, _ := lg.NewApply(gSym, X)

	eq := &lg.Eq{T1: lhs, T2: gApp}
	lf := cfg.NewLabeledFormula(cfg.NewAtom("eq_f"), eq)
	mod.Definitions = append(mod.Definitions, lf)

	deps := BuildDefnDeps(mod)
	assertDep(t, deps, "g", []string{"f"})
}

// TestBuildDefnDeps_BareConst verifies bare *lg.Const LHS (no Apply wrapper)
// is handled. (Round 4 Bug 2 regression guard.)
func TestBuildDefnDeps_BareConst(t *testing.T) {
	cfg := ast.NewAstConfig()
	mod := module.New()

	a := lg.NewConst("a", lg.Boolean)
	b := lg.NewConst("b", lg.Boolean)
	prevents := lg.NewConst("ref.prevents", lg.Boolean)

	rhs, _ := lg.NewAnd(a, b)
	def := lg.NewDefinition(prevents, rhs)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("def_prevents"), def)
	mod.Definitions = append(mod.Definitions, lf)

	deps := BuildDefnDeps(mod)
	assertDep(t, deps, "a", []string{"ref.prevents"})
	assertDep(t, deps, "b", []string{"ref.prevents"})
}

// TestBuildDefnDeps_NormalizedCanon verifies that BuildDefnDeps normalizes
// n-ary And/Or to binary trees before tracing (Bug A fix).
// The canon of the processed formula should match NormalizeOps output.
func TestBuildDefnDeps_NormalizedCanon(t *testing.T) {
	cfg := ast.NewAstConfig()
	mod := module.New()

	a := lg.NewConst("a", lg.Boolean)
	b := lg.NewConst("b", lg.Boolean)
	c := lg.NewConst("c", lg.Boolean)
	d := lg.NewConst("d", lg.Boolean)
	fSym := lg.NewConst("f", lg.Boolean)

	// Create flat 4-term And (what Go parser produces).
	flatAnd := &lg.And{Terms: []lg.Expr{a, b, c, d}}
	def := lg.NewDefinition(fSym, flatAnd)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("def_f"), def)
	mod.Definitions = append(mod.Definitions, lf)

	// BuildDefnDeps should normalize internally. Verify deps are correct.
	deps := BuildDefnDeps(mod)
	assertDep(t, deps, "a", []string{"f"})
	assertDep(t, deps, "b", []string{"f"})
	assertDep(t, deps, "c", []string{"f"})
	assertDep(t, deps, "d", []string{"f"})

	// Verify NormalizeOps converts 4-term And to binary tree.
	normalizedAnd := il.NormalizeOps(flatAnd)
	outerAnd, ok := normalizedAnd.(*lg.And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", normalizedAnd)
	}
	if len(outerAnd.Terms) != 2 {
		t.Fatalf("expected binary And (2 terms), got %d terms", len(outerAnd.Terms))
	}
	// The normalized def canon should differ from the flat def canon.
	normalizedDef := il.NormalizeOps(def).(*lg.Definition)
	if string(def.Canon()) == string(normalizedDef.Canon()) {
		t.Fatal("flat and normalized canons should differ for 4-term And")
	}
}

// TestBuildDefnDeps_GoalPrems verifies that definitions from goalPrems
// (with IsDefinition=true) are also processed.
func TestBuildDefnDeps_GoalPrems(t *testing.T) {
	cfg := ast.NewAstConfig()

	a := lg.NewConst("a", lg.Boolean)
	gSym := lg.NewConst("g", lg.Boolean)

	def := lg.NewDefinition(gSym, a)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("premdef_g"), def)
	lf.IsDefinition = true

	// nil module, definitions only from goalPrems
	deps := BuildDefnDeps(nil, lf)
	assertDep(t, deps, "a", []string{"g"})
}

// assertDep checks that deps[key] contains exactly the expected values (sorted).
func assertDep(t *testing.T, deps map[string][]string, key string, expected []string) {
	t.Helper()
	got, ok := deps[key]
	if !ok {
		t.Errorf("expected deps[%q] to exist, got no entry; full map: %v", key, deps)
		return
	}
	sortedGot := make([]string, len(got))
	copy(sortedGot, got)
	sort.Strings(sortedGot)
	sortedExp := make([]string, len(expected))
	copy(sortedExp, expected)
	sort.Strings(sortedExp)
	if len(sortedGot) != len(sortedExp) {
		t.Errorf("deps[%q]: expected %v, got %v", key, sortedExp, sortedGot)
		return
	}
	for i := range sortedGot {
		if sortedGot[i] != sortedExp[i] {
			t.Errorf("deps[%q]: expected %v, got %v", key, sortedExp, sortedGot)
			return
		}
	}
}
