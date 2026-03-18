package logicutil

import (
	"errors"
	"testing"

	"github.com/glycerine/goivy/logic"
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
	if _, ok := fv[logic.Key(X)]; !ok {
		t.Error("X should be free in X")
	}
	if len(fv) != 1 {
		t.Errorf("expected 1 free var, got %d", len(fv))
	}

	eq, _ := logic.NewEq(X, Y)
	fv = FreeVariables(eq)
	if len(fv) != 2 {
		t.Errorf("expected 2 free vars in Eq(X,Y), got %d", len(fv))
	}
}

func TestFreeVariablesForAll(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Variable{X}, eq)

	fv := FreeVariables(fa)
	if _, ok := fv[logic.Key(X)]; ok {
		t.Error("X should be bound in ForAll X. (X == Y)")
	}
	if _, ok := fv[logic.Key(Y)]; !ok {
		t.Error("Y should be free in ForAll X. (X == Y)")
	}
	if len(fv) != 1 {
		t.Errorf("expected 1 free var, got %d", len(fv))
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
	if _, ok := fv[logic.Key(X2)]; !ok {
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

func TestUsedConstants(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	c := logic.NewSymbol("c", S)
	leq := logic.NewSymbol("leq", mustFS(t, S, S, logic.Boolean))

	// In Python, constants_ast does NOT yield Apply function heads —
	// only standalone constants. So leq(X, c) yields {c} not {leq, c}.
	app, _ := logic.NewApply(leq, X, c)
	uc := UsedConstants(app)
	if _, ok := uc[logic.Key(c)]; !ok {
		t.Error("c should be found as a constant argument")
	}
	if _, ok := uc[logic.Key(leq)]; ok {
		t.Error("leq should NOT be found (it's a function head, not a standalone constant)")
	}
	if len(uc) != 1 {
		t.Errorf("expected 1 constant, got %d", len(uc))
	}

	// Standalone constant
	uc2 := UsedConstants(c)
	if _, ok := uc2[logic.Key(c)]; !ok {
		t.Error("standalone c should be found")
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
	c1 := logic.NewSymbol("c1", S)
	c2 := logic.NewSymbol("c2", S)
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
	leq := logic.NewSymbol("leq", mustFS(t, S, S, logic.Boolean))

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
	if _, ok := fv[logic.Key(X)]; ok {
		t.Error("X should not be free (bound in ForAll)")
	}
	if _, ok := fv[logic.Key(Y)]; !ok {
		t.Error("Y should be free")
	}
	if _, ok := fv[logic.Key(Z)]; !ok {
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
	if _, ok := fv[logic.Key(X)]; ok {
		t.Error("X should be bound in Lambda")
	}
	if _, ok := fv[logic.Key(Y)]; !ok {
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
	if _, ok := fv[logic.Key(X)]; ok {
		t.Error("X should be bound in NamedBinder")
	}
	if _, ok := fv[logic.Key(Y)]; !ok {
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
