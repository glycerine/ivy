package goivy

import (
	"testing"
)

// These tests verify that VariableUniqifier produces the same naming
// behavior as Python's VariableUniqifier. The algorithm should be identical:
// same UniqueRenamer, same suffix sequence (a,b,...,z,a0,b0,...), same
// traversal order, same binder save/restore logic.

func TestVariableUniqifierExactNames(t *testing.T) {
	// IvyForAll([X:S], IvyExists([Y:S], Eq(X, Y)))
	// First call: X → "X", Y → "Y" (both first-time, unused)
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	body := &Eq{T1: X, T2: Y}
	inner := &Exists{Variables: []*Variable{Y}, Body: body}
	outer := &ForAll{Variables: []*Variable{X}, Body: inner}

	vu := NewVariableUniqifier(nil)
	result := vu.Uniquify(outer)

	fa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("expected ForAll, got %T", result)
	}
	if fa.Variables[0].Name != "X" {
		t.Errorf("first ForAll var should be X, got %s", fa.Variables[0].Name)
	}
	ex, ok := fa.Body.(*Exists)
	if !ok {
		t.Fatalf("expected Exists body, got %T", fa.Body)
	}
	if ex.Variables[0].Name != "Y" {
		t.Errorf("first Exists var should be Y, got %s", ex.Variables[0].Name)
	}
	eq, ok := ex.Body.(*Eq)
	if !ok {
		t.Fatalf("expected Eq body, got %T", ex.Body)
	}
	// Both T1 and T2 should be the renamed versions
	t1v, ok := eq.T1.(*Variable)
	if !ok {
		t.Fatalf("expected Variable in Eq.T1, got %T", eq.T1)
	}
	if t1v.Name != "X" {
		t.Errorf("Eq.T1 should be X, got %s", t1v.Name)
	}
	t2v, ok := eq.T2.(*Variable)
	if !ok {
		t.Fatalf("expected Variable in Eq.T2, got %T", eq.T2)
	}
	if t2v.Name != "Y" {
		t.Errorf("Eq.T2 should be Y, got %s", t2v.Name)
	}
}

func TestVariableUniqifierMultipleFormulas(t *testing.T) {
	// Calling Uniquify twice on the same formula with the same uniqifier
	// should produce different names the second time (since names accumulate).
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	body := &ForAll{Variables: []*Variable{X}, Body: X}

	vu := NewVariableUniqifier(nil)

	// First call: X → "X"
	r1 := vu.Uniquify(body)
	fa1 := r1.(*ForAll)
	if fa1.Variables[0].Name != "X" {
		t.Errorf("first call: expected X, got %s", fa1.Variables[0].Name)
	}

	// Second call: X → "X_a" (X is already used)
	r2 := vu.Uniquify(body)
	fa2 := r2.(*ForAll)
	if fa2.Variables[0].Name != "X_a" {
		t.Errorf("second call: expected X_a, got %s", fa2.Variables[0].Name)
	}

	// Third call: X → "X_b"
	r3 := vu.Uniquify(body)
	fa3 := r3.(*ForAll)
	if fa3.Variables[0].Name != "X_b" {
		t.Errorf("third call: expected X_b, got %s", fa3.Variables[0].Name)
	}
}

func TestVariableUniqifierBinderShadowing(t *testing.T) {
	// IvyForAll([X], And(X, IvyForAll([X], X)))
	// Outer X → "X" (first time)
	// Inner X → "X_a" (second unique name for X)
	// The outer body reference to X should use "X", not "X_a"
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	innerBody := &ForAll{Variables: []*Variable{X}, Body: X}
	outerBody := &And{Terms: []Expr{X, innerBody}}
	fmla := &ForAll{Variables: []*Variable{X}, Body: outerBody}

	vu := NewVariableUniqifier(nil)
	result := vu.Uniquify(fmla)

	fa := result.(*ForAll)
	outerVarName := fa.Variables[0].Name
	if outerVarName != "X" {
		t.Errorf("outer var should be X, got %s", outerVarName)
	}

	and := fa.Body.(*And)
	// First term of And should reference the OUTER renamed X
	outerRef, ok := and.Terms[0].(*Variable)
	if !ok {
		t.Fatalf("expected Variable, got %T", and.Terms[0])
	}
	if outerRef.Name != outerVarName {
		t.Errorf("outer ref should be %s, got %s", outerVarName, outerRef.Name)
	}

	// Inner ForAll should have a DIFFERENT name for X
	innerFa := and.Terms[1].(*ForAll)
	innerVarName := innerFa.Variables[0].Name
	if innerVarName == outerVarName {
		t.Errorf("inner var should differ from outer var %s", outerVarName)
	}
	if innerVarName != "X_a" {
		t.Errorf("inner var should be X_a, got %s", innerVarName)
	}

	// Inner body should reference the inner renamed X
	innerRef := innerFa.Body.(*Variable)
	if innerRef.Name != innerVarName {
		t.Errorf("inner body ref should be %s, got %s", innerVarName, innerRef.Name)
	}
}

