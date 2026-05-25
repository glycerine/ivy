package goivy

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

func logicutilMustFS(t *testing.T, sorts ...Sort) *LogicFunctionSort {
	t.Helper()
	fs, err := NewFunctionSort(sorts...)
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

func captureLogicutilStdout(t *testing.T, fn func()) string {
	t.Helper()
	if !xtracer.Enabled {
		t.Skip("xtrace disabled at build time")
	}

	oldStdout := os.Stdout
	oldSuppressed := xtracer.Suppressed
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCh <- buf.String()
	}()

	os.Stdout = w
	xtracer.Suppressed = false
	closedW := false
	defer func() {
		os.Stdout = oldStdout
		xtracer.Suppressed = oldSuppressed
		if !closedW {
			_ = w.Close()
		}
		_ = r.Close()
	}()

	fn()
	_ = w.Close()
	closedW = true
	return <-outCh
}

func TestReduceNamedBindersReducesApplyArgsBeforeNormalizingFunc(t *testing.T) {
	client := &UninterpretedSort{Name: "client"}
	cacheState := &LogicEnumeratedSort{Name: "cachestate", Extension: []string{"invalid", "shared", "exclusive"}}
	message := &LogicEnumeratedSort{Name: "message2_4", Extension: []string{"empty2_4", "invalidate", "grantshared", "grantexclusive"}}
	cache := NewConst("s.cache", logicutilMustFS(t, client, cacheState))
	channel := NewConst("s.channel2_4", logicutilMustFS(t, cacheState, message))
	c := NewConst("c", client)

	outer := MustApply(channel, MustApply(cache, c))
	out := captureLogicutilStdout(t, func() {
		_ = ReduceNamedBinders(outer, nil)
	})

	cacheIdx := strings.Index(out, "name:s.cache")
	channelIdx := strings.Index(out, "name:s.channel2_4")
	if cacheIdx < 0 || channelIdx < 0 {
		t.Fatalf("expected both function symbols in xtrace, got:\n%s", out)
	}
	if channelIdx < cacheIdx {
		t.Fatalf("ReduceNamedBinders normalized apply function before reducing args:\n%s", out)
	}
}

func TestFreeVariablesSimple(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	fv := FreeVariables(X)
	if _, ok := fv.Get2(Key(X)); !ok {
		t.Error("X should be free in X")
	}
	if fv.Len() != 1 {
		t.Errorf("expected 1 free var, got %d", fv.Len())
	}

	eq, _ := NewEq(X, Y)
	fv = FreeVariables(eq)
	if fv.Len() != 2 {
		t.Errorf("expected 2 free vars in Eq(X,Y), got %d", fv.Len())
	}
}

func TestFreeVariablesForAll(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	fa, _ := NewForAll([]*LogicVariable{X}, eq)

	fv := FreeVariables(fa)
	if _, ok := fv.Get2(Key(X)); ok {
		t.Error("X should be bound in ForAll X. (X == Y)")
	}
	if _, ok := fv.Get2(Key(Y)); !ok {
		t.Error("Y should be free in ForAll X. (X == Y)")
	}
	if fv.Len() != 1 {
		t.Errorf("expected 1 free var, got %d", fv.Len())
	}
}

// Reproduce Python __main__ example: different-sort same-name variables
func TestFreeVariablesByIdentity(t *testing.T) {
	s1 := &UninterpretedSort{Name: "s1"}
	s2 := &UninterpretedSort{Name: "s2"}
	X1, _ := NewVariable("X", s1)
	X2, _ := NewVariable("X", s2)

	eq, _ := NewEq(X2, X2)
	fa, _ := NewForAll([]*LogicVariable{X1}, eq)

	// By identity: ForAll X:s1 does NOT bind X:s2
	fv := FreeVariables(fa)
	if _, ok := fv.Get2(Key(X2)); !ok {
		t.Error("X:s2 should be free (bound by identity, not name)")
	}
}

func TestFreeVariablesByName(t *testing.T) {
	s1 := &UninterpretedSort{Name: "s1"}
	s2 := &UninterpretedSort{Name: "s2"}
	X1, _ := NewVariable("X", s1)
	X2, _ := NewVariable("X", s2)

	eq, _ := NewEq(X2, X2)
	fa, _ := NewForAll([]*LogicVariable{X1}, eq)

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
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	fa, _ := NewForAll([]*LogicVariable{X}, eq)

	bv := BoundVariables(fa)
	if _, ok := bv[Key(X)]; !ok {
		t.Error("X should be bound")
	}
	if _, ok := bv[Key(Y)]; ok {
		t.Error("Y should not be bound")
	}
}

