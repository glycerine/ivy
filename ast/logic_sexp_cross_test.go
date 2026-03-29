package ast_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testVectorsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "test_vectors")
}

// buildAllGoSexps constructs all 30+ node types in Go and returns id -> sexp.
func buildAllGoSexps(t *testing.T) map[string]string {
	t.Helper()

	S := &UninterpretedSort{Name: "S"}
	X, err := NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	Y, err := NewVariable("Y", S)
	if err != nil {
		t.Fatal(err)
	}
	A, err := NewVariable("A", S)
	if err != nil {
		t.Fatal(err)
	}
	Z, err := NewVariable("Z", S)
	if err != nil {
		t.Fatal(err)
	}
	eq, err := NewEq(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	eqAZ, err := NewEq(A, Z)
	if err != nil {
		t.Fatal(err)
	}

	fsBin, _ := NewFunctionSort(S, S, Boolean)
	fsUn, _ := NewFunctionSort(S, Boolean)
	fsNull, _ := NewFunctionSort(Boolean)
	fsBoolBool, _ := NewFunctionSort(Boolean, Boolean)

	leq := NewSymbol("leq", fsBin)
	fSym := NewSymbol("f", fsUn)
	gSym := NewSymbol("g", fsBoolBool)
	fNull := NewSymbol("f", fsNull)

	appBin, _ := NewApply(leq, X, Y)
	appNull := &Apply{Func: fNull, Terms: nil}
	appInner, _ := NewApply(fSym, X)
	appNested, _ := NewApply(gSym, appInner)

	not, _ := NewNot(eq)
	and2, _ := NewAnd(eq, eq)
	or2, _ := NewOr(eq, eq)
	imp, _ := NewImplies(eq, eq)
	iff, _ := NewIff(eq, eq)
	ite, _ := NewIte(eq, X, Y)
	globNil, _ := NewGlobally(nil, eq)
	env1 := "env1"
	globEnv, _ := NewGlobally(&env1, eq)
	evNil, _ := NewEventually(nil, eq)
	evEnv, _ := NewEventually(&env1, eq)
	when, _ := NewWhenOperator("when", X, eq)
	cond, _ := NewCond(X, Y)
	faSingle, _ := NewForAll([]*Variable{X}, eq)
	faMulti, _ := NewForAll([]*Variable{Z, A}, eqAZ)
	ex, _ := NewExists([]*Variable{X}, eq)
	lam, _ := NewLambda([]*Variable{X}, X)
	nbNil, _ := NewNamedBinder("nb", []*Variable{X}, nil, eq)
	e1Str := "e1"
	nbEnv, _ := NewNamedBinder("nb", []*Variable{X}, &e1Str, eq)
	def := NewLogicDefinition(X, Y)
	defSchema := NewLogicDefinitionSchema(X, Y)

	m := map[string]string{
		"uninterp_sort_basic":   string(S.Sexp()),
		"boolean_sort":          string(Boolean.Sexp()),
		"func_sort_binary":      string(fsBin.Sexp()),
		"func_sort_unary":       string(fsUn.Sexp()),
		"enum_sort":             string((&EnumeratedSort{Name: "Color", Extension: []string{"red", "green", "blue"}}).Sexp()),
		"enum_sort_single":      string((&EnumeratedSort{Name: "X", Extension: []string{"a"}}).Sexp()),
		"enum_sort_empty":       string((&EnumeratedSort{Name: "E", Extension: []string{}}).Sexp()),
		"range_sort":            string((&RangeSort{Name: "idx", Lb: NumeralBound{Value: "0"}, Ub: NumeralBound{Value: "10"}}).Sexp()),
		"top_sort_default":      string(TopS.Sexp()),
		"top_sort_named":        string((&TopSort{Name: "Alpha"}).Sexp()),
		"variable":              string(X.Sexp()),
		"symbol_constant":       string(NewSymbol("c", S).Sexp()),
		"symbol_unary_func":     string(fSym.Sexp()),
		"apply_binary":          string(appBin.Sexp()),
		"apply_nullary":         string(appNull.Sexp()),
		"apply_nested":          string(appNested.Sexp()),
		"eq":                    string(eq.Sexp()),
		"not":                   string(not.Sexp()),
		"and_two":               string(and2.Sexp()),
		"and_empty":             string(True.Sexp()),
		"or_two":                string(or2.Sexp()),
		"or_empty":              string(False.Sexp()),
		"implies":               string(imp.Sexp()),
		"iff":                   string(iff.Sexp()),
		"ite":                   string(ite.Sexp()),
		"globally_nil_env":      string(globNil.Sexp()),
		"globally_with_env":     string(globEnv.Sexp()),
		"eventually_nil_env":    string(evNil.Sexp()),
		"eventually_with_env":   string(evEnv.Sexp()),
		"when_operator":         string(when.Sexp()),
		"cond":                  string(cond.Sexp()),
		"forall_single":         string(faSingle.Sexp()),
		"forall_multi_sorted":   string(faMulti.Sexp()),
		"exists":                string(ex.Sexp()),
		"lambda":                string(lam.Sexp()),
		"named_binder_nil_env":  string(nbNil.Sexp()),
		"named_binder_with_env": string(nbEnv.Sexp()),
		"definition":            string(def.Sexp()),
		"definition_schema":     string(defSchema.Sexp()),
	}
	return m
}

func TestSexpCrossLanguage(t *testing.T) {
	// Check if python3 is available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}

	goSexps := buildAllGoSexps(t)

	// Run Python subprocess
	pyScript := filepath.Join(testVectorsDir(), "emit_sexp.py")
	cmd := exec.Command("python3", pyScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Python subprocess failed: %v\nOutput: %s", err, out)
	}

	// Parse Python output: id\tsexp per line
	pyMap := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			t.Fatalf("Bad Python output line: %q", line)
		}
		pyMap[parts[0]] = parts[1]
	}

	// Compare Go vs Python for each type
	for id, goSexp := range goSexps {
		pySexp, ok := pyMap[id]
		if !ok {
			t.Errorf("Python missing vector %q", id)
			continue
		}
		if goSexp != pySexp {
			t.Errorf("Cross-language mismatch for %s:\n  Go: %s\n  Py: %s", id, goSexp, pySexp)
		}
	}

	// Check for Python-only vectors
	for id := range pyMap {
		if _, ok := goSexps[id]; !ok {
			t.Errorf("Python has vector %q but Go does not", id)
		}
	}
}
