package logicutil

import (
	"errors"
	"testing"

	"github.com/glycerine/ivy/goivy/logic"
)

func mustFS(t *testing.T, sorts ...logic.Sort) *logic.FunctionSort {
	t.Helper()
	fs, err := logic.NewFunctionSort(sorts...)
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

func TestFreeVariablesSimple(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	fv := FreeVariables(X)
	if _, ok := fv.Get2(logic.Key(X)); !ok {
		t.Error("X should be free in X")
	}
	if fv.Len() != 1 {
		t.Errorf("expected 1 free var, got %d", fv.Len())
	}

	eq, _ := logic.NewEq(X, Y)
	fv = FreeVariables(eq)
	if fv.Len() != 2 {
		t.Errorf("expected 2 free vars in Eq(X,Y), got %d", fv.Len())
	}
}

func TestFreeVariablesForAll(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, eq)

	fv := FreeVariables(fa)
	if _, ok := fv.Get2(logic.Key(X)); ok {
		t.Error("X should be bound in ForAll X. (X == Y)")
	}
	if _, ok := fv.Get2(logic.Key(Y)); !ok {
		t.Error("Y should be free in ForAll X. (X == Y)")
	}
	if fv.Len() != 1 {
		t.Errorf("expected 1 free var, got %d", fv.Len())
	}
}

// Reproduce Python __main__ example: different-sort same-name variables
func TestFreeVariablesByIdentity(t *testing.T) {
	s1 := &logic.UninterpretedSort{Name: "s1"}
	s2 := &logic.UninterpretedSort{Name: "s2"}
	X1, _ := logic.NewVariable("X", s1)
	X2, _ := logic.NewVariable("X", s2)

	eq, _ := logic.NewEq(X2, X2)
	fa, _ := logic.NewForAll([]*logic.Variable{X1}, eq)

	// By identity: ForAll X:s1 does NOT bind X:s2
	fv := FreeVariables(fa)
	if _, ok := fv.Get2(logic.Key(X2)); !ok {
		t.Error("X:s2 should be free (bound by identity, not name)")
	}
}

func TestFreeVariablesByName(t *testing.T) {
	s1 := &logic.UninterpretedSort{Name: "s1"}
	s2 := &logic.UninterpretedSort{Name: "s2"}
	X1, _ := logic.NewVariable("X", s1)
	X2, _ := logic.NewVariable("X", s2)

	eq, _ := logic.NewEq(X2, X2)
	fa, _ := logic.NewForAll([]*logic.Variable{X1}, eq)

	// By name: ForAll X:s1 DOES bind X:s2
	fvn := FreeVariablesByName(fa)
	if _, ok := fvn["X"]; ok {
		t.Error("X should NOT be free when binding by name")
	}
	if len(fvn) != 0 {
		t.Errorf("expected 0 free names, got %d", len(fvn))
	}

	// But the body has X free by name
	fvBody := FreeVariablesByName(eq)
	if _, ok := fvBody["X"]; !ok {
		t.Error("X should be free in body")
	}
}

func TestBoundVariables(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, eq)

	bv := BoundVariables(fa)
	if _, ok := bv[logic.Key(X)]; !ok {
		t.Error("X should be bound")
	}
	if _, ok := bv[logic.Key(Y)]; ok {
		t.Error("Y should not be bound")
	}
}

func TestUsedVariables(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, eq)

	uv := UsedVariables(fa)
	if _, ok := uv[logic.Key(X)]; !ok {
		t.Error("X should be used")
	}
	if _, ok := uv[logic.Key(Y)]; !ok {
		t.Error("Y should be used")
	}
}