func TestUsedVariables(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	fa, _ := NewForAll([]*LogicVariable{X}, eq)

	uv := UsedVariables(fa)
	if _, ok := uv[Key(X)]; !ok {
		t.Error("X should be used")
	}
	if _, ok := uv[Key(Y)]; !ok {
		t.Error("Y should be used")
	}
}

func TestSubstituteSimple(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)

	eq, _ := NewEq(X, Y)
	subs := map[NodeKey]Expr{Key(X): Z}

	result, err := Substitute(eq, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := NewEq(Z, Y)
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSubstituteConst(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	c1 := NewConst("c1", S)
	c2 := NewConst("c2", S)
	X, _ := NewVariable("X", S)

	eq, _ := NewEq(c1, X)
	subs := map[NodeKey]Expr{Key(c1): c2}

	result, err := Substitute(eq, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := NewEq(c2, X)
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSubstituteSkipsBound(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)

	eq, _ := NewEq(X, Y)
	fa, _ := NewForAll([]*LogicVariable{X}, eq)

	// Substituting X -> Z should NOT affect bound X
	subs := map[NodeKey]Expr{Key(X): Z}
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
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	fa, _ := NewForAll([]*LogicVariable{X}, eq)

	// Substituting Y -> X would create capture (X is bound)
	subs := map[NodeKey]Expr{Key(Y): X}
	_, err := Substitute(fa, subs)
	if err == nil {
		t.Error("expected CaptureError")
	}
	var ce *LogicUtilCaptureError
	if !errors.As(err, &ce) {
		t.Errorf("expected CaptureError, got %T: %v", err, err)
	}
}

func TestSubstituteEmptySubs(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)

	result, err := Substitute(X, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != X {
		t.Error("empty substitution should return same node")
	}
}

func TestIsTautologyEquality(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eqSame, _ := NewEq(X, X)
	eqDiff, _ := NewEq(X, Y)

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
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	// ForAll X. X == X  should equal  ForAll Y. Y == Y
	eq1, _ := NewEq(X, X)
	fa1, _ := NewForAll([]*LogicVariable{X}, eq1)

	eq2, _ := NewEq(Y, Y)
	fa2, _ := NewForAll([]*LogicVariable{Y}, eq2)

	if !EqualModAlpha(fa1, fa2) {
		t.Error("ForAll X. X==X should be alpha-equal to ForAll Y. Y==Y")
	}
}

func TestEqualModAlphaNotEqual(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)

	// ForAll X. X == X  vs  ForAll Y. Y == Z
	eq1, _ := NewEq(X, X)
	fa1, _ := NewForAll([]*LogicVariable{X}, eq1)

	eq2, _ := NewEq(Y, Z)
	fa2, _ := NewForAll([]*LogicVariable{Y}, eq2)

	if EqualModAlpha(fa1, fa2) {
		t.Error("should not be alpha-equal")
	}
}

func TestEqualModAlphaFreeVars(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	// X == X  vs  Y == Y (free vars — not alpha-equal since different free vars)
	eq1, _ := NewEq(X, X)
	eq2, _ := NewEq(Y, Y)

	if EqualModAlpha(eq1, eq2) {
		t.Error("different free vars should not be alpha-equal")
	}
}

func TestSubstituteApply(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewConst("leq", logicutilMustFS(t, S, S, Boolean))

	app, _ := NewApply(leq, X, Y)
	subs := map[NodeKey]Expr{Key(X): Y}

	result, err := Substitute(app, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := NewApply(leq, Y, Y)
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSubstituteAnd(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)

	eq1, _ := NewEq(X, Y)
	eq2, _ := NewEq(Y, X)
	and, _ := NewAnd(eq1, eq2)

	subs := map[NodeKey]Expr{Key(X): Z}
	result, err := Substitute(and, subs)
	if err != nil {
		t.Fatal(err)
	}
	expected1, _ := NewEq(Z, Y)
	expected2, _ := NewEq(Y, Z)
	expectedAnd, _ := NewAnd(expected1, expected2)
	if !result.Equal(expectedAnd) {
		t.Errorf("expected %s, got %s", expectedAnd, result)
	}
}

func TestFreeVariablesNested(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)

	eq, _ := NewEq(X, Y)
	fa, _ := NewForAll([]*LogicVariable{X}, eq)
	and, _ := NewAnd(fa)

	// Z is not mentioned at all
	eq2, _ := NewEq(Z, Z)
	or, _ := NewOr(and, eq2)

	fv := FreeVariables(or)
	if _, ok := fv.Get2(Key(X)); ok {
		t.Error("X should not be free (bound in ForAll)")
	}
	if _, ok := fv.Get2(Key(Y)); !ok {
		t.Error("Y should be free")
	}
	if _, ok := fv.Get2(Key(Z)); !ok {
		t.Error("Z should be free")
	}
}

func TestFreeVariablesLambda(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	lam, _ := NewLambda([]*LogicVariable{X}, eq)

	fv := FreeVariables(lam)
	if _, ok := fv.Get2(Key(X)); ok {
		t.Error("X should be bound in Lambda")
	}
	if _, ok := fv.Get2(Key(Y)); !ok {
		t.Error("Y should be free in Lambda body")
	}
}

func TestFreeVariablesNamedBinder(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	nb, _ := NewNamedBinder("nb", []*LogicVariable{X}, nil, eq)

	fv := FreeVariables(nb)
	if _, ok := fv.Get2(Key(X)); ok {
		t.Error("X should be bound in NamedBinder")
	}
	if _, ok := fv.Get2(Key(Y)); !ok {
		t.Error("Y should be free in NamedBinder body")
	}
}

func TestEqualModAlphaExists(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq1, _ := NewEq(X, X)
	ex1, _ := NewExists([]*LogicVariable{X}, eq1)

	eq2, _ := NewEq(Y, Y)
	ex2, _ := NewExists([]*LogicVariable{Y}, eq2)

	if !EqualModAlpha(ex1, ex2) {
		t.Error("Exists X. X==X should be alpha-equal to Exists Y. Y==Y")
	}
}

// === NormalizeNamedBinders tests ===

// TestNNB_Constant: Const is returned unchanged (branch 1).
func TestNNB_Constant(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	c := NewConst("c", S)
	result := NormalizeNamedBinders(c, nil)
	if result != c {
		t.Error("Const should be returned as-is (same pointer)")
	}
}

// TestNNB_ConstantWithFuncSort: Const with FunctionSort is still returned unchanged.
func TestNNB_ConstantWithFuncSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs := logicutilMustFS(t, S, Boolean)
	c := NewConst("f", fs)
	result := NormalizeNamedBinders(c, nil)
	if result != c {
		t.Error("Const with FunctionSort should be returned as-is")
	}
}

