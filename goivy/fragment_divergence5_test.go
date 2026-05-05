package goivy

import (
	"testing"
)

// Divergence 5: Python ivy_fragment.py:260 uses ilu.used_variables_ast(v)
// and Go fragment.go:451 uses lu.FreeVariables(appArgs[i]).
//
// Investigation shows this is a FALSE ALARM: Python's ilu.used_variables_ast
// (from ivy_logic_utils.py:485) wraps variables_ast (line 464) which is
// explicitly commented "# get free variables" and excludes bound variables
// from binders. This matches Go's lu.FreeVariables behavior exactly.
//
// These tests confirm the equivalence by exercising the same expression
// patterns that createMacroMaps passes to FreeVariables.

// TestDivergence5_SimpleTerm verifies that FreeVariables on a simple
// application f(x, y) returns {x, y}, matching Python's
// ilu.used_variables_ast behavior for non-binder expressions.
func TestDivergence5_SimpleTerm(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x, _ := NewVariable("X", S)
	y, _ := NewVariable("Y", S)
	fSort, _ := NewFunctionSort(S, S, S) // domain S,S -> range S
	f := NewConst("f", fSort)
	app, err := NewApply(f, x, y)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}

	fvs := FreeVariables(app)
	if fvs.Len() != 2 {
		t.Errorf("expected 2 free vars in f(X, Y), got %d", fvs.Len())
	}
	if _, ok := fvs.Get2(Key(x)); !ok {
		t.Error("X should be free in f(X, Y)")
	}
	if _, ok := fvs.Get2(Key(y)); !ok {
		t.Error("Y should be free in f(X, Y)")
	}
}

// TestDivergence5_QuantifiedTerm verifies that FreeVariables on
// ForAll([x], P(x, y)) returns {y} only, matching Python's
// ilu.used_variables_ast which excludes bound variables from binders.
func TestDivergence5_QuantifiedTerm(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x, _ := NewVariable("X", S)
	y, _ := NewVariable("Y", S)
	pSort, _ := NewFunctionSort(S, S, Boolean) // domain S,S -> Boolean
	P := NewConst("P", pSort)
	body, err := NewApply(P, x, y)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	fa, err := NewForAll([]*LogicVariable{x}, body)
	if err != nil {
		t.Fatalf("NewForAll: %v", err)
	}

	fvs := FreeVariables(fa)
	if fvs.Len() != 1 {
		t.Errorf("expected 1 free var in ForAll X. P(X,Y), got %d", fvs.Len())
	}
	if _, ok := fvs.Get2(Key(x)); ok {
		t.Error("X should NOT be free in ForAll X. P(X, Y)")
	}
	if _, ok := fvs.Get2(Key(y)); !ok {
		t.Error("Y should be free in ForAll X. P(X, Y)")
	}
}

// TestDivergence5_NestedBinders verifies that FreeVariables on
// ForAll([x], Exists([y], Q(x, y, z))) returns {z} only, confirming
// nested binder scopes work correctly — matching Python's
// ilu.used_variables_ast behavior.
func TestDivergence5_NestedBinders(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x, _ := NewVariable("X", S)
	y, _ := NewVariable("Y", S)
	z, _ := NewVariable("Z", S)
	qSort, _ := NewFunctionSort(S, S, S, Boolean) // domain S,S,S -> Boolean
	Q := NewConst("Q", qSort)
	innerBody, err := NewApply(Q, x, y, z)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	exists, err := NewExists([]*LogicVariable{y}, innerBody)
	if err != nil {
		t.Fatalf("NewExists: %v", err)
	}
	fa, err := NewForAll([]*LogicVariable{x}, exists)
	if err != nil {
		t.Fatalf("NewForAll: %v", err)
	}

	fvs := FreeVariables(fa)
	if fvs.Len() != 1 {
		t.Errorf("expected 1 free var in ForAll X. Exists Y. Q(X,Y,Z), got %d", fvs.Len())
	}
	if _, ok := fvs.Get2(Key(x)); ok {
		t.Error("X should NOT be free (bound by ForAll)")
	}
	if _, ok := fvs.Get2(Key(y)); ok {
		t.Error("Y should NOT be free (bound by Exists)")
	}
	if _, ok := fvs.Get2(Key(z)); !ok {
		t.Error("Z should be free in ForAll X. Exists Y. Q(X,Y,Z)")
	}
}