func TestSubstituteSimple(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	Z, _ := logic.NewVariable("Z", S)

	eq, _ := logic.NewEq(X, Y)
	subs := map[logic.NodeKey]logic.Expr{logic.Key(X): Z}

	result, err := Substitute(eq, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := logic.NewEq(Z, Y)
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSubstituteConst(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	c1 := logic.NewConst("c1", S)
	c2 := logic.NewConst("c2", S)
	X, _ := logic.NewVariable("X", S)

	eq, _ := logic.NewEq(c1, X)
	subs := map[logic.NodeKey]logic.Expr{logic.Key(c1): c2}

	result, err := Substitute(eq, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := logic.NewEq(c2, X)
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSubstituteSkipsBound(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	Z, _ := logic.NewVariable("Z", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, eq)

	// Substituting X -> Z should NOT affect bound X
	subs := map[logic.NodeKey]logic.Expr{logic.Key(X): Z}
	result, err := Substitute(fa, subs)
	if err != nil {
		t.Fatal(err)
	}
	// Body should still have X (bound), Y unchanged
	if result.String() != fa.String() {
		t.Errorf("bound variable should not be substituted: got %s", result)
	}
}

func TestSubstituteCaptureError(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, eq)

	// Substituting Y -> X would create capture (X is bound)
	subs := map[logic.NodeKey]logic.Expr{logic.Key(Y): X}
	_, err := Substitute(fa, subs)
	if err == nil {
		t.Error("expected CaptureError")
	}
	var ce *CaptureError
	if !errors.As(err, &ce) {
		t.Errorf("expected CaptureError, got %T: %v", err, err)
	}
}

func TestSubstituteEmptySubs(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)

	result, err := Substitute(X, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != X {
		t.Error("empty substitution should return same node")
	}
}

func TestIsTautologyEquality(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eqSame, _ := logic.NewEq(X, X)
	eqDiff, _ := logic.NewEq(X, Y)

	if !IsTautologyEquality(eqSame) {
		t.Error("Eq(X,X) should be tautology equality")
	}
	if IsTautologyEquality(eqDiff) {
		t.Error("Eq(X,Y) should not be tautology equality")
	}
	if IsTautologyEquality(X) {
		t.Error("non-Eq should not be tautology equality")
	}
}

func TestEqualModAlphaSimple(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	// ForAll X. X == X  should equal  ForAll Y. Y == Y
	eq1, _ := logic.NewEq(X, X)
	fa1, _ := logic.NewForAll([]*logic.Variable{X}, eq1)

	eq2, _ := logic.NewEq(Y, Y)
	fa2, _ := logic.NewForAll([]*logic.Variable{Y}, eq2)

	if !EqualModAlpha(fa1, fa2) {
		t.Error("ForAll X. X==X should be alpha-equal to ForAll Y. Y==Y")
	}
}

func TestEqualModAlphaNotEqual(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	Z, _ := logic.NewVariable("Z", S)

	// ForAll X. X == X  vs  ForAll Y. Y == Z
	eq1, _ := logic.NewEq(X, X)
	fa1, _ := logic.NewForAll([]*logic.Variable{X}, eq1)

	eq2, _ := logic.NewEq(Y, Z)
	fa2, _ := logic.NewForAll([]*logic.Variable{Y}, eq2)

	if EqualModAlpha(fa1, fa2) {
		t.Error("should not be alpha-equal")
	}
}

func TestEqualModAlphaFreeVars(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	// X == X  vs  Y == Y (free vars — not alpha-equal since different free vars)
	eq1, _ := logic.NewEq(X, X)
	eq2, _ := logic.NewEq(Y, Y)

	if EqualModAlpha(eq1, eq2) {
		t.Error("different free vars should not be alpha-equal")
	}
}

func TestSubstituteApply(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	leq := logic.NewConst("leq", mustFS(t, S, S, logic.Boolean))

	app, _ := logic.NewApply(leq, X, Y)
	subs := map[logic.NodeKey]logic.Expr{logic.Key(X): Y}

	result, err := Substitute(app, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := logic.NewApply(leq, Y, Y)
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSubstituteAnd(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	Z, _ := logic.NewVariable("Z", S)

	eq1, _ := logic.NewEq(X, Y)
	eq2, _ := logic.NewEq(Y, X)
	and, _ := logic.NewAnd(eq1, eq2)

	subs := map[logic.NodeKey]logic.Expr{logic.Key(X): Z}
	result, err := Substitute(and, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected1, _ := logic.NewEq(Z, Y)
	expected2, _ := logic.NewEq(Y, Z)
	expectedAnd, _ := logic.NewAnd(expected1, expected2)
	if !result.Equal(expectedAnd) {
		t.Errorf("expected %s, got %s", expectedAnd, result)
	}
}

func TestFreeVariablesNested(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	Z, _ := logic.NewVariable("Z", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, eq)
	and, _ := logic.NewAnd(fa)

	// Z is not mentioned at all
	eq2, _ := logic.NewEq(Z, Z)
	or, _ := logic.NewOr(and, eq2)

	fv := FreeVariables(or)
	if _, ok := fv.Get2(logic.Key(X)); ok {
		t.Error("X should not be free (bound in ForAll)")
	}
	if _, ok := fv.Get2(logic.Key(Y)); !ok {
		t.Error("Y should be free")
	}
	if _, ok := fv.Get2(logic.Key(Z)); !ok {
		t.Error("Z should be free")
	}
}

func TestFreeVariablesLambda(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	lam, _ := logic.NewLambda([]*logic.Variable{X}, eq)

	fv := FreeVariables(lam)
	if _, ok := fv.Get2(logic.Key(X)); ok {
		t.Error("X should be bound in Lambda")
	}
	if _, ok := fv.Get2(logic.Key(Y)); !ok {
		t.Error("Y should be free in Lambda body")
	}
}

func TestFreeVariablesNamedBinder(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	nb, _ := logic.NewNamedBinder("nb", []*logic.Variable{X}, nil, eq)

	fv := FreeVariables(nb)
	if _, ok := fv.Get2(logic.Key(X)); ok {
		t.Error("X should be bound in NamedBinder")
	}
	if _, ok := fv.Get2(logic.Key(Y)); !ok {
		t.Error("Y should be free in NamedBinder body")
	}
}

func TestEqualModAlphaExists(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq1, _ := logic.NewEq(X, X)
	ex1, _ := logic.NewExists([]*logic.Variable{X}, eq1)

	eq2, _ := logic.NewEq(Y, Y)
	ex2, _ := logic.NewExists([]*logic.Variable{Y}, eq2)

	if !EqualModAlpha(ex1, ex2) {
		t.Error("Exists X. X==X should be alpha-equal to Exists Y. Y==Y")
	}
}

// === NormalizeNamedBinders tests ===

// TestNNB_Constant: Const is returned unchanged (branch 1).
func TestNNB_Constant(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	c := logic.NewConst("c", S)
	result := NormalizeNamedBinders(c, nil)
	if result != c {
		t.Error("Const should be returned as-is (same pointer)")
	}
}

// TestNNB_ConstantWithFuncSort: Const with FunctionSort is still returned unchanged.
func TestNNB_ConstantWithFuncSort(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, S, logic.Boolean)
	c := logic.NewConst("f", fs)
	result := NormalizeNamedBinders(c, nil)
	if result != c {
		t.Error("Const with FunctionSort should be returned as-is")
	}
}

// TestNNB_Variable: Variable goes through general clone path (branch 4).
func TestNNB_Variable(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	result := NormalizeNamedBinders(X, nil)
	if !result.(logic.Expr).Equal(X) {
		t.Errorf("Variable should be structurally equal after normalization, got %s", result)
	}
}

// TestNNB_SimpleNamedBinder: NamedBinder with 1 variable normalized to V0 (branch 2).
func TestNNB_SimpleNamedBinder(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	eq, _ := logic.NewEq(X, X)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{X}, nil, eq)

	result := NormalizeNamedBinders(nb, nil)
	rnb, ok := result.(*logic.NamedBinder)
	if !ok {
		t.Fatalf("expected *NamedBinder, got %T", result)
	}
	if rnb.Name != "g" {
		t.Errorf("expected name 'g', got '%s'", rnb.Name)
	}
	if len(rnb.Variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(rnb.Variables))
	}
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected variable V0, got %s", rnb.Variables[0].Name)
	}
	// Body should have V0 substituted for X
	body, ok := rnb.Body.(*logic.Eq)
	if !ok {
		t.Fatalf("expected *Eq body, got %T", rnb.Body)
	}
	v0t1, ok := body.T1.(*logic.Variable)
	if !ok || v0t1.Name != "V0" {
		t.Errorf("expected V0 in Eq.T1, got %s", body.T1)
	}
	v0t2, ok := body.T2.(*logic.Variable)
	if !ok || v0t2.Name != "V0" {
		t.Errorf("expected V0 in Eq.T2, got %s", body.T2)
	}
}

// TestNNB_MultiVarNamedBinder: NamedBinder with 2 variables → V0, V1 (branch 2).
func TestNNB_MultiVarNamedBinder(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	eq, _ := logic.NewEq(X, Y)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{X, Y}, nil, eq)

	result := NormalizeNamedBinders(nb, nil)
	rnb := result.(*logic.NamedBinder)
	if len(rnb.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(rnb.Variables))
	}
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0, got %s", rnb.Variables[0].Name)
	}
	if rnb.Variables[1].Name != "V1" {
		t.Errorf("expected V1, got %s", rnb.Variables[1].Name)
	}
	body := rnb.Body.(*logic.Eq)
	if body.T1.(*logic.Variable).Name != "V0" {
		t.Errorf("expected V0 in T1, got %s", body.T1)
	}
	if body.T2.(*logic.Variable).Name != "V1" {
		t.Errorf("expected V1 in T2, got %s", body.T2)
	}
}