// TestNNB_Variable: Variable goes through general clone path (branch 4).
func TestNNB_Variable(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	result := NormalizeNamedBinders(X, nil)
	if !result.(Expr).Equal(X) {
		t.Errorf("Variable should be structurally equal after normalization, got %s", result)
	}
}

// TestNNB_SimpleNamedBinder: NamedBinder with 1 variable normalized to V0 (branch 2).
func TestNNB_SimpleNamedBinder(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	nb, _ := NewNamedBinder("g", []*LogicVariable{X}, nil, eq)

	result := NormalizeNamedBinders(nb, nil)
	rnb, ok := result.(*LogicNamedBinder)
	if !ok {
		t.Fatalf("expected *LogicNamedBinder, got %T", result)
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
	body, ok := rnb.Body.(*Eq)
	if !ok {
		t.Fatalf("expected *Eq body, got %T", rnb.Body)
	}
	v0t1, ok := body.T1.(*LogicVariable)
	if !ok || v0t1.Name != "V0" {
		t.Errorf("expected V0 in Eq.T1, got %s", body.T1)
	}
	v0t2, ok := body.T2.(*LogicVariable)
	if !ok || v0t2.Name != "V0" {
		t.Errorf("expected V0 in Eq.T2, got %s", body.T2)
	}
}

// TestNNB_MultiVarNamedBinder: NamedBinder with 2 variables → V0, V1 (branch 2).
func TestNNB_MultiVarNamedBinder(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	eq, _ := NewEq(X, Y)
	nb, _ := NewNamedBinder("g", []*LogicVariable{X, Y}, nil, eq)

	result := NormalizeNamedBinders(nb, nil)
	rnb := result.(*LogicNamedBinder)
	if len(rnb.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(rnb.Variables))
	}
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0, got %s", rnb.Variables[0].Name)
	}
	if rnb.Variables[1].Name != "V1" {
		t.Errorf("expected V1, got %s", rnb.Variables[1].Name)
	}
	body := rnb.Body.(*Eq)
	if body.T1.(*LogicVariable).Name != "V0" {
		t.Errorf("expected V0 in T1, got %s", body.T1)
	}
	if body.T2.(*LogicVariable).Name != "V1" {
		t.Errorf("expected V1 in T2, got %s", body.T2)
	}
}

