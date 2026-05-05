package goivy_test

// separate logicutil_test avoids this import cycle:
//
// package github.com/glycerine/ivy/goivy/logicutil
//  imports github.com/glycerine/ivy/goivy/ivylogic from div_conformance_test.go
//  imports github.com/glycerine/ivy/goivy/logicutil from classify.go: import cycle not allowed in test

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func divPyScript() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "pytesthelper", "div_conformance.py")
}

func runPyDiv(t *testing.T, name string) map[string]string {
	t.Helper()
	script := divPyScript()
	cmd := exec.Command("python3", script, name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Python %s failed: %v\nOutput: %s", name, err, out)
	}
	result := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "XTRACE:") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// TestDIV2_SubstituteByNameLambda tests that SubstituteByName does NOT
// protect Lambda's bound variables from substitution.
//
// Python: Lambda is not in is_quantifier, so substitute_ast flows
// substitution into the Lambda body unchanged.
//
// Go BUG: Lambda case in substituteByNameRec removes bound names
// from subs before recursing, preventing substitution.
func TestDIV2_SubstituteByNameLambda(t *testing.T) {
	S := &goivy.UninterpretedSort{Name: "S"}
	X, _ := goivy.NewVariable("X", S)
	Y, _ := goivy.NewVariable("Y", S)
	Z, _ := goivy.NewVariable("Z", S)

	eq, _ := goivy.NewEq(X, Y)
	lam := &goivy.Lambda{Variables: []*goivy.Variable{X}, Body: eq}

	subs := map[string]goivy.Expr{"X": Z}
	result := goivy.SubstituteByName(lam, subs)

	resultLam, ok := result.(*goivy.Lambda)
	if !ok {
		t.Fatalf("result should be *Lambda, got %T", result)
	}
	resultEq, ok := resultLam.Body.(*goivy.Eq)
	if !ok {
		t.Fatalf("body should be *Eq, got %T", resultLam.Body)
	}

	// Python: body.t1 becomes Z (substitution flows through Lambda).
	// Go bug: body.t1 stays X (Lambda blocks substitution).
	t1Var, ok := resultEq.T1.(*goivy.Variable)
	if !ok {
		t.Fatalf("body.T1 should be *Variable, got %T", resultEq.T1)
	}
	if t1Var.Name != "Z" {
		t.Errorf("DIV-2: Go SubstituteByName blocks substitution inside Lambda.\n"+
			"  body.T1.Name = %q, want %q (Python does not protect Lambda vars).",
			t1Var.Name, "Z")
	}

	// Cross-check with Python
	py := runPyDiv(t, "div2")
	if pyName, ok := py["body_t1_name"]; ok && pyName != "Z" {
		t.Errorf("unexpected Python result: body_t1_name=%s", pyName)
	}
}

// TestDIV4_SubstituteApplyMissingAssertion tests that SubstituteApply
// detects when a substitution function introduces new free variables.
//
// Python: asserts fv(result) <= fv(input_terms) after substitution.
// Go BUG: no such check; new free variables silently appear.
func TestDIV4_SubstituteApplyMissingAssertion(t *testing.T) {
	S := &goivy.UninterpretedSort{Name: "S"}
	X, _ := goivy.NewVariable("X", S)
	Z, _ := goivy.NewVariable("Z", S)

	fs, _ := goivy.NewFunctionSort(S, S)
	f := goivy.NewConst("f", fs)
	app := goivy.MustApply(f, X)

	// Substitution that introduces Z as a new free variable
	subs := make(map[goivy.NodeKey]goivy.SubstituteApplyFunc)
	subs[goivy.Key(f)] = func(terms []goivy.Expr) goivy.Expr {
		eq, _ := goivy.NewEq(terms[0], Z)
		return eq
	}

	// Python: assert fvr <= fvt fires. Go should panic.
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		goivy.
			SubstituteApply(app, subs)
	}()

	if !panicked {
		t.Errorf("DIV-4: Go SubstituteApply did not panic on new free variable Z.\n" +
			"  Python asserts fv(result) <= fv(terms). Go should panic.")
	}

	// Cross-check with Python
	py := runPyDiv(t, "div4")
	if py["assertion"] != "yes" {
		t.Errorf("expected Python to fire assertion, got: %v", py)
	}
}