// TestNNB_NamedBinderNotInNames: NamedBinder with name NOT in names set.
// Body is recursed but variables are unchanged (branch 2 skip → branch 4).
func TestNNB_NamedBinderNotInNames(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	eq, _ := logic.NewEq(X, X)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{X}, nil, eq)

	// Only normalize binders named "h", not "g"
	names := map[string]bool{"h": true}
	result := NormalizeNamedBinders(nb, names)
	rnb := result.(*logic.NamedBinder)
	// Variables should NOT be renamed (name not in set)
	if rnb.Variables[0].Name != "X" {
		t.Errorf("expected X unchanged, got %s", rnb.Variables[0].Name)
	}
}

// TestNNB_NamesNil: All NamedBinders normalized when names=nil (branch 2).
func TestNNB_NamesNil(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	eqX, _ := logic.NewEq(X, X)
	eqY, _ := logic.NewEq(Y, Y)
	nbG, _ := logic.NewNamedBinder("g", []*logic.Variable{X}, nil, eqX)
	nbH, _ := logic.NewNamedBinder("h", []*logic.Variable{Y}, nil, eqY)

	// Test each NamedBinder independently (can't put FunctionSort binders in And)
	rG := NormalizeNamedBinders(nbG, nil).(*logic.NamedBinder)
	rH := NormalizeNamedBinders(nbH, nil).(*logic.NamedBinder)
	if rG.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in g binder, got %s", rG.Variables[0].Name)
	}
	if rH.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in h binder, got %s", rH.Variables[0].Name)
	}
}