// TestNNB_NamedBinderNotInNames: NamedBinder with name NOT in names set.
// Body is recursed but variables are unchanged (branch 2 skip → branch 4).
func TestNNB_NamedBinderNotInNames(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	nb, _ := NewNamedBinder("g", []*LogicVariable{X}, nil, eq)

	// Only normalize binders named "h", not "g"
	names := map[string]bool{"h": true}
	result := NormalizeNamedBinders(nb, names)
	rnb := result.(*LogicNamedBinder)
	// Variables should NOT be renamed (name not in set)
	if rnb.Variables[0].Name != "X" {
		t.Errorf("expected X unchanged, got %s", rnb.Variables[0].Name)
	}
}

// TestNNB_NamesNil: All NamedBinders normalized when names=nil (branch 2).
func TestNNB_NamesNil(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	eqX, _ := NewEq(X, X)
	eqY, _ := NewEq(Y, Y)
	nbG, _ := NewNamedBinder("g", []*LogicVariable{X}, nil, eqX)
	nbH, _ := NewNamedBinder("h", []*LogicVariable{Y}, nil, eqY)

	// Test each NamedBinder independently (can't put FunctionSort binders in And)
	rG := NormalizeNamedBinders(nbG, nil).(*LogicNamedBinder)
	rH := NormalizeNamedBinders(nbH, nil).(*LogicNamedBinder)
	if rG.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in g binder, got %s", rG.Variables[0].Name)
	}
	if rH.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in h binder, got %s", rH.Variables[0].Name)
	}
}

// TestNNB_ApplyResult: Apply node result is structurally correct (branch 3).
func TestNNB_ApplyResult(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs := logicutilMustFS(t, S, Boolean)
	f := NewConst("f", fs)
	X, _ := NewVariable("X", S)
	app := MustApply(f, X)

	result := NormalizeNamedBinders(app, nil)
	rApp, ok := result.(*Apply)
	if !ok {
		t.Fatalf("expected *Apply, got %T", result)
	}
	// Func should be the same Const
	rFunc, ok := rApp.Func.(*Const)
	if !ok || rFunc.Name != "f" {
		t.Errorf("expected Const 'f', got %v", rApp.Func)
	}
	// Term should be Variable X
	rTerm, ok := rApp.Terms[0].(*LogicVariable)
	if !ok || rTerm.Name != "X" {
		t.Errorf("expected Variable X, got %v", rApp.Terms[0])
	}
}

// TestNNB_ApplyMultipleTerms: Apply with multiple terms all get recursed (branch 3).
func TestNNB_ApplyMultipleTerms(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs := logicutilMustFS(t, S, S, Boolean)
	f := NewConst("f", fs)
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	app := MustApply(f, X, Y)

	result := NormalizeNamedBinders(app, nil)
	rApp := result.(*Apply)
	if rApp.Func.(*Const).Name != "f" {
		t.Errorf("expected Const 'f', got %v", rApp.Func)
	}
	if len(rApp.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(rApp.Terms))
	}
	if rApp.Terms[0].(*LogicVariable).Name != "X" {
		t.Errorf("expected X, got %s", rApp.Terms[0])
	}
	if rApp.Terms[1].(*LogicVariable).Name != "Y" {
		t.Errorf("expected Y, got %s", rApp.Terms[1])
	}
}

