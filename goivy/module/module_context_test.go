package module

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
)

func TestEnterExit(t *testing.T) {
	m1 := New()
	cfg := m1.Cfg

	m1.Enter()
	if cfg.CurrentModule != m1 {
		t.Fatal("expected m1 to be current after Enter")
	}

	m2 := New()
	m2.Cfg = cfg
	m2.Enter()
	if cfg.CurrentModule != m2 {
		t.Fatal("expected m2 to be current after nested Enter")
	}

	m2.Exit()
	if cfg.CurrentModule != m1 {
		t.Fatal("expected m1 to be restored after m2.Exit")
	}

	m1.Exit()
	if cfg.CurrentModule != nil {
		t.Fatal("expected nil after m1.Exit")
	}
}

func TestRelevantDefinitionsEmpty(t *testing.T) {
	m := New()
	syms := map[string]bool{"x": true}
	result := RelevantDefinitions(m, syms)
	if len(result) != 0 {
		t.Errorf("expected 0 relevant definitions, got %d", len(result))
	}
}

func TestRelevantDefinitionsReachable(t *testing.T) {
	m := New()

	tSort := &lg.UninterpretedSort{Name: "t"}
	fSort, _ := lg.NewFunctionSort(tSort, tSort)
	gSort, _ := lg.NewFunctionSort(tSort, tSort)

	fSym := lg.NewConst("f", fSort)
	gSym := lg.NewConst("g", gSort)
	x, _ := lg.NewVariable("X", tSort)

	// f(X) = g(X)
	fApp, _ := lg.NewApply(fSym, x)
	gApp, _ := lg.NewApply(gSym, x)
	def1 := il.NewDefinition(fApp, gApp)

	// g(X) = X
	gApp2, _ := lg.NewApply(gSym, x)
	def2 := il.NewDefinition(gApp2, x)

	m.Definitions = []*ast.LabeledFormula{
		{Formula: def1},
		{Formula: def2},
	}

	// Starting from "f", should reach both "f" and "g".
	syms := map[string]bool{"f": true}
	result := RelevantDefinitions(m, syms)
	if len(result) != 2 {
		t.Errorf("expected 2 relevant definitions (f and g), got %d", len(result))
	}

	// Starting from "g", should only reach "g".
	syms2 := map[string]bool{"g": true}
	result2 := RelevantDefinitions(m, syms2)
	if len(result2) != 1 {
		t.Errorf("expected 1 relevant definition (g only), got %d", len(result2))
	}
}

func TestRelevantDefinitionsUnreachable(t *testing.T) {
	m := New()
	tSort := &lg.UninterpretedSort{Name: "t"}
	fSort, _ := lg.NewFunctionSort(tSort, tSort)
	fSym := lg.NewConst("f", fSort)
	x, _ := lg.NewVariable("X", tSort)
	fApp, _ := lg.NewApply(fSym, x)
	def := il.NewDefinition(fApp, x)

	m.Definitions = []*ast.LabeledFormula{
		{Formula: def},
	}

	// Starting from "z" (not related to "f").
	syms := map[string]bool{"z": true}
	result := RelevantDefinitions(m, syms)
	if len(result) != 0 {
		t.Errorf("expected 0 relevant definitions for unreachable sym, got %d", len(result))
	}
}

func TestSortDependencyGraph(t *testing.T) {
	m := New()
	tSort := &lg.UninterpretedSort{Name: "t"}
	uSort := &lg.UninterpretedSort{Name: "u"}
	dSort, _ := lg.NewFunctionSort(tSort, uSort)
	destr := lg.NewConst("d", dSort)
	m.SortDestructors["t"] = []*lg.Const{destr}
	m.SortOrder = []string{"t", "u"}

	graph := m.SortDependencyGraph()
	deps, ok := graph["t"]
	if !ok {
		t.Fatal("expected sort 't' in dependency graph")
	}
	if len(deps) != 1 || deps[0] != "u" {
		t.Errorf("expected deps [u], got %v", deps)
	}
	// "u" should have no deps (no destructors).
	if _, ok := graph["u"]; ok {
		t.Error("'u' should have no dependencies")
	}
}

func TestPrevModuleField(t *testing.T) {
	m := New()
	if m.prevModule != nil {
		t.Error("prevModule should be nil on new module")
	}
}

// Suppress unused import warnings.
var _ = il.NewSig
