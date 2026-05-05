package goivy

import (
	"testing"
)

// TestEqExprKeyDifferentExpsSameSort verifies that eqExprKey produces
// different keys for different expressions of the same sort, matching
// Python's per-expression equality keying: il.Symbol('=', fmla.args[0]).
func TestEqExprKeyDifferentExpsSameSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x, _ := NewVariable("X", S)
	y, _ := NewVariable("Y", S)
	z, _ := NewVariable("Z", S)

	key1 := eqExprKey(x)
	key2 := eqExprKey(z)
	key3 := eqExprKey(y)
	if key1 == key2 {
		t.Errorf("eqExprKey(x) == eqExprKey(z): both %s; should differ", key1)
	}
	if key1 == key3 {
		t.Errorf("eqExprKey(x) == eqExprKey(y): both %s; should differ", key1)
	}
}

// TestEqExprKeySameExprSameKey verifies that structurally identical
// expressions produce the same eqExprKey (structural equality via Sexp).
func TestEqExprKeySameExprSameKey(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x1, _ := NewVariable("X", S)
	x2, _ := NewVariable("X", S)

	key1 := eqExprKey(x1)
	key2 := eqExprKey(x2)
	if key1 != key2 {
		t.Errorf("eqExprKey for two Variables named 'x' of sort S differ: %s vs %s", key1, key2)
	}
}

// TestMapFmlaEqualityPerExpressionKeys verifies that mapFmla creates
// separate strat nodes for equalities with different first operands,
// matching Python's behavior (Divergence 1 fix).
func TestMapFmlaEqualityPerExpressionKeys(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x, _ := NewVariable("X", S)
	y, _ := NewVariable("Y", S)
	z, _ := NewVariable("Z", S)

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)

	// Register x, y, z as universal vars
	for _, v := range []*LogicVariable{x, y, z} {
		vid := makeVarID(v)
		c.universallyQuantifiedVars[vid] = v
	}

	eq1 := &Eq{T1: x, T2: y}
	eq2 := &Eq{T1: z, T2: y}

	c.mapFmla(0, eq1, 0)
	c.mapFmla(0, eq2, 0)

	// Both expression-keyed entries should exist in the strat map.
	nodeForX := c.stratMap[eqExprKey(x)]
	nodeForZ := c.stratMap[eqExprKey(z)]
	if nodeForX == nil {
		t.Fatal("stratMap should have an entry for eqExprKey(x)")
	}
	if nodeForZ == nil {
		t.Fatal("stratMap should have an entry for eqExprKey(z)")
	}
	// The keys must be different for different expressions of the same sort.
	if eqExprKey(x) == eqExprKey(z) {
		t.Error("keys should be different for different expressions of the same sort")
	}
}
