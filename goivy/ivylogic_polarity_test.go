package goivy

import (
	"testing"
)

// helper to build p(X) where p: node -> bool
func makePofX(t *testing.T) (pOfX Expr, x *Variable) {
	t.Helper()
	nodeSort := &UninterpretedSort{Name: "node"}
	boolSort := Boolean

	var err error
	x, err = NewVariable("X", nodeSort)
	if err != nil {
		t.Fatal(err)
	}
	pSort := &FunctionSort{Sorts: []Sort{nodeSort, boolSort}}
	pSym := NewConst("p", pSort)
	pOfX, err = pSym.Call(x)
	if err != nil {
		t.Fatal(err)
	}
	return pOfX, x
}

// TestPolarityFlipThroughNot_SymbolsOverUniversals tests that Not
// correctly flips polarity in symbolsOverUniversalsRec.
//
// In Python ivy_logic.py lines 503-505:
//
//	if isinstance(fmla, Not):
//	    pos = not pos       # mutates pos before recursive call
//	argres = all([symbols_over_universals_rec(a, syms, pos, univs) ...])
//
// The formula: Not(IvyForAll(X, p(X)))
//
// Starting at pos=true:
//   - We hit Not, so pos flips to false.
//   - We recurse into IvyForAll(X, p(X)) with pos=false.
//   - Since pos(false) != isinstance(ForAll)(true), the quantifier
//     is NOT treated as universal (it's existential after skolemization).
//   - So X should NOT be added to univs.
//   - Therefore p should NOT appear in symbols_over_universals.
//
// BUG: Without the fix, pos stays true through Not, ForAll is treated
// as universal, X is added to univs, and p is incorrectly reported as
// a symbol over universals.
func TestPolarityFlipThroughNot_SymbolsOverUniversals(t *testing.T) {
	pOfX, x := makePofX(t)

	forallXpX, err := NewForAll([]*Variable{x}, pOfX)
	if err != nil {
		t.Fatal(err)
	}
	notForallXpX := &Not{Body: forallXpX}

	syms := SymbolsOverUniversals([]Expr{notForallXpX})
	// p should NOT be in the result — X is existential (Not flips ForAll → Exists)
	for _, s := range syms {
		if s.Name == "p" {
			t.Errorf("SymbolsOverUniversals incorrectly reports 'p' as over universals in Not(IvyForAll(X, p(X))). "+
				"Not should flip polarity, making ForAll existential. Got symbols: %v", syms)
		}
	}
}

// TestPolarityFlipThroughNot_UniversalVariables tests that Not
// correctly flips polarity in universalVariablesRec.
//
// The formula: Not(IvyForAll(X, p(X)))
//
// Starting at pos=true:
//   - We hit Not, so pos flips to false.
//   - We recurse into IvyForAll(X, p(X)) with pos=false.
//   - Since pos(false) != isinstance(ForAll)(true), X is NOT universal.
//   - So X should NOT appear in UniversalVariables.
//
// BUG: Without the fix, pos stays true through Not, and X is
// incorrectly reported as universally quantified.
func TestPolarityFlipThroughNot_UniversalVariables(t *testing.T) {
	pOfX, x := makePofX(t)

	forallXpX, err := NewForAll([]*Variable{x}, pOfX)
	if err != nil {
		t.Fatal(err)
	}
	notForallXpX := &Not{Body: forallXpX}

	univars := UniversalVariables([]Expr{notForallXpX})
	// X should NOT be universal — Not(IvyForAll(X, ...)) is IvyExists(X, ...)
	for _, v := range univars {
		if v.Name == "X" {
			t.Errorf("UniversalVariables incorrectly reports X as universal in Not(IvyForAll(X, p(X))). "+
				"Not should flip polarity, making ForAll existential. Got vars: %v", univars)
		}
	}
}

// TestPolarityDoubleNegation_UniversalVariables tests that double
// negation restores the original polarity.
//
// Not(Not(IvyForAll(X, p(X)))) — double negation, X should be universal again.
func TestPolarityDoubleNegation_UniversalVariables(t *testing.T) {
	pOfX, x := makePofX(t)

	forallXpX, err := NewForAll([]*Variable{x}, pOfX)
	if err != nil {
		t.Fatal(err)
	}
	notNotForallXpX := &Not{Body: &Not{Body: forallXpX}}

	univars := UniversalVariables([]Expr{notNotForallXpX})
	found := false
	for _, v := range univars {
		if v.Name == "X" {
			found = true
		}
	}
	if !found {
		t.Errorf("UniversalVariables should report X as universal in Not(Not(IvyForAll(X, p(X)))). "+
			"Double negation restores polarity. Got vars: %v", univars)
	}
}

// TestPolarityFlip_ExistsUnderNot_IsUniversal tests that
// Not(IvyExists(X, p(X))) makes X universal.
//
// pos flips to false via Not. Then we check: pos(false) == isForAll(false).
// false == false is true, so we DO add vars to univs.
// X SHOULD be universal in Not(IvyExists(X, p(X))).
func TestPolarityFlip_ExistsUnderNot_IsUniversal(t *testing.T) {
	pOfX, x := makePofX(t)

	existsXpX, err := NewExists([]*Variable{x}, pOfX)
	if err != nil {
		t.Fatal(err)
	}
	notExistsXpX := &Not{Body: existsXpX}

	univars := UniversalVariables([]Expr{notExistsXpX})
	found := false
	for _, v := range univars {
		if v.Name == "X" {
			found = true
		}
	}
	if !found {
		t.Errorf("UniversalVariables should report X as universal in Not(IvyExists(X, p(X))). "+
			"Not flips polarity: Exists becomes universal. Got vars: %v", univars)
	}
}