// TestNNB_ApplyResult: Apply node result is structurally correct (branch 3).
func TestNNB_ApplyResult(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, S, logic.Boolean)
	f := logic.NewConst("f", fs)
	X, _ := logic.NewVariable("X", S)
	app := logic.MustApply(f, X)

	result := NormalizeNamedBinders(app, nil)
	rApp, ok := result.(*logic.Apply)
	if !ok {
		t.Fatalf("expected *Apply, got %T", result)
	}
	// Func should be the same Const
	rFunc, ok := rApp.Func.(*logic.Const)
	if !ok || rFunc.Name != "f" {
		t.Errorf("expected Const 'f', got %v", rApp.Func)
	}
	// Term should be Variable X
	rTerm, ok := rApp.Terms[0].(*logic.Variable)
	if !ok || rTerm.Name != "X" {
		t.Errorf("expected Variable X, got %v", rApp.Terms[0])
	}
}

// TestNNB_ApplyMultipleTerms: Apply with multiple terms all get recursed (branch 3).
func TestNNB_ApplyMultipleTerms(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, S, S, logic.Boolean)
	f := logic.NewConst("f", fs)
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	app := logic.MustApply(f, X, Y)

	result := NormalizeNamedBinders(app, nil)
	rApp := result.(*logic.Apply)
	if rApp.Func.(*logic.Const).Name != "f" {
		t.Errorf("expected Const 'f', got %v", rApp.Func)
	}
	if len(rApp.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(rApp.Terms))
	}
	if rApp.Terms[0].(*logic.Variable).Name != "X" {
		t.Errorf("expected X, got %s", rApp.Terms[0])
	}
	if rApp.Terms[1].(*logic.Variable).Name != "Y" {
		t.Errorf("expected Y, got %s", rApp.Terms[1])
	}
}

// TestNNB_NestedApplyInFormula: Apply inside And (branch 4 → branch 3).
func TestNNB_NestedApplyInFormula(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, S, logic.Boolean)
	f := logic.NewConst("f", fs)
	X, _ := logic.NewVariable("X", S)
	app := logic.MustApply(f, X)
	and, _ := logic.NewAnd(app)

	result := NormalizeNamedBinders(and, nil)
	rAnd := result.(*logic.And)
	rApp := rAnd.Terms[0].(*logic.Apply)
	if rApp.Func.(*logic.Const).Name != "f" {
		t.Errorf("expected Const 'f' in nested Apply, got %v", rApp.Func)
	}
}

