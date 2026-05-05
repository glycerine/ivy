package goivy

import (
	"testing"
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
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	S := &UninterpretedSort{Name: "S"}
	v, _ := NewVariable("X", S)

	tests := []struct {
		name string
		fmla Expr
		pos  int
	}{
		{"Not/0", &Not{Body: a}, 0},
		{"Implies/0", &Implies{T1: a, T2: b}, 0},
		{"Implies/1", &Implies{T1: a, T2: b}, 1},
		{"And/0", &And{Terms: []Expr{a, b}}, 0},
		{"And/1", &And{Terms: []Expr{a, b}}, 1},
		{"Or/0", &Or{Terms: []Expr{a, b}}, 0},
		{"ForAll/0", &ForAll{Variables: []*Variable{v}, Body: a}, 0},
		{"Exists/0", &Exists{Variables: []*Variable{v}, Body: a}, 0},
		{"Ite/0", &Ite{Cond: a, Then: b, Else: a}, 0},
		{"Ite/1", &Ite{Cond: a, Then: b, Else: a}, 1},
		{"Ite/2", &Ite{Cond: a, Then: b, Else: a}, 2},
		{"Iff/0", &Iff{T1: a, T2: b}, 0},
		{"Iff/1", &Iff{T1: a, T2: b}, 1},
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
	S := &UninterpretedSort{Name: "S"}
	fs, _ := NewFunctionSort(S, Boolean)
	f := NewConst("f", fs)
	x := NewConst("x", S)
	app := MustApply(f, x)

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
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)

	// Not(Implies(a, b)) with initial pol=-1
	// Not → NegatePolarity(-1) = -1
	notPol := Polar(&Not{Body: &Implies{T1: a, T2: b}}, 0, -1)
	if notPol != -1 {
		t.Errorf("Not with pol=-1 should give -1, got %d", notPol)
	}
	// Implies under Not with pol=-1:
	//   pos=0 → NegatePolarity(-1) = -1
	//   pos=1 → -1
	imp := &Implies{T1: a, T2: b}
	if Polar(imp, 0, notPol) != -1 {
		t.Error("Implies/0 under mixed should be -1")
	}
	if Polar(imp, 1, notPol) != -1 {
		t.Error("Implies/1 under mixed should be -1")
	}
}
