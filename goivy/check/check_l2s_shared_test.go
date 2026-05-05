package check

import (
	"sort"
	"strings"
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

// assertDep checks that deps contains an entry whose key Sexp contains keyName,
// and that the values' Sexps contain exactly the expected names (sorted).
func assertDep(t *testing.T, deps map[lg.NodeKey][]lg.NodeKey, keyName string, expectedNames []string) {
	t.Helper()
	// Find the key whose Sexp contains keyName.
	var matchKey lg.NodeKey
	var found bool
	for k := range deps {
		if strings.Contains(string(k), keyName) {
			matchKey = k
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected deps entry containing %q, got no match; keys: %v", keyName, func() []string {
			var ks []string
			for k := range deps {
				ks = append(ks, string(k))
			}
			return ks
		}())
		return
	}
	got := deps[matchKey]
	gotNames := make([]string, len(got))
	for i, v := range got {
		// Extract the name from the NodeKey Sexp.
		for _, en := range expectedNames {
			if strings.Contains(string(v), en) {
				gotNames[i] = en
				break
			}
		}
		if gotNames[i] == "" {
			gotNames[i] = string(v)
		}
	}
	sort.Strings(gotNames)
	sortedExp := make([]string, len(expectedNames))
	copy(sortedExp, expectedNames)
	sort.Strings(sortedExp)
	if len(gotNames) != len(sortedExp) {
		t.Errorf("deps[%q]: expected %v, got %v", keyName, sortedExp, gotNames)
		return
	}
	for i := range gotNames {
		if gotNames[i] != sortedExp[i] {
			t.Errorf("deps[%q]: expected %v, got %v", keyName, sortedExp, gotNames)
			return
		}
	}
}