// TestNNB_NestedApplyInFormula: Apply inside And (branch 4 → branch 3).
func TestNNB_NestedApplyInFormula(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs := logicutilMustFS(t, S, Boolean)
	f := NewConst("f", fs)
	X, _ := NewVariable("X", S)
	app := MustApply(f, X)
	and, _ := NewAnd(app)

	result := NormalizeNamedBinders(and, nil)
	rAnd := result.(*LogicAnd)
	rApp := rAnd.Terms[0].(*Apply)
	if rApp.Func.(*Const).Name != "f" {
		t.Errorf("expected Const 'f' in nested Apply, got %v", rApp.Func)
	}
}

// TestNNB_NamedBinderAsApplyFunc: Apply where Func is a NamedBinder (branch 3).
func TestNNB_NamedBinderAsApplyFunc(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	// NamedBinder g(X) = X : S -> S
	nb, _ := NewNamedBinder("g", []*LogicVariable{X}, nil, X)
	// Apply g to Y
	app := MustApply(nb, Y)

	result := NormalizeNamedBinders(app, nil)
	rApp := result.(*Apply)
	rnb := rApp.Func.(*LogicNamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in Apply.Func NamedBinder, got %s", rnb.Variables[0].Name)
	}
	// Body of the NamedBinder should have V0 substituted for X
	rBody := rnb.Body.(*LogicVariable)
	if rBody.Name != "V0" {
		t.Errorf("expected V0 in NamedBinder body, got %s", rBody.Name)
	}
	// The term Y should be unchanged
	rTerm := rApp.Terms[0].(*LogicVariable)
	if rTerm.Name != "Y" {
		t.Errorf("expected Y in Apply term, got %s", rTerm.Name)
	}
}

// TestNNB_ForAllBody: ForAll body recursed, variables preserved (branch 4).
func TestNNB_ForAllBody(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	// nb = g(Y) { Eq(X, Y) } : FunctionSort(S, Boolean)
	eq, _ := NewEq(X, Y)
	nb, _ := NewNamedBinder("g", []*LogicVariable{Y}, nil, eq)
	// Apply nb to X → Boolean
	app := MustApply(nb, X)
	// ForAll X. nb(X)
	fa, _ := NewForAll([]*LogicVariable{X}, app)

	result := NormalizeNamedBinders(fa, nil)
	rFa := result.(*ForAll)
	// ForAll variables should be preserved
	if rFa.Variables[0].Name != "X" {
		t.Errorf("expected ForAll variable X preserved, got %s", rFa.Variables[0].Name)
	}
	// Apply inside ForAll should have NamedBinder as Func
	rApp := rFa.Body.(*Apply)
	rnb := rApp.Func.(*LogicNamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in inner NamedBinder, got %s", rnb.Variables[0].Name)
	}
}

// TestNNB_NotAndOrImplies: Simple formulas have children recursed (branch 4).
func TestNNB_NotAndOrImplies(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	eq, _ := NewEq(X, Y)
	// nb = g(X) { Eq(X, Y) } : FunctionSort(S, Boolean)
	nb, _ := NewNamedBinder("g", []*LogicVariable{X}, nil, eq)
	// Apply nb to Y → BooleanSort
	app := MustApply(nb, Y)

	// Not
	notApp, _ := NewNot(app)
	rNot := NormalizeNamedBinders(notApp, nil).(*LogicNot)
	rNotApp := rNot.Body.(*Apply)
	if rNotApp.Func.(*LogicNamedBinder).Variables[0].Name != "V0" {
		t.Error("Not: inner NamedBinder not normalized")
	}

	// And
	andApp, _ := NewAnd(app)
	rAnd := NormalizeNamedBinders(andApp, nil).(*LogicAnd)
	rAndApp := rAnd.Terms[0].(*Apply)
	if rAndApp.Func.(*LogicNamedBinder).Variables[0].Name != "V0" {
		t.Error("And: inner NamedBinder not normalized")
	}

	// Or
	orApp, _ := NewOr(app)
	rOr := NormalizeNamedBinders(orApp, nil).(*LogicOr)
	rOrApp := rOr.Terms[0].(*Apply)
	if rOrApp.Func.(*LogicNamedBinder).Variables[0].Name != "V0" {
		t.Error("Or: inner NamedBinder not normalized")
	}

	// Implies
	implApp, _ := NewImplies(app, app)
	rImpl := NormalizeNamedBinders(implApp, nil).(*LogicImplies)
	rImplT1 := rImpl.T1.(*Apply)
	rImplT2 := rImpl.T2.(*Apply)
	if rImplT1.Func.(*LogicNamedBinder).Variables[0].Name != "V0" {
		t.Error("Implies.T1: inner NamedBinder not normalized")
	}
	if rImplT2.Func.(*LogicNamedBinder).Variables[0].Name != "V0" {
		t.Error("Implies.T2: inner NamedBinder not normalized")
	}
}

