package fragment

import (
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
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
	S := &lg.UninterpretedSort{Name: "S"}
	x, _ := lg.NewVariable("X", S)
	y, _ := lg.NewVariable("Y", S)
	fSort, _ := lg.NewFunctionSort(S, S, S) // domain S,S -> range S
	f := lg.NewConst("f", fSort)
	app, err := lg.NewApply(f, x, y)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}

	fvs := lu.FreeVariables(app)
	if fvs.Len() != 2 {
		t.Errorf("expected 2 free vars in f(X, Y), got %d", fvs.Len())
	}
	if _, ok := fvs.Get2(lg.Key(x)); !ok {
		t.Error("X should be free in f(X, Y)")
	}
	if _, ok := fvs.Get2(lg.Key(y)); !ok {
		t.Error("Y should be free in f(X, Y)")
	}
}

// TestDivergence5_QuantifiedTerm verifies that FreeVariables on
// ForAll([x], P(x, y)) returns {y} only, matching Python's
// ilu.used_variables_ast which excludes bound variables from binders.
func TestDivergence5_QuantifiedTerm(t *testing.T) {
	S := &lg.UninterpretedSort{Name: "S"}
	x, _ := lg.NewVariable("X", S)
	y, _ := lg.NewVariable("Y", S)
	pSort, _ := lg.NewFunctionSort(S, S, lg.Boolean) // domain S,S -> Boolean
	P := lg.NewConst("P", pSort)
	body, err := lg.NewApply(P, x, y)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	fa, err := lg.NewForAll([]*lg.Variable{x}, body)
	if err != nil {
		t.Fatalf("NewForAll: %v", err)
	}

	fvs := lu.FreeVariables(fa)
	if fvs.Len() != 1 {
		t.Errorf("expected 1 free var in ForAll X. P(X,Y), got %d", fvs.Len())
	}
	if _, ok := fvs.Get2(lg.Key(x)); ok {
		t.Error("X should NOT be free in ForAll X. P(X, Y)")
	}
	if _, ok := fvs.Get2(lg.Key(y)); !ok {
		t.Error("Y should be free in ForAll X. P(X, Y)")
	}
}

// TestDivergence5_NestedBinders verifies that FreeVariables on
// ForAll([x], Exists([y], Q(x, y, z))) returns {z} only, confirming
// nested binder scopes work correctly — matching Python's
// ilu.used_variables_ast behavior.
func TestDivergence5_NestedBinders(t *testing.T) {
	S := &lg.UninterpretedSort{Name: "S"}
	x, _ := lg.NewVariable("X", S)
	y, _ := lg.NewVariable("Y", S)
	z, _ := lg.NewVariable("Z", S)
	qSort, _ := lg.NewFunctionSort(S, S, S, lg.Boolean) // domain S,S,S -> Boolean
	Q := lg.NewConst("Q", qSort)
	innerBody, err := lg.NewApply(Q, x, y, z)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	exists, err := lg.NewExists([]*lg.Variable{y}, innerBody)
	if err != nil {
		t.Fatalf("NewExists: %v", err)
	}
	fa, err := lg.NewForAll([]*lg.Variable{x}, exists)
	if err != nil {
		t.Fatalf("NewForAll: %v", err)
	}

	fvs := lu.FreeVariables(fa)
	if fvs.Len() != 1 {
		t.Errorf("expected 1 free var in ForAll X. Exists Y. Q(X,Y,Z), got %d", fvs.Len())
	}
	if _, ok := fvs.Get2(lg.Key(x)); ok {
		t.Error("X should NOT be free (bound by ForAll)")
	}
	if _, ok := fvs.Get2(lg.Key(y)); ok {
		t.Error("Y should NOT be free (bound by Exists)")
	}
	if _, ok := fvs.Get2(lg.Key(z)); !ok {
		t.Error("Z should be free in ForAll X. Exists Y. Q(X,Y,Z)")
	}
}
