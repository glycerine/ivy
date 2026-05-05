package logic

import (
	"path/filepath"
	"runtime"
	"testing"

	tv "github.com/glycerine/ivy/goivy/test_vectors"
)

func logicVectorsPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "test_vectors", "sexp_vectors.sexp")
}

func logicLoadVectors(t *testing.T) map[string]string {
	t.Helper()
	vecs, err := tv.LoadVectors(logicVectorsPath())
	if err != nil {
		t.Fatalf("Failed to load vectors: %v", err)
	}
	m := make(map[string]string, len(vecs))
	for _, v := range vecs {
		m[v.ID] = v.Expected
	}
	return m
}

func logicMustVar(t *testing.T, name string, s Sort) *Variable {
	t.Helper()
	v, err := NewVariable(name, s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func logicMustFuncSort(t *testing.T, sorts ...Sort) *FunctionSort {
	t.Helper()
	fs, err := NewFunctionSort(sorts...)
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

func logicMustApply(t *testing.T, fn Expr, terms ...Expr) *Apply {
	t.Helper()
	a, err := NewApply(fn, terms...)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func logicMustEq(t *testing.T, t1, t2 Expr) *Eq {
	t.Helper()
	e, err := NewEq(t1, t2)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// helper sorts and variables used across tests
func setupCommon(t *testing.T) (Sort, *Variable, *Variable, *Eq) {
	t.Helper()
	S := &UninterpretedSort{Name: "S"}
	X := logicMustVar(t, "X", S)
	Y := logicMustVar(t, "Y", S)
	eq := logicMustEq(t, X, Y)
	return S, X, Y, eq
}

func logicCheckSexp(t *testing.T, vecs map[string]string, id string, got string) {
	t.Helper()
	expected, ok := vecs[id]
	if !ok {
		t.Fatalf("No vector for id %q", id)
	}
	if got != expected {
		t.Errorf("Sexp mismatch for %s\n  got:  %s\n  want: %s", id, got, expected)
	}
}

// ============================================================
// Sort types
// ============================================================

func TestLogicSexpUninterpretedSort(t *testing.T) {
	vecs := logicLoadVectors(t)
	s := &UninterpretedSort{Name: "S"}
	logicCheckSexp(t, vecs, "uninterp_sort_basic", string(s.Sexp()))
}

func TestLogicSexpBooleanSort(t *testing.T) {
	vecs := logicLoadVectors(t)
	logicCheckSexp(t, vecs, "boolean_sort", string(Boolean.Sexp()))
}

func TestLogicSexpFunctionSort(t *testing.T) {
	vecs := logicLoadVectors(t)
	S := &UninterpretedSort{Name: "S"}

	fs2 := logicMustFuncSort(t, S, S, Boolean)
	logicCheckSexp(t, vecs, "func_sort_binary", string(fs2.Sexp()))

	fs1 := logicMustFuncSort(t, S, Boolean)
	logicCheckSexp(t, vecs, "func_sort_unary", string(fs1.Sexp()))
}

func TestLogicSexpEnumeratedSort(t *testing.T) {
	vecs := logicLoadVectors(t)

	es := &EnumeratedSort{Name: "Color", Extension: []string{"red", "green", "blue"}}
	logicCheckSexp(t, vecs, "enum_sort", string(es.Sexp()))

	es1 := &EnumeratedSort{Name: "X", Extension: []string{"a"}}
	logicCheckSexp(t, vecs, "enum_sort_single", string(es1.Sexp()))

	es0 := &EnumeratedSort{Name: "E", Extension: []string{}}
	logicCheckSexp(t, vecs, "enum_sort_empty", string(es0.Sexp()))
}

func TestLogicSexpRangeSort(t *testing.T) {
	vecs := logicLoadVectors(t)
	rs := &RangeSort{Name: "idx", Lb: NumeralBound{Value: "0"}, Ub: NumeralBound{Value: "10"}}
	logicCheckSexp(t, vecs, "range_sort", string(rs.Sexp()))
}

func TestLogicSexpTopSort(t *testing.T) {
	vecs := logicLoadVectors(t)
	logicCheckSexp(t, vecs, "top_sort_default", string(TopS.Sexp()))

	alpha := &TopSort{Name: "Alpha"}
	logicCheckSexp(t, vecs, "top_sort_named", string(alpha.Sexp()))
}

// ============================================================
// Term types
// ============================================================

func TestLogicSexpVariable(t *testing.T) {
	vecs := logicLoadVectors(t)
	S := &UninterpretedSort{Name: "S"}
	X := logicMustVar(t, "X", S)
	logicCheckSexp(t, vecs, "variable", string(X.Sexp()))
}

func TestLogicSexpSymbol(t *testing.T) {
	vecs := logicLoadVectors(t)
	S := &UninterpretedSort{Name: "S"}

	c := NewConst("c", S)
	logicCheckSexp(t, vecs, "symbol_constant", string(c.Sexp()))

	f := NewConst("f", logicMustFuncSort(t, S, Boolean))
	logicCheckSexp(t, vecs, "symbol_unary_func", string(f.Sexp()))
}

func TestLogicSexpApply(t *testing.T) {
	vecs := logicLoadVectors(t)
	S := &UninterpretedSort{Name: "S"}
	X := logicMustVar(t, "X", S)
	Y := logicMustVar(t, "Y", S)

	leq := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))
	app := logicMustApply(t, leq, X, Y)
	logicCheckSexp(t, vecs, "apply_binary", string(app.Sexp()))

	// Nullary: f : () -> Boolean
	fNull := NewConst("f", logicMustFuncSort(t, Boolean))
	appNull := &Apply{Func: fNull, Terms: nil}
	logicCheckSexp(t, vecs, "apply_nullary", string(appNull.Sexp()))

	// Nested: g(f(X))
	f := NewConst("f", logicMustFuncSort(t, S, Boolean))
	g := NewConst("g", logicMustFuncSort(t, Boolean, Boolean))
	inner := logicMustApply(t, f, X)
	outer := logicMustApply(t, g, inner)
	logicCheckSexp(t, vecs, "apply_nested", string(outer.Sexp()))
}

// ============================================================
// Formula types
// ============================================================

func TestLogicSexpEq(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, Y, eq := setupCommon(t)
	_ = X
	_ = Y
	logicCheckSexp(t, vecs, "eq", string(eq.Sexp()))
}

func TestLogicSexpNot(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, _, _, eq := setupCommon(t)
	not, err := NewNot(eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "not", string(not.Sexp()))
}

func TestLogicSexpAnd(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, _, _, eq := setupCommon(t)

	and2, err := NewAnd(eq, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "and_two", string(and2.Sexp()))

	// Empty And = true
	logicCheckSexp(t, vecs, "and_empty", string(True.Sexp()))
}

func TestLogicSexpOr(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, _, _, eq := setupCommon(t)

	or2, err := NewOr(eq, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "or_two", string(or2.Sexp()))

	// Empty Or = false
	logicCheckSexp(t, vecs, "or_empty", string(False.Sexp()))
}

func TestLogicSexpImplies(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, _, _, eq := setupCommon(t)
	imp, err := NewImplies(eq, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "implies", string(imp.Sexp()))
}

func TestLogicSexpIff(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, _, _, eq := setupCommon(t)
	iff, err := NewIff(eq, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "iff", string(iff.Sexp()))
}

func TestLogicSexpIte(t *testing.T) {
	vecs := logicLoadVectors(t)
	S, X, Y, eq := setupCommon(t)
	_ = S
	ite, err := NewIte(eq, X, Y)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "ite", string(ite.Sexp()))
}

func TestLogicSexpGlobally(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, _, _, eq := setupCommon(t)

	g1, err := NewGlobally(nil, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "globally_nil_env", string(g1.Sexp()))

	env := "env1"
	g2, err := NewGlobally(&env, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "globally_with_env", string(g2.Sexp()))
}

func TestLogicSexpEventually(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, _, _, eq := setupCommon(t)

	e1, err := NewEventually(nil, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "eventually_nil_env", string(e1.Sexp()))

	env := "env1"
	e2, err := NewEventually(&env, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "eventually_with_env", string(e2.Sexp()))
}

func TestLogicSexpWhenOperator(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, _, eq := setupCommon(t)
	w, err := NewWhenOperator("when", X, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "when_operator", string(w.Sexp()))
}

func TestLogicSexpCond(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, Y, _ := setupCommon(t)
	c, err := NewCond(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "cond", string(c.Sexp()))
}

// ============================================================
// Quantifier types
// ============================================================

func TestLogicSexpForAll(t *testing.T) {
	vecs := logicLoadVectors(t)
	S, X, _, eq := setupCommon(t)

	fa, err := NewForAll([]*Variable{X}, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "forall_single", string(fa.Sexp()))

	// Multi-var sorted: pass Z, A — should appear as A, Z
	A := logicMustVar(t, "A", S)
	Z := logicMustVar(t, "Z", S)
	eqAZ := logicMustEq(t, A, Z)
	fa2, err := NewForAll([]*Variable{Z, A}, eqAZ)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "forall_multi_sorted", string(fa2.Sexp()))
}

func TestLogicSexpExists(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, _, eq := setupCommon(t)
	ex, err := NewExists([]*Variable{X}, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "exists", string(ex.Sexp()))
}

func TestLogicSexpLambda(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, _, _ := setupCommon(t)
	lam, err := NewLambda([]*Variable{X}, X)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "lambda", string(lam.Sexp()))
}

