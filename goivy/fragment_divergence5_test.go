package goivy

import (
	"testing"
)

// Divergence 5: Python ivy_fragment.py:260 uses ilu.used_variables_ast(v)
// and Go fragment.go used lu.FreeVariables(appArgs[i]).
//
// The set of variables matched, but order also matters: Python's
// ilu.used_variables_ast wraps variables_ast with dict.fromkeys, preserving
// DFS first-occurrence order. Go's FreeVariables stores an ordered map keyed
// structurally, which is deterministic but not source-order preserving.
//
// These tests exercise the same expression ordering used by createMacroMaps
// when Python records dependencies for lower-case macro formals.

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

func TestCreateMacroMapsLowercaseFormalUsesActualVariableSourceOrderLikePython(t *testing.T) {
	ResetCounter()

	S := &UninterpretedSort{Name: "S"}
	Z, _ := NewVariable("Z", S)
	A, _ := NewVariable("A", S)
	n := NewConst("n", S)

	pairSort, err := NewFunctionSort(S, S, S)
	if err != nil {
		t.Fatal(err)
	}
	pair := NewConst("pair", pairSort)
	pairZA, err := NewApply(pair, Z, A)
	if err != nil {
		t.Fatal(err)
	}

	macroSort, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	macro := NewConst("macro", macroSort)
	macroLHS, err := NewApply(macro, n)
	if err != nil {
		t.Fatal(err)
	}
	macroDef := NewDefinition(macroLHS, &Eq{T1: n, T2: n})

	macroCall, err := NewApply(macro, pairZA)
	if err != nil {
		t.Fatal(err)
	}

	acfg := NewAstConfig()
	lfMacro := acfg.NewLabeledFormula(nil, nil)
	lfAssume := acfg.NewLabeledFormula(nil, nil)
	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)
	c.universallyQuantifiedVars[makeVarID(Z)] = Z
	c.universallyQuantifiedVars[makeVarID(A)] = A

	c.createMacroMaps(
		[]fmlaPair{{fmla: macroCall, source: lfAssume, lineno: 1}},
		nil,
		[]fmlaPair{{fmla: macroDef, source: lfMacro, lineno: 2}},
	)

	zNode := c.stratMap[varKey(Z)]
	aNode := c.stratMap[varKey(A)]
	if zNode == nil || aNode == nil {
		t.Fatalf("expected strat nodes for Z and A, got Z=%v A=%v", zNode, aNode)
	}
	if zNode.ID >= aNode.ID {
		t.Fatalf("lower-case macro formal dependencies were created out of source order: Z id=%d A id=%d", zNode.ID, aNode.ID)
	}
	deps := c.macroDepMap[varID{name: "n", sort: "S"}]
	if len(deps) != 2 || !deps[zNode] || !deps[aNode] {
		t.Fatalf("macroDepMap[n:S] = %v, want Z and A nodes", deps)
	}
}