func TestVariableUniqifierFreeVariables(t *testing.T) {
	// Free variables also get uniquified.
	// Formula: Eq(X, Y) — both X and Y are free
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	fmla := &Eq{T1: X, T2: Y}

	vu := NewVariableUniqifier(nil)
	result := vu.Uniquify(fmla)

	eq := result.(*Eq)
	xv := eq.T1.(*Variable)
	yv := eq.T2.(*Variable)

	// First time: X → "X", Y → "Y"
	if xv.Name != "X" {
		t.Errorf("free X should be X, got %s", xv.Name)
	}
	if yv.Name != "Y" {
		t.Errorf("free Y should be Y, got %s", yv.Name)
	}

	// Second call with same uniqifier: X → "X_a", Y → "Y_a"
	result2 := vu.Uniquify(fmla)
	eq2 := result2.(*Eq)
	xv2 := eq2.T1.(*Variable)
	yv2 := eq2.T2.(*Variable)
	if xv2.Name != "X_a" {
		t.Errorf("second free X should be X_a, got %s", xv2.Name)
	}
	if yv2.Name != "Y_a" {
		t.Errorf("second free Y should be Y_a, got %s", yv2.Name)
	}
}

func TestVariableUniqifierApplyPreservesFunc(t *testing.T) {
	// IvyApply(f, [X, Y]) — f should be preserved, X and Y get uniquified
	S := &UninterpretedSort{Name: "S"}
	fs, _ := NewFunctionSort(S, S, Boolean)
	f := NewConst("f", fs)
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	app := MustApply(f, X, Y)

	vu := NewVariableUniqifier(nil)
	result := vu.Uniquify(app)

	rapp, ok := result.(*Apply)
	if !ok {
		t.Fatalf("expected Apply, got %T", result)
	}
	// Func should be preserved (same pointer)
	if rapp.Func != f {
		t.Error("Apply.Func should be preserved through uniquification")
	}
	// Terms should be uniquified variables
	if len(rapp.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(rapp.Terms))
	}
	xv := rapp.Terms[0].(*Variable)
	yv := rapp.Terms[1].(*Variable)
	if xv.Name != "X" {
		t.Errorf("first term should be X, got %s", xv.Name)
	}
	if yv.Name != "Y" {
		t.Errorf("second term should be Y, got %s", yv.Name)
	}
}

func TestVariableUniqifierSuffixSequence(t *testing.T) {
	// Verify the suffix sequence matches Python: a, b, c, ..., z, a0, b0, ...
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("V", S)
	fmla := &ForAll{Variables: []*Variable{X}, Body: X}

	vu := NewVariableUniqifier(nil)

	// Generate 28 unique names for "V"
	names := make([]string, 28)
	for i := 0; i < 28; i++ {
		r := vu.Uniquify(fmla)
		fa := r.(*ForAll)
		names[i] = fa.Variables[0].Name
	}

	// First: "V" (original, unused)
	if names[0] != "V" {
		t.Errorf("names[0] should be V, got %s", names[0])
	}
	// Second through 27th: V_a, V_b, ..., V_z
	for i := 1; i <= 26; i++ {
		expected := "V_" + string(rune('a'+i-1))
		if names[i] != expected {
			t.Errorf("names[%d] should be %s, got %s", i, expected, names[i])
		}
	}
	// 28th: V_a0
	if names[27] != "V_a0" {
		t.Errorf("names[27] should be V_a0, got %s", names[27])
	}
}

