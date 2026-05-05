package goivy

import (
	"testing"
)

// TestDivergence9SameNameDifferentSortNotRecursive verifies that the
// definition recursion check uses structural equality (name+sort), not
// just name strings. A definition f:S1->Bool should NOT be considered
// recursive just because f:S2->Bool appears in the RHS.
// This matches Python's `defines() not in symbols_ilu_ast(rhs)` which
// uses recstruct __eq__ comparing name+sort.
func TestDivergence9SameNameDifferentSortNotRecursive(t *testing.T) {
	S1 := &UninterpretedSort{Name: "S1"}
	S2 := &UninterpretedSort{Name: "S2"}

	// Function sorts: S1->Bool and S2->Bool
	fSortS1, err := NewFunctionSort(S1, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	fSortS2, err := NewFunctionSort(S2, Boolean)
	if err != nil {
		t.Fatal(err)
	}

	// Two "f" symbols with different sorts
	fS1 := NewConst("f", fSortS1) // defining symbol
	fS2 := NewConst("f", fSortS2) // appears in RHS

	// Build definition: f:S1(x:S1) = f:S2(x:S2)
	xS1, err := NewVariable("X", S1)
	if err != nil {
		t.Fatal(err)
	}
	xS2, err := NewVariable("X", S2)
	if err != nil {
		t.Fatal(err)
	}
	lhs, err := NewApply(fS1, xS1)
	if err != nil {
		t.Fatal(err)
	}
	rhs, err := NewApply(fS2, xS2)
	if err != nil {
		t.Fatal(err)
	}
	def := NewDefinition(lhs, rhs)

	// Check recursion the same way fragment.go does
	defSym := def.Defines()
	symsInRHS := SymbolsIluAst(def.Rhs)
	isRecursive := false
	for s := range symsInRHS {
		if defSym != nil && Key(s) == Key(defSym) {
			isRecursive = true
			break
		}
	}

	if isRecursive {
		t.Errorf("definition should NOT be recursive: defSym=%s (sort %s) vs RHS sym=%s (sort %s)",
			defSym, fSortS1.Sexp(), fS2.Name, fSortS2.Sexp())
	}
}

// TestDivergence9SameNameSameSortIsRecursive verifies that genuine
// recursion (same name AND same sort) is still detected correctly.
func TestDivergence9SameNameSameSortIsRecursive(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fSort, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}

	f := NewConst("f", fSort)
	X, err := NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}

	// Build definition: f(X) = f(X) — genuinely recursive
	lhs, err := NewApply(f, X)
	if err != nil {
		t.Fatal(err)
	}
	// RHS uses a separately constructed Const with same name+sort
	fAgain := NewConst("f", fSort)
	rhs, err := NewApply(fAgain, X)
	if err != nil {
		t.Fatal(err)
	}
	def := NewDefinition(lhs, rhs)

	defSym := def.Defines()
	symsInRHS := SymbolsIluAst(def.Rhs) // here, not defined.
	isRecursive := false
	for s := range symsInRHS {
		if defSym != nil && Key(s) == Key(defSym) {
			isRecursive = true
			break
		}
	}

	if !isRecursive {
		t.Error("definition f(x) = f(x) should be detected as recursive")
	}
}

// TestDivergence9KeyStructuralEquality confirms that lg.Key() on Const
// objects with same name but different sorts produces different keys,
// while same name+sort produces equal keys.
func TestDivergence9KeyStructuralEquality(t *testing.T) {
	S1 := &UninterpretedSort{Name: "S1"}
	S2 := &UninterpretedSort{Name: "S2"}

	f1 := NewConst("f", S1)
	f2 := NewConst("f", S2)
	f3 := NewConst("f", S1) // same as f1

	if Key(f1) == Key(f2) {
		t.Errorf("same name different sort should have different keys: %s vs %s", Key(f1), Key(f2))
	}
	if Key(f1) != Key(f3) {
		t.Errorf("same name same sort should have equal keys: %s vs %s", Key(f1), Key(f3))
	}
}