// TestNNB_DeepNesting: NamedBinder inside Apply inside And (all branches).
func TestNNB_DeepNesting(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	// nb = g(X) { Eq(X,X) } : FunctionSort(S, Boolean)
	eqXX, _ := NewEq(X, X)
	nb, _ := NewNamedBinder("g", []*LogicVariable{X}, nil, eqXX)
	// app = nb(Y) → BooleanSort
	app := MustApply(nb, Y)
	// and = LogicAnd(app) → BooleanSort
	and, _ := NewAnd(app)

	result := NormalizeNamedBinders(and, nil)
	rAnd := result.(*LogicAnd)
	rApp := rAnd.Terms[0].(*Apply)
	rnb := rApp.Func.(*LogicNamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("deep nesting: expected V0, got %s", rnb.Variables[0].Name)
	}
	// Body should have V0 substituted for X in Eq(X,X) → Eq(V0,V0)
	rBody := rnb.Body.(*Eq)
	if rBody.T1.(*LogicVariable).Name != "V0" {
		t.Errorf("deep nesting: expected V0 in body Eq.T1, got %s", rBody.T1)
	}
}

// TestNNB_Idempotent: Running NormalizeNamedBinders twice gives same result.
func TestNNB_Idempotent(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	eq, _ := NewEq(X, Y)
	nb, _ := NewNamedBinder("g", []*LogicVariable{X, Y}, nil, eq)

	r1 := NormalizeNamedBinders(nb, nil)
	r2 := NormalizeNamedBinders(r1, nil)

	rnb1 := r1.(*LogicNamedBinder)
	rnb2 := r2.(*LogicNamedBinder)
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
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	// nb = g(Y) { Eq(X, Y) } : FunctionSort(S, Boolean)
	eq, _ := NewEq(X, Y)
	nb, _ := NewNamedBinder("g", []*LogicVariable{Y}, nil, eq)
	// Apply nb to X → Boolean
	app := MustApply(nb, X)
	// Exists X. nb(X)
	ex, _ := NewExists([]*LogicVariable{X}, app)

	result := NormalizeNamedBinders(ex, nil)
	rEx := result.(*LogicExists)
	if rEx.Variables[0].Name != "X" {
		t.Errorf("expected Exists variable X preserved, got %s", rEx.Variables[0].Name)
	}
	rApp := rEx.Body.(*Apply)
	rnb := rApp.Func.(*LogicNamedBinder)
	if rnb.Variables[0].Name != "V0" {
		t.Errorf("expected V0 in inner NamedBinder, got %s", rnb.Variables[0].Name)
	}
}

// TestNNB_Eq: Eq children recursed (branch 4).
func TestNNB_Eq(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	result := NormalizeNamedBinders(logicutilMustEq(t, X, Y), nil)
	rEq := result.(*Eq)
	if rEq.T1.(*LogicVariable).Name != "X" {
		t.Errorf("expected X in Eq.T1, got %s", rEq.T1)
	}
	if rEq.T2.(*LogicVariable).Name != "Y" {
		t.Errorf("expected Y in Eq.T2, got %s", rEq.T2)
	}
}

func logicutilMustEq(t *testing.T, a, b Expr) *Eq {
	t.Helper()
	e, err := NewEq(a, b)
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
		S := &UninterpretedSort{Name: "S"}
		X, _ := NewVariable("X", S)
		Y, _ := NewVariable("Y", S)
		Z, _ := NewVariable("Z", S)
		eq, _ := NewEq(X, Y)

		subs := map[NodeKey]Expr{}
		if subX {
			subs[Key(X)] = Z
		}
		if subY {
			subs[Key(Y)] = Z
		}
		// Should not panic
		result, err := Substitute(eq, subs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = result.String()
	})
}