func TestVariableUniqifierInvMap(t *testing.T) {
	// Verify InvMap tracks renamed → original correctly
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	fmla := &ForAll{Variables: []*Variable{X}, Body: X}

	vu := NewVariableUniqifier(nil)
	vu.Uniquify(fmla) // X → X
	vu.Uniquify(fmla) // X → X_a

	// InvMap should have entries for both renamed vars
	if len(vu.InvMap) < 2 {
		t.Errorf("InvMap should have at least 2 entries, got %d", len(vu.InvMap))
	}

	// Check that "X_a" maps back to original "X"
	found := false
	for k, v := range vu.InvMap {
		if k == Key(&Variable{Name: "X_a", VSort: S}) {
			if v.Name != "X" {
				t.Errorf("InvMap[X_a] should map to X, got %s", v.Name)
			}
			found = true
		}
	}
	if !found {
		t.Error("InvMap should contain entry for X_a")
	}
}

func TestAlphaAvoidMapEmptyVsRenamesShadowedBound(t *testing.T) {
	// Regression: AlphaAvoidMap must still rename bound variables that
	// shadow free variables even when vs is empty. Python's alpha_avoid
	// has no early return for empty vs — it always collects free variables
	// and renames clashing bound variables.
	//
	// Formula: And(M, IvyForAll([M:S], Eq(M, M)))
	//   - M is FREE in the first And term
	//   - M is BOUND in the ForAll
	// After AlphaAvoidMap with empty vs, the bound M must become M_a.
	S := &UninterpretedSort{Name: "S"}
	M, _ := NewVariable("M", S)
	eqBody := &Eq{T1: M, T2: M}
	forall := &ForAll{Variables: []*Variable{M}, Body: eqBody}
	fmla := &And{Terms: []Expr{M, forall}}

	emptyVs := make(map[NodeKey]Expr)
	result := AlphaAvoidMap(fmla, emptyVs)

	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}

	// Free M in first term must stay "M"
	freeM, ok := and.Terms[0].(*Variable)
	if !ok {
		t.Fatalf("expected Variable in And.Terms[0], got %T", and.Terms[0])
	}
	if freeM.Name != "M" {
		t.Errorf("free M should stay M, got %s", freeM.Name)
	}

	// Bound M in ForAll must be renamed to M_a
	fa, ok := and.Terms[1].(*ForAll)
	if !ok {
		t.Fatalf("expected ForAll in And.Terms[1], got %T", and.Terms[1])
	}
	if fa.Variables[0].Name != "M_a" {
		t.Errorf("bound M should be renamed to M_a, got %s", fa.Variables[0].Name)
	}

	// Body references inside ForAll must use the renamed M_a
	eq, ok := fa.Body.(*Eq)
	if !ok {
		t.Fatalf("expected Eq in ForAll body, got %T", fa.Body)
	}
	t1v := eq.T1.(*Variable)
	t2v := eq.T2.(*Variable)
	if t1v.Name != "M_a" {
		t.Errorf("Eq.T1 inside ForAll should be M_a, got %s", t1v.Name)
	}
	if t2v.Name != "M_a" {
		t.Errorf("Eq.T2 inside ForAll should be M_a, got %s", t2v.Name)
	}
}

func TestAlphaAvoidMapNonEmptyVsAlsoRenamesShadowed(t *testing.T) {
	// When vs is non-empty AND the formula has free-vs-bound shadowing,
	// both the vs names and free variable names must be reserved.
	//
	// Formula: And(M, IvyForAll([M:S], Eq(M, M)))
	// vs contains a variable named "Q"
	// Both "Q" and "M" (free) should be reserved. Bound M → M_a.
	S := &UninterpretedSort{Name: "S"}
	M, _ := NewVariable("M", S)
	Q, _ := NewVariable("Q", S)
	eqBody := &Eq{T1: M, T2: M}
	forall := &ForAll{Variables: []*Variable{M}, Body: eqBody}
	fmla := &And{Terms: []Expr{M, forall}}

	vs := map[NodeKey]Expr{Key(Q): Q}
	result := AlphaAvoidMap(fmla, vs)

	and := result.(*And)
	fa := and.Terms[1].(*ForAll)
	if fa.Variables[0].Name != "M_a" {
		t.Errorf("bound M should be renamed to M_a, got %s", fa.Variables[0].Name)
	}
}
