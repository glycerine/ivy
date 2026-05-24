package goivy

import (
	"sort"
	"strings"
	"testing"
)

// helper: make a FunctionSort for (domain... -> Boolean)
func boolFuncSort(domains ...Sort) Sort {
	sorts := make([]Sort, len(domains)+1)
	copy(sorts, domains)
	sorts[len(domains)] = Boolean
	return &LogicFunctionSort{Sorts: sorts}
}

// TestBuildDefnDeps_Definition verifies that *lg.Definition formulas
// in mod.Definitions are processed (Bug B fix: addEq skipped them
// because it only handled *lg.Eq).
func TestBuildDefnDeps_Definition(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()

	// Build: f(X) = LogicAnd(g(X), h(X))
	uSort := &UninterpretedSort{Name: "T"}
	fSym := NewConst("f", boolFuncSort(uSort))
	gSym := NewConst("g", boolFuncSort(uSort))
	hSym := NewConst("h", boolFuncSort(uSort))
	X := &LogicVariable{Name: "X", VSort: uSort}

	lhs, _ := NewApply(fSym, X)
	gApp, _ := NewApply(gSym, X)
	hApp, _ := NewApply(hSym, X)
	rhs, _ := NewAnd(gApp, hApp)

	def := NewDefinition(lhs, rhs)
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
	cfg := NewAstConfig()
	mod := New()

	uSort := &UninterpretedSort{Name: "T"}
	fSym := NewConst("f", boolFuncSort(uSort))
	gSym := NewConst("g", boolFuncSort(uSort))
	X := &LogicVariable{Name: "X", VSort: uSort}

	lhs, _ := NewApply(fSym, X)
	gApp, _ := NewApply(gSym, X)

	eq := &Eq{T1: lhs, T2: gApp}
	lf := cfg.NewLabeledFormula(cfg.NewAtom("eq_f"), eq)
	mod.Definitions = append(mod.Definitions, lf)

	deps := BuildDefnDeps(mod)
	assertDep(t, deps, "g", []string{"f"})
}

// TestBuildDefnDeps_BareConst verifies bare *lg.Const LHS (no Apply wrapper)
// is handled. (Round 4 Bug 2 regression guard.)
func TestBuildDefnDeps_BareConst(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()

	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	prevents := NewConst("ref.prevents", Boolean)

	rhs, _ := NewAnd(a, b)
	def := NewDefinition(prevents, rhs)
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
	cfg := NewAstConfig()
	mod := New()

	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	c := NewConst("c", Boolean)
	d := NewConst("d", Boolean)
	fSym := NewConst("f", Boolean)

	// Create flat 4-term And (what Go parser produces).
	flatAnd := &LogicAnd{Terms: []Expr{a, b, c, d}}
	def := NewDefinition(fSym, flatAnd)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("def_f"), def)
	mod.Definitions = append(mod.Definitions, lf)

	// BuildDefnDeps should normalize internally. Verify deps are correct.
	deps := BuildDefnDeps(mod)
	assertDep(t, deps, "a", []string{"f"})
	assertDep(t, deps, "b", []string{"f"})
	assertDep(t, deps, "c", []string{"f"})
	assertDep(t, deps, "d", []string{"f"})

	// Verify NormalizeOps converts 4-term And to binary tree.
	normalizedAnd := NormalizeOps(flatAnd)
	outerAnd, ok := normalizedAnd.(*LogicAnd)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", normalizedAnd)
	}
	if len(outerAnd.Terms) != 2 {
		t.Fatalf("expected binary And (2 terms), got %d terms", len(outerAnd.Terms))
	}
	// The normalized def canon should differ from the flat def canon.
	normalizedDef := NormalizeOps(def).(*LogicDefinition)
	if string(def.Canon()) == string(normalizedDef.Canon()) {
		t.Fatal("flat and normalized canons should differ for 4-term And")
	}
}

func TestSharedStep3FullKeepsModifiedLocals(t *testing.T) {
	ticket := &UninterpretedSort{Name: "ticket"}
	mod := New()
	mod.Sig = NewSig()
	if err := mod.Sig.AddSort(ticket); err != nil {
		t.Fatalf("AddSort: %v", err)
	}
	if _, err := mod.Sig.AddSymbol("next_ticket", ticket); err != nil {
		t.Fatalf("AddSymbol: %v", err)
	}

	actCfg := NewActionsConfig()
	local := NewConst("loc:ntn", ticket)
	nextTicket := NewConst("next_ticket", ticket)
	zero := NewConst("zero", ticket)
	body := NewSequence(
		NewAssignAction(local, zero),
		NewAssignAction(nextTicket, local),
	)
	action := NewLocalActionOn(actCfg, "test.local.save", local, body)
	model := &NormalProgram{
		Bindings: []*ActionTermBinding{{
			Name:   "step",
			Action: &ActionTerm{Stmt: action},
		}},
	}
	cfg := &InstrumentationConfig{Mod: mod, NotLf: True}

	SharedStep3_CollectNamedBinders(cfg, model, true)

	if len(cfg.ToSave) < 2 {
		t.Fatalf("ToSave len = %d, want at least loc:ntn and next_ticket", len(cfg.ToSave))
	}
	got0, ok := cfg.ToSave[0].Body.(*Const)
	if !ok || got0.Name != "loc:ntn" {
		t.Fatalf("ToSave[0] = %s, want modified local loc:ntn before globals", cfg.ToSave[0].Body.Canon())
	}
	got1, ok := cfg.ToSave[1].Body.(*Const)
	if !ok || got1.Name != "next_ticket" {
		t.Fatalf("ToSave[1] = %s, want next_ticket after local initializer", cfg.ToSave[1].Body.Canon())
	}
}

func TestL2STraceSymbolShowsNumeralSort(t *testing.T) {
	ticket := &UninterpretedSort{Name: "ticket"}
	if got := l2sTraceSymbol(NewConst("1", ticket)); got != "1:ticket" {
		t.Fatalf("l2sTraceSymbol(numeral) = %q, want Python-style sort-qualified numeral", got)
	}
	if got := l2sTraceSymbol(NewConst("service", ticket)); got != "service" {
		t.Fatalf("l2sTraceSymbol(service) = %q, want bare non-numeral name", got)
	}
}

// TestBuildDefnDeps_GoalPrems verifies that definitions from goalPrems
// (with IsDefinition=true) are also processed.
func TestBuildDefnDeps_GoalPrems(t *testing.T) {
	cfg := NewAstConfig()

	a := NewConst("a", Boolean)
	gSym := NewConst("g", Boolean)

	def := NewDefinition(gSym, a)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("premdef_g"), def)
	lf.IsDefinition = true

	// nil module, definitions only from goalPrems
	deps := BuildDefnDeps(nil, lf)
	assertDep(t, deps, "a", []string{"g"})
}

// assertDep checks that deps contains an entry whose key Sexp contains keyName,
// and that the values' Sexps contain exactly the expected names (sorted).
func assertDep(t *testing.T, deps map[NodeKey][]NodeKey, keyName string, expectedNames []string) {
	t.Helper()
	// Find the key whose Sexp contains keyName.
	var matchKey NodeKey
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
