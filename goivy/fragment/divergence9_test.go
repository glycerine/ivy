package fragment

import (
	"testing"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// TestDivergence9SameNameDifferentSortNotRecursive verifies that the
// definition recursion check uses structural equality (name+sort), not
// just name strings. A definition f:S1->Bool should NOT be considered
// recursive just because f:S2->Bool appears in the RHS.
// This matches Python's `defines() not in symbols_ilu_ast(rhs)` which
// uses recstruct __eq__ comparing name+sort.
func TestDivergence9SameNameDifferentSortNotRecursive(t *testing.T) {
	S1 := &lg.UninterpretedSort{Name: "S1"}
	S2 := &lg.UninterpretedSort{Name: "S2"}

	// Function sorts: S1->Bool and S2->Bool
	fSortS1, err := lg.NewFunctionSort(S1, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	fSortS2, err := lg.NewFunctionSort(S2, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}

	// Two "f" symbols with different sorts
	fS1 := lg.NewConst("f", fSortS1) // defining symbol
	fS2 := lg.NewConst("f", fSortS2) // appears in RHS

	// Build definition: f:S1(x:S1) = f:S2(x:S2)
	xS1, err := lg.NewVariable("X", S1)
	if err != nil {
		t.Fatal(err)
	}
	xS2, err := lg.NewVariable("X", S2)
	if err != nil {
		t.Fatal(err)
	}
	lhs, err := lg.NewApply(fS1, xS1)
	if err != nil {
		t.Fatal(err)
	}
	rhs, err := lg.NewApply(fS2, xS2)
	if err != nil {
		t.Fatal(err)
	}
	def := lg.NewDefinition(lhs, rhs)

	// Check recursion the same way fragment.go does
	defSym := def.Defines()
	symsInRHS := il.SymbolsAst(def.Rhs)
	isRecursive := false
	for _, s := range symsInRHS {
		if defSym != nil && lg.Key(s) == lg.Key(defSym) {
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
	S := &lg.UninterpretedSort{Name: "S"}
	fSort, err := lg.NewFunctionSort(S, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}

	f := lg.NewConst("f", fSort)
	X, err := lg.NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}

	// Build definition: f(X) = f(X) — genuinely recursive
	lhs, err := lg.NewApply(f, X)
	if err != nil {
		t.Fatal(err)
	}
	// RHS uses a separately constructed Const with same name+sort
	fAgain := lg.NewConst("f", fSort)
	rhs, err := lg.NewApply(fAgain, X)
	if err != nil {
		t.Fatal(err)
	}
	def := lg.NewDefinition(lhs, rhs)

	defSym := def.Defines()
	symsInRHS := il.SymbolsAst(def.Rhs)
	isRecursive := false
	for _, s := range symsInRHS {
		if defSym != nil && lg.Key(s) == lg.Key(defSym) {
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
	S1 := &lg.UninterpretedSort{Name: "S1"}
	S2 := &lg.UninterpretedSort{Name: "S2"}

	f1 := lg.NewConst("f", S1)
	f2 := lg.NewConst("f", S2)
	f3 := lg.NewConst("f", S1) // same as f1

	if lg.Key(f1) == lg.Key(f2) {
		t.Errorf("same name different sort should have different keys: %s vs %s", lg.Key(f1), lg.Key(f2))
	}
	if lg.Key(f1) != lg.Key(f3) {
		t.Errorf("same name same sort should have equal keys: %s vs %s", lg.Key(f1), lg.Key(f3))
	}
}