// TestNNB_NamedBinderAsApplyFunc: Apply where Func is a NamedBinder (branch 3).
func TestNNB_NamedBinderAsApplyFunc(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	// NamedBinder g(X) = X : S -> S
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{X}, nil, X)
	// Apply g to Y
	app := logic.MustApply(nb, Y)

	result := NormalizeNamedBinders(app, nil)
	rApp := result.(*logic.Apply)
	rnb := rApp.Func.(*logic.NamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in Apply.Func NamedBinder, got %s", rnb.Variables[0].Name)
	}
	// Body of the NamedBinder should have V0 substituted for X
	rBody := rnb.Body.(*logic.Variable)
	if rBody.Name != "V0" {
		t.Errorf("expected V0 in NamedBinder body, got %s", rBody.Name)
	}
	// The term Y should be unchanged
	rTerm := rApp.Terms[0].(*logic.Variable)
	if rTerm.Name != "Y" {
		t.Errorf("expected Y in Apply term, got %s", rTerm.Name)
	}
}

// TestNNB_ForAllBody: ForAll body recursed, variables preserved (branch 4).
func TestNNB_ForAllBody(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	// nb = g(Y) { Eq(X, Y) } : FunctionSort(S, Boolean)
	eq, _ := logic.NewEq(X, Y)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{Y}, nil, eq)
	// Apply nb to X → Boolean
	app := logic.MustApply(nb, X)
	// ForAll X. nb(X)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, app)

	result := NormalizeNamedBinders(fa, nil)
	rFa := result.(*logic.ForAll)
	// ForAll variables should be preserved
	if rFa.Variables[0].Name != "X" {
		t.Errorf("expected ForAll variable X preserved, got %s", rFa.Variables[0].Name)
	}
	// Apply inside ForAll should have NamedBinder as Func
	rApp := rFa.Body.(*logic.Apply)
	rnb := rApp.Func.(*logic.NamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in inner NamedBinder, got %s", rnb.Variables[0].Name)
	}
}

// TestNNB_NotAndOrImplies: Simple formulas have children recursed (branch 4).
func TestNNB_NotAndOrImplies(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	eq, _ := logic.NewEq(X, Y)
	// nb = g(X) { Eq(X, Y) } : FunctionSort(S, Boolean)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{X}, nil, eq)
	// Apply nb to Y → BooleanSort
	app := logic.MustApply(nb, Y)

	// Not
	notApp, _ := logic.NewNot(app)
	rNot := NormalizeNamedBinders(notApp, nil).(*logic.Not)
	rNotApp := rNot.Body.(*logic.Apply)
	if rNotApp.Func.(*logic.NamedBinder).Variables[0].Name != "V0" {
		t.Error("Not: inner NamedBinder not normalized")
	}

	// And
	andApp, _ := logic.NewAnd(app)
	rAnd := NormalizeNamedBinders(andApp, nil).(*logic.And)
	rAndApp := rAnd.Terms[0].(*logic.Apply)
	if rAndApp.Func.(*logic.NamedBinder).Variables[0].Name != "V0" {
		t.Error("And: inner NamedBinder not normalized")
	}

	// Or
	orApp, _ := logic.NewOr(app)
	rOr := NormalizeNamedBinders(orApp, nil).(*logic.Or)
	rOrApp := rOr.Terms[0].(*logic.Apply)
	if rOrApp.Func.(*logic.NamedBinder).Variables[0].Name != "V0" {
		t.Error("Or: inner NamedBinder not normalized")
	}

	// Implies
	implApp, _ := logic.NewImplies(app, app)
	rImpl := NormalizeNamedBinders(implApp, nil).(*logic.Implies)
	rImplT1 := rImpl.T1.(*logic.Apply)
	rImplT2 := rImpl.T2.(*logic.Apply)
	if rImplT1.Func.(*logic.NamedBinder).Variables[0].Name != "V0" {
		t.Error("Implies.T1: inner NamedBinder not normalized")
	}
	if rImplT2.Func.(*logic.NamedBinder).Variables[0].Name != "V0" {
		t.Error("Implies.T2: inner NamedBinder not normalized")
	}
}

