package goivy_test

import (
	goivy "github.com/glycerine/ivy/goivy"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testVectorsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "test_vectors")
}

// buildAllGoSexps constructs all 30+ node types in Go and returns id -> sexp.
func buildAllGoSexps(t *testing.T) map[string]string {
	t.Helper()

	S := &goivy.UninterpretedSort{Name: "S"}
	X, err := goivy.NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	Y, err := goivy.NewVariable("Y", S)
	if err != nil {
		t.Fatal(err)
	}
	A, err := goivy.NewVariable("A", S)
	if err != nil {
		t.Fatal(err)
	}
	Z, err := goivy.NewVariable("Z", S)
	if err != nil {
		t.Fatal(err)
	}
	eq, err := goivy.NewEq(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	eqAZ, err := goivy.NewEq(A, Z)
	if err != nil {
		t.Fatal(err)
	}

	fsBin, _ := goivy.NewFunctionSort(S, S, goivy.Boolean)
	fsUn, _ := goivy.NewFunctionSort(S, goivy.Boolean)
	fsNull, _ := goivy.NewFunctionSort(goivy.Boolean)
	fsBoolBool, _ := goivy.NewFunctionSort(goivy.Boolean, goivy.Boolean)

	leq := goivy.NewConst("leq", fsBin)
	fSym := goivy.NewConst("f", fsUn)
	gSym := goivy.NewConst("g", fsBoolBool)
	fNull := goivy.NewConst("f", fsNull)

	appBin, _ := goivy.NewApply(leq, X, Y)
	appNull := goivy.MustApply(fNull)
	appInner, _ := goivy.NewApply(fSym, X)
	appNested, _ := goivy.NewApply(gSym, appInner)

	not, _ := goivy.NewNot(eq)
	and2, _ := goivy.NewAnd(eq, eq)
	or2, _ := goivy.NewOr(eq, eq)
	imp, _ := goivy.NewImplies(eq, eq)
	iff, _ := goivy.NewIff(eq, eq)
	ite, _ := goivy.NewIte(eq, X, Y)
	globNil, _ := goivy.NewGlobally(nil, eq)
	env1 := "env1"
	globEnv, _ := goivy.NewGlobally(&env1, eq)
	evNil, _ := goivy.NewEventually(nil, eq)
	evEnv, _ := goivy.NewEventually(&env1, eq)
	when, _ := goivy.NewWhenOperator("when", X, eq)
	cond, _ := goivy.NewCond(X, Y)
	faSingle, _ := goivy.NewForAll([]*goivy.LogicVariable{X}, eq)
	faMulti, _ := goivy.NewForAll([]*goivy.LogicVariable{Z, A}, eqAZ)
	ex, _ := goivy.NewExists([]*goivy.LogicVariable{X}, eq)
	lam, _ := goivy.NewLambda([]*goivy.LogicVariable{X}, X)
	nbNil, _ := goivy.NewNamedBinder("nb", []*goivy.LogicVariable{X}, nil, eq)
	e1Str := "e1"
	nbEnv, _ := goivy.NewNamedBinder("nb", []*goivy.LogicVariable{X}, &e1Str, eq)
	def := goivy.NewDefinition(X, Y)
	defSchema := goivy.NewDefinitionSchema(X, Y)

	m := map[string]string{
		"uninterp_sort_basic":   string(S.Sexp()),
		"boolean_sort":          string(goivy.Boolean.Sexp()),
		"func_sort_binary":      string(fsBin.Sexp()),
		"func_sort_unary":       string(fsUn.Sexp()),
		"enum_sort":             string((&goivy.LogicEnumeratedSort{Name: "Color", Extension: []string{"red", "green", "blue"}}).Sexp()),
		"enum_sort_single":      string((&goivy.LogicEnumeratedSort{Name: "X", Extension: []string{"a"}}).Sexp()),
		"enum_sort_empty":       string((&goivy.LogicEnumeratedSort{Name: "E", Extension: []string{}}).Sexp()),
		"range_sort":            string((&goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "10"}}).Sexp()),
		"top_sort_default":      string(goivy.TopS.Sexp()),
		"top_sort_named":        string((&goivy.TopSort{Name: "Alpha"}).Sexp()),
		"variable":              string(X.Sexp()),
		"symbol_constant":       string(goivy.NewConst("c", S).Sexp()),
		"symbol_unary_func":     string(fSym.Sexp()),
		"apply_binary":          string(appBin.Sexp()),
		"apply_nullary":         string(appNull.Sexp()),
		"apply_nested":          string(appNested.Sexp()),
		"eq":                    string(eq.Sexp()),
		"not":                   string(not.Sexp()),
		"and_two":               string(and2.Sexp()),
		"and_empty":             string(goivy.True.Sexp()),
		"or_two":                string(or2.Sexp()),
		"or_empty":              string(goivy.False.Sexp()),
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
	goSexps := buildAllGoSexps(t)

	// Run Python subprocess
	pyScript := filepath.Join(testVectorsDir(), "emit_sexp.py")
	cmd := externalPythonIvyCommandForTest(t, pyScript)
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
