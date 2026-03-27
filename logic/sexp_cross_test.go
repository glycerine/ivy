package logic_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	lg "github.com/glycerine/goivy/logic"
)

func testVectorsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "test_vectors")
}

// buildAllGoSexps constructs all 30+ node types in Go and returns id -> sexp.
func buildAllGoSexps(t *testing.T) map[string]string {
	t.Helper()

	S := &lg.UninterpretedSort{Name: "S"}
	X, err := lg.NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	Y, err := lg.NewVariable("Y", S)
	if err != nil {
		t.Fatal(err)
	}
	A, err := lg.NewVariable("A", S)
	if err != nil {
		t.Fatal(err)
	}
	Z, err := lg.NewVariable("Z", S)
	if err != nil {
		t.Fatal(err)
	}
	eq, err := lg.NewEq(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	eqAZ, err := lg.NewEq(A, Z)
	if err != nil {
		t.Fatal(err)
	}

	fsBin, _ := lg.NewFunctionSort(S, S, lg.Boolean)
	fsUn, _ := lg.NewFunctionSort(S, lg.Boolean)
	fsNull, _ := lg.NewFunctionSort(lg.Boolean)
	fsBoolBool, _ := lg.NewFunctionSort(lg.Boolean, lg.Boolean)

	leq := lg.NewSymbol("leq", fsBin)
	fSym := lg.NewSymbol("f", fsUn)
	gSym := lg.NewSymbol("g", fsBoolBool)
	fNull := lg.NewSymbol("f", fsNull)

	appBin, _ := lg.NewApply(leq, X, Y)
	appNull := &lg.Apply{Func: fNull, Terms: nil}
	appInner, _ := lg.NewApply(fSym, X)
	appNested, _ := lg.NewApply(gSym, appInner)

	not, _ := lg.NewNot(eq)
	and2, _ := lg.NewAnd(eq, eq)
	or2, _ := lg.NewOr(eq, eq)
	imp, _ := lg.NewImplies(eq, eq)
	iff, _ := lg.NewIff(eq, eq)
	ite, _ := lg.NewIte(eq, X, Y)
	globNil, _ := lg.NewGlobally(nil, eq)
	env1 := "env1"
	globEnv, _ := lg.NewGlobally(&env1, eq)
	evNil, _ := lg.NewEventually(nil, eq)
	evEnv, _ := lg.NewEventually(&env1, eq)
	when, _ := lg.NewWhenOperator("when", X, eq)
	cond, _ := lg.NewCond(X, Y)
	faSingle, _ := lg.NewForAll([]*lg.Variable{X}, eq)
	faMulti, _ := lg.NewForAll([]*lg.Variable{Z, A}, eqAZ)
	ex, _ := lg.NewExists([]*lg.Variable{X}, eq)
	lam, _ := lg.NewLambda([]*lg.Variable{X}, X)
	nbNil, _ := lg.NewNamedBinder("nb", []*lg.Variable{X}, nil, eq)
	e1Str := "e1"
	nbEnv, _ := lg.NewNamedBinder("nb", []*lg.Variable{X}, &e1Str, eq)
	def := lg.NewDefinition(X, Y)
	defSchema := lg.NewDefinitionSchema(X, Y)

	m := map[string]string{
		"uninterp_sort_basic":  string(S.Sexp()),
		"boolean_sort":         string(lg.Boolean.Sexp()),
		"func_sort_binary":     string(fsBin.Sexp()),
		"func_sort_unary":      string(fsUn.Sexp()),
		"enum_sort":            string((&lg.EnumeratedSort{Name: "Color", Extension: []string{"red", "green", "blue"}}).Sexp()),
		"enum_sort_single":     string((&lg.EnumeratedSort{Name: "X", Extension: []string{"a"}}).Sexp()),
		"enum_sort_empty":      string((&lg.EnumeratedSort{Name: "E", Extension: []string{}}).Sexp()),
		"range_sort":           string((&lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "10"}}).Sexp()),
		"top_sort_default":     string(lg.TopS.Sexp()),
		"top_sort_named":       string((&lg.TopSort{Name: "Alpha"}).Sexp()),
		"variable":             string(X.Sexp()),
		"symbol_constant":      string(lg.NewSymbol("c", S).Sexp()),
		"symbol_unary_func":    string(fSym.Sexp()),
		"apply_binary":         string(appBin.Sexp()),
		"apply_nullary":        string(appNull.Sexp()),
		"apply_nested":         string(appNested.Sexp()),
		"eq":                   string(eq.Sexp()),
		"not":                  string(not.Sexp()),
		"and_two":              string(and2.Sexp()),
		"and_empty":            string(lg.True.Sexp()),
		"or_two":               string(or2.Sexp()),
		"or_empty":             string(lg.False.Sexp()),
		"implies":              string(imp.Sexp()),
		"iff":                  string(iff.Sexp()),
		"ite":                  string(ite.Sexp()),
		"globally_nil_env":     string(globNil.Sexp()),
		"globally_with_env":    string(globEnv.Sexp()),
		"eventually_nil_env":   string(evNil.Sexp()),
		"eventually_with_env":  string(evEnv.Sexp()),
		"when_operator":        string(when.Sexp()),
		"cond":                 string(cond.Sexp()),
		"forall_single":        string(faSingle.Sexp()),
		"forall_multi_sorted":  string(faMulti.Sexp()),
		"exists":               string(ex.Sexp()),
		"lambda":               string(lam.Sexp()),
		"named_binder_nil_env": string(nbNil.Sexp()),
		"named_binder_with_env": string(nbEnv.Sexp()),
		"definition":           string(def.Sexp()),
		"definition_schema":    string(defSchema.Sexp()),
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
