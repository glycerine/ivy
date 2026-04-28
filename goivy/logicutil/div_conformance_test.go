package logicutil_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	"github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/logicutil"
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
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)
	Z, _ := logic.NewVariable("Z", S)

	eq, _ := logic.NewEq(X, Y)
	lam := &logic.Lambda{Variables: []*logic.Variable{X}, Body: eq}

	subs := map[string]logic.Expr{"X": Z}
	result := logicutil.SubstituteByName(lam, subs)

	resultLam, ok := result.(*logic.Lambda)
	if !ok {
		t.Fatalf("result should be *Lambda, got %T", result)
	}
	resultEq, ok := resultLam.Body.(*logic.Eq)
	if !ok {
		t.Fatalf("body should be *Eq, got %T", resultLam.Body)
	}

	// Python: body.t1 becomes Z (substitution flows through Lambda).
	// Go bug: body.t1 stays X (Lambda blocks substitution).
	t1Var, ok := resultEq.T1.(*logic.Variable)
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
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Z, _ := logic.NewVariable("Z", S)

	fs, _ := logic.NewFunctionSort(S, S)
	f := logic.NewConst("f", fs)
	app := logic.MustApply(f, X)

	// Substitution that introduces Z as a new free variable
	subs := make(map[logic.NodeKey]logicutil.SubstituteApplyFunc)
	subs[logic.Key(f)] = func(terms []logic.Expr) logic.Expr {
		eq, _ := logic.NewEq(terms[0], Z)
		return eq
	}

	result := logicutil.SubstituteApply(app, subs)

	// Check: result has Z as a free variable that wasn't in the input terms
	inputFVs := logicutil.FreeVariables(app)
	resultFVs := logicutil.FreeVariables(result)

	if _, zInInput := inputFVs.Get2(logic.Key(Z)); zInInput {
		t.Fatal("Z should not be free in original Apply(f, X)")
	}
	_, zInResult := resultFVs.Get2(logic.Key(Z))
	if !zInResult {
		t.Fatal("Z should be free in result (substitution introduced it)")
	}

	// DIV-4: Go should detect this and panic/error (Python asserts).
	// Currently Go silently allows new free variables.
	t.Errorf("DIV-4: Go SubstituteApply does not detect new free variable Z.\n"+
		"  Python asserts fv(result) <= fv(terms). Go should do the same.\n"+
		"  result fvs: %v", freeVarNames(result))

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
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	Y, _ := logic.NewVariable("Y", S)

	eq, _ := logic.NewEq(X, Y)
	some := il.NewSome([]logic.Expr{X}, eq)

	fvs := logicutil.VariablesAstList(some)
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
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVariable("X", S)
	eq, _ := logic.NewEq(X, X)
	lam := &logic.Lambda{Variables: []*logic.Variable{X}, Body: eq}

	// Go: silently returns Lambda unchanged.
	// Python: assert False, type(t) — a "should never reach here" guard.
	result := logicutil.NormalizeQuantifiers(lam)
	if _, isLam := result.(*logic.Lambda); isLam {
		t.Errorf("DIV-7: Go NormalizeQuantifiers silently accepts Lambda.\n"+
			"  Python crashes with 'assert False, type(t)' on Lambda input.\n"+
			"  Go should either panic or refuse Lambda.")
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
	S := &logic.UninterpretedSort{Name: "S"}
	T := &logic.UninterpretedSort{Name: "T"}
	B_T, _ := logic.NewVariable("B", T)
	C_S, _ := logic.NewVariable("C", S)
	A_S, _ := logic.NewVariable("A", S)

	// B:T appears first in DFS, then C:S, then A:S.
	// Python ForAll sorts by name → [A:S, B:T, C:S].
	// Go FreeVariablesList sorts by NodeKey → [A:S, B:T, C:S].
	inner, _ := logic.NewEq(C_S, A_S)
	outer, _ := logic.NewEq(B_T, B_T)
	fmla := &logic.Implies{T1: outer, T2: inner}

	result := logicutil.CloseEPR(fmla)
	fa, ok := result.(*logic.ForAll)
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

func nameList(vars []*logic.Variable) []string {
	names := make([]string, len(vars))
	for i, v := range vars {
		names[i] = v.Name
	}
	return names
}

func freeVarNames(t logic.Expr) []string {
	fvs := logicutil.FreeVariablesList(t)
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
