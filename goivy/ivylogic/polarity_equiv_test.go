package ivylogic

import (
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// These tests verify that Go's pol=-1 is equivalent to Python's pol=None
// for the polarity system. This documents the intentional mapping:
//   pol=0  → positive polarity (even negations)
//   pol=1  → negative polarity (odd negations)
//   pol=-1 → both/mixed polarity (Python's None)

func TestNegatePolarityMixed(t *testing.T) {
	// Python: negate_polarity(None) → None
	// Go:     NegatePolarity(-1)    → -1
	if NegatePolarity(-1) != -1 {
		t.Errorf("NegatePolarity(-1) should be -1, got %d", NegatePolarity(-1))
	}
}

func TestNegatePolarityNormal(t *testing.T) {
	if NegatePolarity(0) != 1 {
		t.Errorf("NegatePolarity(0) should be 1, got %d", NegatePolarity(0))
	}
	if NegatePolarity(1) != 0 {
		t.Errorf("NegatePolarity(1) should be 0, got %d", NegatePolarity(1))
	}
}

func TestPolarMixedPropagatesToAllFormulas(t *testing.T) {
	// When pol=-1, Polar should return -1 for ALL formula types and positions.
	// This matches Python where pol=None always propagates as None.
	a := lg.NewConst("a", lg.Boolean)
	b := lg.NewConst("b", lg.Boolean)
	S := &lg.UninterpretedSort{Name: "S"}
	v, _ := lg.NewVariable("X", S)

	tests := []struct {
		name string
		fmla lg.Expr
		pos  int
	}{
		{"Not/0", &lg.Not{Body: a}, 0},
		{"Implies/0", &lg.Implies{T1: a, T2: b}, 0},
		{"Implies/1", &lg.Implies{T1: a, T2: b}, 1},
		{"And/0", &lg.And{Terms: []lg.Expr{a, b}}, 0},
		{"And/1", &lg.And{Terms: []lg.Expr{a, b}}, 1},
		{"Or/0", &lg.Or{Terms: []lg.Expr{a, b}}, 0},
		{"ForAll/0", &lg.ForAll{Variables: []*lg.Variable{v}, Body: a}, 0},
		{"Exists/0", &lg.Exists{Variables: []*lg.Variable{v}, Body: a}, 0},
		{"Ite/0", &lg.Ite{Cond: a, Then: b, Else: a}, 0},
		{"Ite/1", &lg.Ite{Cond: a, Then: b, Else: a}, 1},
		{"Ite/2", &lg.Ite{Cond: a, Then: b, Else: a}, 2},
		{"Iff/0", &lg.Iff{T1: a, T2: b}, 0},
		{"Iff/1", &lg.Iff{T1: a, T2: b}, 1},
	}

	for _, tt := range tests {
		result := Polar(tt.fmla, tt.pos, -1)
		if result != -1 {
			t.Errorf("Polar(%s, %d, -1) should be -1, got %d", tt.name, tt.pos, result)
		}
	}
}

func TestPolarApplyDefaultIsMixed(t *testing.T) {
	// Apply (and other unknown types) always return -1 regardless of input pol.
	// Python: polar() returns None for unknown types.
	S := &lg.UninterpretedSort{Name: "S"}
	fs, _ := lg.NewFunctionSort(S, lg.Boolean)
	f := lg.NewConst("f", fs)
	x := lg.NewConst("x", S)
	app := &lg.Apply{Func: f, Terms: []lg.Expr{x}}

	for _, pol := range []int{0, 1, -1} {
		result := Polar(app, 0, pol)
		if result != -1 {
			t.Errorf("Polar(Apply, 0, %d) should be -1, got %d", pol, result)
		}
	}
}

func TestIsStrictInequalityMixed(t *testing.T) {
	// Python: is_strict_inequality_symbol(sym, None) → False for all symbols
	// Go:     IsStrictInequalitySymbol(name, -1) → false for all names
	for _, name := range []string{"<", ">", "<=", ">=", "=", "+"} {
		if IsStrictInequalitySymbol(name, -1) {
			t.Errorf("IsStrictInequalitySymbol(%q, -1) should be false", name)
		}
	}
}

func TestIsStrictInequalityNormal(t *testing.T) {
	// pol=0: < and > are strict
	if !IsStrictInequalitySymbol("<", 0) {
		t.Error("< should be strict at pol=0")
	}
	if !IsStrictInequalitySymbol(">", 0) {
		t.Error("> should be strict at pol=0")
	}
	if IsStrictInequalitySymbol("<=", 0) {
		t.Error("<= should NOT be strict at pol=0")
	}

	// pol=1: <= and >= are strict (negation flips)
	if !IsStrictInequalitySymbol("<=", 1) {
		t.Error("<= should be strict at pol=1")
	}
	if !IsStrictInequalitySymbol(">=", 1) {
		t.Error(">= should be strict at pol=1")
	}
	if IsStrictInequalitySymbol("<", 1) {
		t.Error("< should NOT be strict at pol=1")
	}
}

func TestPolarityMixedPropagatesRecursively(t *testing.T) {
	// Simulate recursive polarity propagation through nested formulas.
	// Starting with pol=-1, every child should also get pol=-1.
	// This matches Python where None stays None through any nesting.
	a := lg.NewConst("a", lg.Boolean)
	b := lg.NewConst("b", lg.Boolean)

	// Not(Implies(a, b)) with initial pol=-1
	// Not → NegatePolarity(-1) = -1
	notPol := Polar(&lg.Not{Body: &lg.Implies{T1: a, T2: b}}, 0, -1)
	if notPol != -1 {
		t.Errorf("Not with pol=-1 should give -1, got %d", notPol)
	}
	// Implies under Not with pol=-1:
	//   pos=0 → NegatePolarity(-1) = -1
	//   pos=1 → -1
	imp := &lg.Implies{T1: a, T2: b}
	if Polar(imp, 0, notPol) != -1 {
		t.Error("Implies/0 under mixed should be -1")
	}
	if Polar(imp, 1, notPol) != -1 {
		t.Error("Implies/1 under mixed should be -1")
	}
}