// TestNNB_DeepNesting: NamedBinder inside Apply inside And (all branches).
func TestNNB_DeepNesting(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	// nb = g(X) { Eq(X,X) } : FunctionSort(S, Boolean)
	eqXX, _ := logic.NewEq(X, X)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{X}, nil, eqXX)
	// app = nb(Y) → BooleanSort
	app := logic.MustApply(nb, Y)
	// and = And(app) → BooleanSort
	and, _ := logic.NewAnd(app)

	result := NormalizeNamedBinders(and, nil)
	rAnd := result.(*logic.And)
	rApp := rAnd.Terms[0].(*logic.Apply)
	rnb := rApp.Func.(*logic.NamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("deep nesting: expected V0, got %s", rnb.Variables[0].Name)
	}
	// Body should have V0 substituted for X in Eq(X,X) → Eq(V0,V0)
	rBody := rnb.Body.(*logic.Eq)
	if rBody.T1.(*logic.Variable).Name != "V0" {
		t.Errorf("deep nesting: expected V0 in body Eq.T1, got %s", rBody.T1)
	}
}

// TestNNB_Idempotent: Running NormalizeNamedBinders twice gives same result.
func TestNNB_Idempotent(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	eq, _ := logic.NewEq(X, Y)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{X, Y}, nil, eq)

	r1 := NormalizeNamedBinders(nb, nil)
	r2 := NormalizeNamedBinders(r1, nil)

	rnb1 := r1.(*logic.NamedBinder)
	rnb2 := r2.(*logic.NamedBinder)
	if rnb1.Variables[0].Name != rnb2.Variables[0].Name ||
		rnb1.Variables[1].Name != rnb2.Variables[1].Name {
		t.Errorf("not idempotent: first=%v second=%v",
			rnb1.Variables, rnb2.Variables)
	}
	if !rnb1.Body.Equal(rnb2.Body) {
		t.Error("not idempotent: bodies differ")
	}
}

// TestNNB_ExistsBody: Exists body recursed, variables preserved (branch 4).
func TestNNB_ExistsBody(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	// nb = g(Y) { Eq(X, Y) } : FunctionSort(S, Boolean)
	eq, _ := logic.NewEq(X, Y)
	nb, _ := logic.NewNamedBinder("g", []*logic.Variable{Y}, nil, eq)
	// Apply nb to X → Boolean
	app := logic.MustApply(nb, X)
	// Exists X. nb(X)
	ex, _ := logic.NewExists([]*logic.Variable{X}, app)

	result := NormalizeNamedBinders(ex, nil)
	rEx := result.(*logic.Exists)
	if rEx.Variables[0].Name != "X" {
		t.Errorf("expected Exists variable X preserved, got %s", rEx.Variables[0].Name)
	}
	rApp := rEx.Body.(*logic.Apply)
	rnb := rApp.Func.(*logic.NamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in inner NamedBinder, got %s", rnb.Variables[0].Name)
	}
}

// TestNNB_Eq: Eq children recursed (branch 4).
func TestNNB_Eq(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	result := NormalizeNamedBinders(mustEq(t, X, Y), nil)
	rEq := result.(*logic.Eq)
	if rEq.T1.(*logic.Variable).Name != "X" {
		t.Errorf("expected X in Eq.T1, got %s", rEq.T1)
	}
	if rEq.T2.(*logic.Variable).Name != "Y" {
		t.Errorf("expected Y in Eq.T2, got %s", rEq.T2)
	}
}

func mustEq(t *testing.T, a, b logic.Expr) *logic.Eq {
	t.Helper()
	e, err := logic.NewEq(a, b)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func FuzzSubstitute(f *testing.F) {
	f.Add(true, true)
	f.Add(true, false)
	f.Add(false, false)
	f.Fuzz(func(t *testing.T, subX, subY bool) {
		S := &logic.UninterpretedSort{Name: "S"}
		X, _ := logic.NewVariable("X", S)
		Y, _ := logic.NewVariable("Y", S)
		Z, _ := logic.NewVariable("Z", S)
		eq, _ := logic.NewEq(X, Y)

		subs := map[logic.NodeKey]logic.Expr{}
		if subX {
			subs[logic.Key(X)] = Z
		}
		if subY {
			subs[logic.Key(Y)] = Z
		}
		// Should not panic
		result, err := Substitute(eq, subs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = result.String()
	})
}