// TestDIV5_SomeBinderNotExcluded tests that FreeVariables/VariablesAstList
// correctly excludes bound variables from ivylogic.Some.
//
// Python: is_binder includes Some, so variables_ast excludes Some's params.
// Go BUG: no case for *ivylogic.Some in variablesAstRec — params leak as free.
func TestDIV5_SomeBinderNotExcluded(t *testing.T) {
	S := &goivy.UninterpretedSort{Name: "S"}
	X, _ := goivy.NewVariable("X", S)
	Y, _ := goivy.NewVariable("Y", S)

	eq, _ := goivy.NewEq(X, Y)
	some := goivy.NewSome([]goivy.Expr{X}, eq)

	fvs := goivy.VariablesAstList(some)
	fvNames := make(map[string]bool)
	for _, v := range fvs {
		fvNames[v.Name] = true
	}

	// X is bound by Some — should NOT appear in free variables.
	if fvNames["X"] {
		t.Errorf("DIV-5: Go VariablesAstList includes X from Some(X, Eq(X,Y)).\n"+
			"  X is bound by Some and should be excluded (Python does this via is_binder).\n"+
			"  Got free vars: %v", nameList(fvs))
	}
	// Y should be free.
	if !fvNames["Y"] {
		t.Errorf("Y should be free in Some(X, Eq(X,Y)), got: %v", nameList(fvs))
	}

	// Cross-check with Python
	py := runPyDiv(t, "div5")
	if pyFV, ok := py["free_vars"]; ok && pyFV != "Y" {
		t.Errorf("expected Python free_vars=Y, got %s", pyFV)
	}
}

// TestDIV7_NormalizeQuantifiersLambda tests that NormalizeQuantifiers
// rejects Lambda input (Python asserts False).
//
// Python: hits assert False, type(t) — Lambda should never reach here.
// Go BUG: returns Lambda unchanged, no error.
func TestDIV7_NormalizeQuantifiersLambda(t *testing.T) {
	S := &goivy.UninterpretedSort{Name: "S"}
	X, _ := goivy.NewVariable("X", S)
	eq, _ := goivy.NewEq(X, X)
	lam := &goivy.Lambda{Variables: []*goivy.Variable{X}, Body: eq}

	// Python: assert False, type(t) — a "should never reach here" guard.
	// Go should panic to match.
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		goivy.
			NormalizeQuantifiers(lam)
	}()

	if !panicked {
		t.Errorf("DIV-7: Go NormalizeQuantifiers did not panic on Lambda.\n" +
			"  Python crashes with 'assert False, type(t)' on Lambda input.")
	}

	// Cross-check with Python
	py := runPyDiv(t, "div7")
	if py["panics"] != "yes" {
		t.Errorf("expected Python to panic on Lambda, got: %v", py)
	}
}

// TestDIV8_CloseEPRVariableOrdering verifies that CloseEPR variable
// ordering matches between Go and Python.
//
// VERIFIED NON-BUG: Python's ForAll._preprocess_ (logic.py:380) sorts
// variables by name: tuple(sorted(set(variables), key=lambda v: v.name)).
// Go's FreeVariablesList sorts by NodeKey string, which also sorts by
// name first. Both produce the same alphabetical ordering.
//
// This test uses a 3-variable mixed-sort formula to exercise the hardest
// ordering case and confirms they agree.
func TestDIV8_CloseEPRVariableOrdering(t *testing.T) {
	S := &goivy.UninterpretedSort{Name: "S"}
	T := &goivy.UninterpretedSort{Name: "T"}
	B_T, _ := goivy.NewVariable("B", T)
	C_S, _ := goivy.NewVariable("C", S)
	A_S, _ := goivy.NewVariable("A", S)

	// B:T appears first in DFS, then C:S, then A:S.
	// Python ForAll sorts by name → [A:S, B:T, C:S].
	// Go FreeVariablesList sorts by NodeKey → [A:S, B:T, C:S].
	inner, _ := goivy.NewEq(C_S, A_S)
	outer, _ := goivy.NewEq(B_T, B_T)
	fmla := &goivy.Implies{T1: outer, T2: inner}

	result := goivy.CloseEPR(fmla)
	fa, ok := result.(*goivy.ForAll)
	if !ok {
		t.Fatalf("CloseEPR should produce ForAll, got %T", result)
	}

	goOrder := nameList(fa.Variables)

	// Get Python's ordering for the same formula
	py := runPyDiv(t, "div8")
	pyOrder := py["var_order"]

	goOrderStr := strings.Join(goOrder, ",")
	if goOrderStr != pyOrder {
		t.Errorf("DIV-8: CloseEPR variable ordering differs.\n"+
			"  Go order:     %s\n"+
			"  Python order: %s\n"+
			"  Both should sort alphabetically by name.",
			goOrderStr, pyOrder)
	}
}

func nameList(vars []*goivy.Variable) []string {
	names := make([]string, len(vars))
	for i, v := range vars {
		names[i] = v.Name
	}
	return names
}

func freeVarNames(t goivy.Expr) []string {
	fvs := goivy.FreeVariablesList(t)
	var names []string
	for _, v := range fvs {
		names = append(names, v.Name)
	}
	return names
}

func init() {
	// Silence for tests
	_ = fmt.Sprintf
}