func TestLogicSexpNamedBinder(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, _, eq := setupCommon(t)

	nb1, err := NewNamedBinder("nb", []*Variable{X}, nil, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "named_binder_nil_env", string(nb1.Sexp()))

	env := "e1"
	nb2, err := NewNamedBinder("nb", []*Variable{X}, &env, eq)
	if err != nil {
		t.Fatal(err)
	}
	logicCheckSexp(t, vecs, "named_binder_with_env", string(nb2.Sexp()))
}

// ============================================================
// Definition types
// ============================================================

func TestLogicSexpDefinition(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, Y, _ := setupCommon(t)
	d := NewDefinition(X, Y)
	logicCheckSexp(t, vecs, "definition", string(d.Sexp()))
}

func TestLogicSexpDefinitionSchema(t *testing.T) {
	vecs := logicLoadVectors(t)
	_, X, Y, _ := setupCommon(t)
	ds := NewDefinitionSchema(X, Y)
	logicCheckSexp(t, vecs, "definition_schema", string(ds.Sexp()))
}

// ============================================================
// Canon == Sexp for all types
// ============================================================

func TestLogicCanonEqualsSexp(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X := logicMustVar(t, "X", S)
	Y := logicMustVar(t, "Y", S)
	eq := logicMustEq(t, X, Y)
	env := "e1"
	fs := logicMustFuncSort(t, S, Boolean)

	type testCase struct {
		name  string
		sexp  string
		canon string
	}

	nodes := []testCase{
		{"UninterpretedSort", string(S.Sexp()), string(S.Canon())},
		{"BooleanSort", string(Boolean.Sexp()), string(Boolean.Canon())},
		{"FunctionSort", string(fs.Sexp()), string(fs.Canon())},
		{"EnumeratedSort", string((&EnumeratedSort{Name: "E", Extension: []string{"a"}}).Sexp()), string((&EnumeratedSort{Name: "E", Extension: []string{"a"}}).Canon())},
		{"RangeSort", string((&RangeSort{Name: "r", Lb: NumeralBound{"0"}, Ub: NumeralBound{"5"}}).Sexp()), string((&RangeSort{Name: "r", Lb: NumeralBound{"0"}, Ub: NumeralBound{"5"}}).Canon())},
		{"TopSort", string(TopS.Sexp()), string(TopS.Canon())},
		{"Variable", string(X.Sexp()), string(X.Canon())},
		{"Symbol", string(NewConst("c", S).Sexp()), string(NewConst("c", S).Canon())},
	}

	app := logicMustApply(t, NewConst("f", fs), X)
	nodes = append(nodes, testCase{"Apply", string(app.Sexp()), string(app.Canon())})
	nodes = append(nodes, testCase{"Eq", string(eq.Sexp()), string(eq.Canon())})

	not, _ := NewNot(eq)
	nodes = append(nodes, testCase{"Not", string(not.Sexp()), string(not.Canon())})

	and, _ := NewAnd(eq)
	nodes = append(nodes, testCase{"And", string(and.Sexp()), string(and.Canon())})

	or, _ := NewOr(eq)
	nodes = append(nodes, testCase{"Or", string(or.Sexp()), string(or.Canon())})

	imp, _ := NewImplies(eq, eq)
	nodes = append(nodes, testCase{"Implies", string(imp.Sexp()), string(imp.Canon())})

	iff, _ := NewIff(eq, eq)
	nodes = append(nodes, testCase{"Iff", string(iff.Sexp()), string(iff.Canon())})

	ite, _ := NewIte(eq, X, Y)
	nodes = append(nodes, testCase{"Ite", string(ite.Sexp()), string(ite.Canon())})

	glob, _ := NewGlobally(&env, eq)
	nodes = append(nodes, testCase{"Globally", string(glob.Sexp()), string(glob.Canon())})

	ev, _ := NewEventually(nil, eq)
	nodes = append(nodes, testCase{"Eventually", string(ev.Sexp()), string(ev.Canon())})

	when, _ := NewWhenOperator("w", X, eq)
	nodes = append(nodes, testCase{"WhenOperator", string(when.Sexp()), string(when.Canon())})

	cond, _ := NewCond(X, Y)
	nodes = append(nodes, testCase{"Cond", string(cond.Sexp()), string(cond.Canon())})

	fa, _ := NewForAll([]*Variable{X}, eq)
	nodes = append(nodes, testCase{"ForAll", string(fa.Sexp()), string(fa.Canon())})

	ex, _ := NewExists([]*Variable{X}, eq)
	nodes = append(nodes, testCase{"Exists", string(ex.Sexp()), string(ex.Canon())})

	lam, _ := NewLambda([]*Variable{X}, X)
	nodes = append(nodes, testCase{"Lambda", string(lam.Sexp()), string(lam.Canon())})

	nb, _ := NewNamedBinder("nb", []*Variable{X}, nil, eq)
	nodes = append(nodes, testCase{"NamedBinder", string(nb.Sexp()), string(nb.Canon())})

	def := NewDefinition(X, Y)
	nodes = append(nodes, testCase{"Definition", string(def.Sexp()), string(def.Canon())})

	ds := NewDefinitionSchema(X, Y)
	nodes = append(nodes, testCase{"DefinitionSchema", string(ds.Sexp()), string(ds.Canon())})

	for _, tc := range nodes {
		if tc.sexp != tc.canon {
			t.Errorf("Canon != Sexp for %s:\n  sexp:  %s\n  canon: %s", tc.name, tc.sexp, tc.canon)
		}
	}
}
