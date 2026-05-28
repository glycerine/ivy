package goivy

import (
	"testing"
)

// TestMacroMapKeyIncludesSort verifies that macroMap uses structural keys
// (lg.NodeKey) that include the sort, matching Python's recstruct equality.
// Two macros defining the same name "f" but on different sorts must be
// stored as separate entries (Divergence 4 fix).
func TestMacroMapKeyIncludesSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	T := &UninterpretedSort{Name: "T"}

	// Create two function sorts: S->Bool and T->Bool
	fSortS, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	fSortT, err := NewFunctionSort(T, Boolean)
	if err != nil {
		t.Fatal(err)
	}

	// Two symbols with same name "f" but different sorts
	fS := NewConst("f", fSortS)
	fT := NewConst("f", fSortT)

	// Params for each definition
	xS, err := NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	xT, err := NewVariable("X", T)
	if err != nil {
		t.Fatal(err)
	}

	// LHS: f(X) for each sort
	lhsS, err := NewApply(fS, xS)
	if err != nil {
		t.Fatal(err)
	}
	lhsT, err := NewApply(fT, xT)
	if err != nil {
		t.Fatal(err)
	}

	// RHS: just use the variable itself (simplest possible body)
	defS := NewDefinition(lhsS, xS)
	defT := NewDefinition(lhsT, xT)

	// Create labeled formulas
	acfg := NewAstConfig()
	lfS := acfg.NewLabeledFormula(nil, nil)
	lfT := acfg.NewLabeledFormula(nil, nil)

	// Build fmlaPairs
	macros := []fmlaPair{
		{fmla: defS, source: lfS, loc: Location{Line: 1}},
		{fmla: defT, source: lfT, loc: Location{Line: 2}},
	}

	sig := NewSig()
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
	S := &UninterpretedSort{Name: "S"}
	fSort, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}

	f1 := NewConst("f", fSort)
	f2 := NewConst("f", fSort)

	key1 := Key(f1)
	key2 := Key(f2)

	if key1 != key2 {
		t.Errorf("same-name same-sort Consts have different keys: %s vs %s", key1, key2)
	}

	// Different sort → different key
	T := &UninterpretedSort{Name: "T"}
	fSortT, err := NewFunctionSort(T, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	f3 := NewConst("f", fSortT)
	key3 := Key(f3)

	if key1 == key3 {
		t.Errorf("same-name different-sort Consts have same key: %s", key1)
	}
}

// TestMacroValueMapKeyIncludesSort verifies that macroValueMap memoization
// uses structural keys, so two macros for the same name on different sorts
// get separate cached results.
func TestMacroValueMapKeyIncludesSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	T := &UninterpretedSort{Name: "T"}

	fSortS, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	fSortT, err := NewFunctionSort(T, Boolean)
	if err != nil {
		t.Fatal(err)
	}

	fS := NewConst("f", fSortS)
	fT := NewConst("f", fSortT)

	// Verify the keys used for macroValueMap would be distinct
	keyS := Key(fS)
	keyT := Key(fT)

	if keyS == keyT {
		t.Fatalf("macroValueMap keys should differ for same name different sorts: both=%s", keyS)
	}

	// Verify they work correctly in a map
	m := make(map[NodeKey]mapFmlaRes)
	m[keyS] = mapFmlaRes{}
	m[keyT] = mapFmlaRes{}

	if len(m) != 2 {
		t.Errorf("map has %d entries, want 2", len(m))
	}
}
