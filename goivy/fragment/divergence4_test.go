package fragment

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// TestMacroMapKeyIncludesSort verifies that macroMap uses structural keys
// (lg.NodeKey) that include the sort, matching Python's recstruct equality.
// Two macros defining the same name "f" but on different sorts must be
// stored as separate entries (Divergence 4 fix).
func TestMacroMapKeyIncludesSort(t *testing.T) {
	S := &lg.UninterpretedSort{Name: "S"}
	T := &lg.UninterpretedSort{Name: "T"}

	// Create two function sorts: S->Bool and T->Bool
	fSortS, err := lg.NewFunctionSort(S, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	fSortT, err := lg.NewFunctionSort(T, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}

	// Two symbols with same name "f" but different sorts
	fS := lg.NewConst("f", fSortS)
	fT := lg.NewConst("f", fSortT)

	// Params for each definition
	xS, err := lg.NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	xT, err := lg.NewVariable("X", T)
	if err != nil {
		t.Fatal(err)
	}

	// LHS: f(X) for each sort
	lhsS, err := lg.NewApply(fS, xS)
	if err != nil {
		t.Fatal(err)
	}
	lhsT, err := lg.NewApply(fT, xT)
	if err != nil {
		t.Fatal(err)
	}

	// RHS: just use the variable itself (simplest possible body)
	defS := lg.NewDefinition(lhsS, xS)
	defT := lg.NewDefinition(lhsT, xT)

	// Create labeled formulas
	acfg := ast.NewAstConfig()
	lfS := acfg.NewLabeledFormula(nil, nil)
	lfT := acfg.NewLabeledFormula(nil, nil)

	// Build fmlaPairs
	macros := []fmlaPair{
		{fmla: defS, source: lfS, lineno: 1},
		{fmla: defT, source: lfT, lineno: 2},
	}

	sig := il.NewSig()
	sig.AddSort(S)
	sig.AddSort(T)
	c := newChecker(sig, nil)

	c.createMacroMaps(nil, nil, macros)

	if len(c.macroMap) != 2 {
		t.Errorf("macroMap has %d entries, want 2 (same-name different-sort macros should be distinct keys)", len(c.macroMap))
		for k, v := range c.macroMap {
			t.Logf("  key=%s def.Lhs=%s", k, v.def.Lhs.Sexp())
		}
	}
}

// TestMacroMapKeyMatchesPythonEquality verifies that two Const symbols
// with the same name AND same sort produce the same lg.Key(), confirming
// structural equality matches Python's recstruct __eq__.
func TestMacroMapKeyMatchesPythonEquality(t *testing.T) {
	S := &lg.UninterpretedSort{Name: "S"}
	fSort, err := lg.NewFunctionSort(S, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}

	f1 := lg.NewConst("f", fSort)
	f2 := lg.NewConst("f", fSort)

	key1 := lg.Key(f1)
	key2 := lg.Key(f2)

	if key1 != key2 {
		t.Errorf("same-name same-sort Consts have different keys: %s vs %s", key1, key2)
	}

	// Different sort → different key
	T := &lg.UninterpretedSort{Name: "T"}
	fSortT, err := lg.NewFunctionSort(T, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	f3 := lg.NewConst("f", fSortT)
	key3 := lg.Key(f3)

	if key1 == key3 {
		t.Errorf("same-name different-sort Consts have same key: %s", key1)
	}
}

// TestMacroValueMapKeyIncludesSort verifies that macroValueMap memoization
// uses structural keys, so two macros for the same name on different sorts
// get separate cached results.
func TestMacroValueMapKeyIncludesSort(t *testing.T) {
	S := &lg.UninterpretedSort{Name: "S"}
	T := &lg.UninterpretedSort{Name: "T"}

	fSortS, err := lg.NewFunctionSort(S, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	fSortT, err := lg.NewFunctionSort(T, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}

	fS := lg.NewConst("f", fSortS)
	fT := lg.NewConst("f", fSortT)

	// Verify the keys used for macroValueMap would be distinct
	keyS := lg.Key(fS)
	keyT := lg.Key(fT)

	if keyS == keyT {
		t.Fatalf("macroValueMap keys should differ for same name different sorts: both=%s", keyS)
	}

	// Verify they work correctly in a map
	m := make(map[lg.NodeKey]mapFmlaRes)
	m[keyS] = mapFmlaRes{}
	m[keyT] = mapFmlaRes{}

	if len(m) != 2 {
		t.Errorf("map has %d entries, want 2", len(m))
	}
}
