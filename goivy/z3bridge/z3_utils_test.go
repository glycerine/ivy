package z3bridge

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/logic"
)

// --- helpers ---

func mustVar(t *testing.T, name string, sort logic.Sort) *logic.Variable {
	t.Helper()
	v, err := logic.NewVariable(name, sort)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustApply(t *testing.T, fn logic.Expr, terms ...logic.Expr) *logic.Apply {
	t.Helper()
	app, err := logic.NewApply(fn, terms...)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func mustForAll(t *testing.T, vars []*logic.Variable, body logic.Expr) *logic.ForAll {
	t.Helper()
	fa, err := logic.NewForAll(vars, body)
	if err != nil {
		t.Fatal(err)
	}
	return fa
}

func mustExists(t *testing.T, vars []*logic.Variable, body logic.Expr) *logic.Exists {
	t.Helper()
	ex, err := logic.NewExists(vars, body)
	if err != nil {
		t.Fatal(err)
	}
	return ex
}

func mustEq(t *testing.T, t1, t2 logic.Expr) *logic.Eq {
	t.Helper()
	eq, err := logic.NewEq(t1, t2)
	if err != nil {
		t.Fatal(err)
	}
	return eq
}

func mustIte(t *testing.T, cond, then_, else_ logic.Expr) *logic.Ite {
	t.Helper()
	ite, err := logic.NewIte(cond, then_, else_)
	if err != nil {
		t.Fatal(err)
	}
	return ite
}

// --- Group 1: Sort Translation ---

func TestToZ3BooleanSort(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	result, err := u.ToZ3(logic.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	zs, ok := result.(Sort)
	if !ok {
		t.Fatalf("expected Sort, got %T", result)
	}
	if zs.Kind() != SortBool {
		t.Errorf("expected Bool sort, got kind %d", zs.Kind())
	}
}

func TestToZ3UninterpretedSort(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	result, err := u.ToZ3(s)
	if err != nil {
		t.Fatal(err)
	}
	zs, ok := result.(Sort)
	if !ok {
		t.Fatalf("expected Sort, got %T", result)
	}
	if zs.Kind() != SortUninterpreted {
		t.Errorf("expected Uninterpreted sort, got kind %d", zs.Kind())
	}
}

func TestToZ3UninterpretedSortCache(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	r1, err := u.ToZ3(s)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := u.ToZ3(s)
	if err != nil {
		t.Fatal(err)
	}
	// Both should be the exact same Sort (from cache)
	s1 := r1.(Sort)
	s2 := r2.(Sort)
	if s1.GetId() != s2.GetId() {
		t.Error("expected same Z3 sort from cache")
	}
}

func TestToZ3FunctionSortError(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, s, s)
	_, err := u.ToZ3(fs)
	if err == nil {
		t.Fatal("expected error for FunctionSort")
	}
	if !strings.Contains(err.Error(), "FunctionSort") {
		t.Errorf("error should mention FunctionSort: %s", err)
	}
}

// --- Group 2: Term Translation ---

func TestToZ3Variable(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	v := mustVar(t, "X", s)
	result, err := u.ToZ3Expr(v)
	if err != nil {
		t.Fatal(err)
	}
	name := result.String()
	// Z3 wraps names containing ':' in |pipes|
	if name != "X:S" && name != "|X:S|" {
		t.Errorf("expected Z3 const named X:S, got %q", name)
	}
}

func TestToZ3Const(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	c := logic.NewConst("c", s)
	result, err := u.ToZ3Expr(c)
	if err != nil {
		t.Fatal(err)
	}
	name := result.String()
	if !strings.Contains(name, "c:S") {
		t.Errorf("expected Z3 const named c:S, got %q", name)
	}
}

func TestToZ3ConstBoolSort(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	c := logic.NewConst("p", logic.Boolean)
	result, err := u.ToZ3Expr(c)
	if err != nil {
		t.Fatal(err)
	}
	name := result.String()
	if !strings.Contains(name, "p:Boolean") {
		t.Errorf("expected Z3 const named p:Boolean, got %q", name)
	}
}

func TestToZ3ConstNullaryFunc(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, s) // arity 0: FunctionSort(S) → range is S
	c := logic.NewConst("c", fs)
	result, err := u.ToZ3Expr(c)
	if err != nil {
		t.Fatal(err)
	}
	// Nullary function converts to first-order constant with range sort
	name := result.String()
	if !strings.Contains(name, "c:S") {
		t.Errorf("expected Z3 const named c:S, got %q", name)
	}
}

func TestToZ3ConstHigherOrder(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, s, s, logic.Boolean) // S * S -> Boolean
	c := logic.NewConst("leq", fs)
	result, err := u.ToZ3(c)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := result.(FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl for higher-order const, got %T", result)
	}
}

func TestToZ3VariableHigherOrderError(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, s, s)
	// Construct Variable with FunctionSort directly (NewVariable would reject)
	v := &logic.Variable{Name: "F", VSort: fs}
	_, err := u.ToZ3(v)
	if err == nil {
		t.Fatal("expected error for higher-order variable")
	}
	if !strings.Contains(err.Error(), "high-order") {
		t.Errorf("error should mention high-order: %s", err)
	}
}

// --- Group 3: Apply Translation ---

func TestToZ3ApplyNullary(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	c := logic.NewConst("c", s)
	app := &logic.Apply{Func: c, Terms: []logic.Expr{}}
	result, err := u.ToZ3Expr(app)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.String(), "c:S") {
		t.Errorf("expected c:S, got %q", result.String())
	}
}

func TestToZ3ApplyWithArgs(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, s, s, logic.Boolean) // S * S -> Bool
	leq := logic.NewConst("leq", fs)
	x := logic.NewConst("x", s)
	y := logic.NewConst("y", s)
	app := mustApply(t, leq, x, y)

	result, err := u.ToZ3Expr(app)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	// Should be something like "leq(x:S, y:S)"
	if !strings.Contains(str, "leq") {
		t.Errorf("expected apply result to contain 'leq', got %q", str)
	}
}

// --- Group 4: Formula Translation ---

func TestToZ3Eq(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	x := logic.NewConst("x", s)
	y := logic.NewConst("y", s)
	eq := mustEq(t, x, y)
	result, err := u.ToZ3Expr(eq)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExprSort().Kind() != SortBool {
		t.Error("Eq should produce Bool-sorted expression")
	}
}

func TestToZ3Not(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	not := &logic.Not{Body: p}
	result, err := u.ToZ3Expr(not)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	if !strings.Contains(str, "Not") && !strings.Contains(str, "not") {
		t.Errorf("expected Not in result, got %q", str)
	}
}

func TestToZ3And(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	and := &logic.And{Terms: []logic.Expr{p, q}}
	result, err := u.ToZ3Expr(and)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	if !strings.Contains(str, "And") && !strings.Contains(str, "and") && !strings.Contains(str, "p:Boolean") {
		t.Errorf("expected And with p in result, got %q", str)
	}
}

func TestToZ3Or(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	or := &logic.Or{Terms: []logic.Expr{p, q}}
	result, err := u.ToZ3Expr(or)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	if !strings.Contains(str, "Or") && !strings.Contains(str, "or") && !strings.Contains(str, "p:Boolean") {
		t.Errorf("expected Or with p in result, got %q", str)
	}
}

func TestToZ3AndEmpty(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	// Empty And is logic.True → z3 BoolVal(true)
	and := &logic.And{Terms: []logic.Expr{}}
	result, err := u.ToZ3Expr(and)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	if str != "true" {
		t.Errorf("expected 'true', got %q", str)
	}
}

func TestToZ3OrEmpty(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	// Empty Or is logic.False → z3 BoolVal(false)
	or := &logic.Or{Terms: []logic.Expr{}}
	result, err := u.ToZ3Expr(or)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	if str != "false" {
		t.Errorf("expected 'false', got %q", str)
	}
}

func TestToZ3Implies(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	imp := &logic.Implies{T1: p, T2: q}
	result, err := u.ToZ3Expr(imp)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	if !strings.Contains(str, "Implies") && !strings.Contains(str, "=>") {
		t.Errorf("expected Implies in result, got %q", str)
	}
}

func TestToZ3Iff(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	// Python maps Iff to z3 equality (==), NOT z3 Iff
	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	iff := &logic.Iff{T1: p, T2: q}
	result, err := u.ToZ3Expr(iff)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	// Should be "p:Boolean = q:Boolean" or "(= p:Boolean q:Boolean)"
	if !strings.Contains(str, "=") {
		t.Errorf("Iff should map to equality, got %q", str)
	}
}

func TestToZ3Ite(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	b := logic.NewConst("b", logic.Boolean)
	x := logic.NewConst("x", s)
	y := logic.NewConst("y", s)
	ite := mustIte(t, b, x, y)
	result, err := u.ToZ3Expr(ite)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	if !strings.Contains(str, "If") && !strings.Contains(str, "ite") {
		t.Errorf("expected Ite in result, got %q", str)
	}
}

// --- Group 5: Quantifier Translation ---

func TestToZ3ForAll(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, s, logic.Boolean)
	p := logic.NewConst("P", fs)
	x := mustVar(t, "X", s)
	body := mustApply(t, p, x) // P(X)
	fa := mustForAll(t, []*logic.Variable{x}, body)

	result, err := u.ToZ3Expr(fa)
	if err != nil {
		t.Fatal(err)
	}
	str := result.String()
	t.Log("ForAll:", str)
	// Should contain quantifier syntax
	if result.ExprSort().Kind() != SortBool {
		t.Error("ForAll should produce Bool-sorted expression")
	}
}

func TestToZ3Exists(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	s := &logic.UninterpretedSort{Name: "S"}
	fs := mustFS(t, s, logic.Boolean)
	p := logic.NewConst("P", fs)
	x := mustVar(t, "X", s)
	body := mustApply(t, p, x)
	ex := mustExists(t, []*logic.Variable{x}, body)

	result, err := u.ToZ3Expr(ex)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExprSort().Kind() != SortBool {
		t.Error("Exists should produce Bool-sorted expression")
	}
}

func TestToZ3ForAllEmpty(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	// Use struct literal for empty vars (NewForAll rejects empty)
	fa := &logic.ForAll{Variables: []*logic.Variable{}, Body: p}
	result, err := u.ToZ3Expr(fa)
	if err != nil {
		t.Fatal(err)
	}
	// Should return ToZ3Expr(p) directly
	if !strings.Contains(result.String(), "p:Boolean") {
		t.Errorf("expected p:Boolean, got %q", result.String())
	}
}

func TestToZ3ExistsEmpty(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	ex := &logic.Exists{Variables: []*logic.Variable{}, Body: p}
	result, err := u.ToZ3Expr(ex)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.String(), "p:Boolean") {
		t.Errorf("expected p:Boolean, got %q", result.String())
	}
}

// --- Group 6: Caching ---

func TestToZ3CacheHit(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	r1, err := u.ToZ3(p)
	if err != nil {
		t.Fatal(err)
	}

	// Cache should have the entry now
	key := p.Sexp()
	if _, ok := u.toZ3Cache[key]; !ok {
		t.Error("cache should contain entry after first call")
	}

	r2, err := u.ToZ3(p)
	if err != nil {
		t.Fatal(err)
	}

	// Both should be same Expr
	e1, e2 := r1.(Expr), r2.(Expr)
	if e1.String() != e2.String() {
		t.Errorf("cached result differs: %q vs %q", e1.String(), e2.String())
	}
}

func TestToZ3ClearResetsCache(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	_, err := u.ToZ3(p)
	if err != nil {
		t.Fatal(err)
	}

	key := p.Sexp()
	if _, ok := u.toZ3Cache[key]; !ok {
		t.Error("cache should contain entry")
	}

	u.Clear()

	if _, ok := u.toZ3Cache[key]; ok {
		t.Error("cache should be empty after Clear")
	}

	// Should still work after Clear
	_, err = u.ToZ3(p)
	if err != nil {
		t.Fatal("ToZ3 should work after Clear:", err)
	}
}

// --- Group 7: Z3Implies ---

func TestZ3UtilsImpliesValid(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	pAndQ := &logic.And{Terms: []logic.Expr{p, q}}

	result, err := u.Z3Implies(pAndQ, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("p AND q should imply p")
	}
}

func TestZ3UtilsImpliesInvalid(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)

	result, err := u.Z3Implies(p, q, false)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Error("p should not imply q")
	}
}

func TestZ3UtilsImpliesCache(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	r1, err := u.Z3Implies(p, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r1 {
		t.Error("p should imply p")
	}

	// Second call should hit cache
	r2, err := u.Z3Implies(p, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r2 {
		t.Error("cached: p should imply p")
	}

	key := [2]logic.NodeKey{p.Sexp(), p.Sexp()}
	if _, ok := u.ImpliesCache[key]; !ok {
		t.Error("ImpliesCache should contain the entry")
	}
}

func TestZ3UtilsImpliesTautology(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	notP := &logic.Not{Body: p}
	pOrNotP := &logic.Or{Terms: []logic.Expr{p, notP}}
	trueVal := &logic.And{Terms: []logic.Expr{}} // logic.True

	result, err := u.Z3Implies(trueVal, pOrNotP, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("true should imply (p OR NOT p)")
	}
}

func TestZ3UtilsImpliesTimeout(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	pAndQ := &logic.And{Terms: []logic.Expr{p, q}}

	result, err := u.Z3Implies(pAndQ, p, true) // timeout=true
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("p AND q should imply p (with timeout)")
	}
}

// --- Group 8: Z3ImpliesBatch ---

func TestZ3UtilsBatchValid(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	pAndQ := &logic.And{Terms: []logic.Expr{p, q}}

	results, err := u.Z3ImpliesBatch(pAndQ, []logic.Expr{p, q}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0] {
		t.Error("p AND q should imply p")
	}
	if !results[1] {
		t.Error("p AND q should imply q")
	}
}

func TestZ3UtilsBatchInvalid(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)

	results, err := u.Z3ImpliesBatch(p, []logic.Expr{q}, false)
	if err != nil {
		t.Fatal(err)
	}
	if results[0] {
		t.Error("p should not imply q")
	}
}

func TestZ3UtilsBatchMixed(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	r := logic.NewConst("r", logic.Boolean)
	pAndQ := &logic.And{Terms: []logic.Expr{p, q}}

	results, err := u.Z3ImpliesBatch(pAndQ, []logic.Expr{p, r, q}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0] {
		t.Error("should imply p")
	}
	if results[1] {
		t.Error("should not imply r")
	}
	if !results[2] {
		t.Error("should imply q")
	}
}

func TestZ3UtilsBatchEmpty(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	results, err := u.Z3ImpliesBatch(p, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for nil formulas, got %d", len(results))
	}
}

func TestZ3UtilsBatchCacheHit(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	p := logic.NewConst("p", logic.Boolean)
	q := logic.NewConst("q", logic.Boolean)
	pAndQ := &logic.And{Terms: []logic.Expr{p, q}}

	// First call
	r1, err := u.Z3ImpliesBatch(pAndQ, []logic.Expr{p}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r1[0] {
		t.Error("first call: should imply p")
	}

	// Second call should hit cache
	r2, err := u.Z3ImpliesBatch(pAndQ, []logic.Expr{p}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r2[0] {
		t.Error("cached call: should imply p")
	}

	key := [2]logic.NodeKey{pAndQ.Sexp(), p.Sexp()}
	if _, ok := u.ImpliesCache[key]; !ok {
		t.Error("ImpliesCache should contain the entry")
	}
}

// --- Group 9: Python __main__ block (z3_utils.py lines 174-197) ---

// buildTransitivityFixtures creates the logic objects from Python z3_utils.py __main__:
//
//	S = UninterpretedSort('S')
//	X, Y, Z = (Var(n, S) for n in ['X', 'Y', 'Z'])
//	BinRel = FunctionSort(S, S, Boolean)
//	leq = Const('leq', BinRel)
func buildTransitivityFixtures(t *testing.T) (
	leqXY, leqYZ, leqXZ logic.Expr,
	X, Y, Z *logic.Variable,
	leq *logic.Const,
	S *logic.UninterpretedSort,
) {
	t.Helper()
	S = &logic.UninterpretedSort{Name: "S"}
	X = mustVar(t, "X", S)
	Y = mustVar(t, "Y", S)
	Z = mustVar(t, "Z", S)
	binRel := mustFS(t, S, S, logic.Boolean) // S * S -> Boolean
	leq = logic.NewConst("leq", binRel)
	leqXY = mustApply(t, leq, X, Y)
	leqYZ = mustApply(t, leq, Y, Z)
	leqXZ = mustApply(t, leq, X, Z)
	return
}

// TestZ3UtilsTransitivityEquivalences ports Python z3_utils.py __main__ lines 179-186:
//
//	transitive1 = ForAll((X,Y,Z), Implies(And(leq(X,Y), leq(Y,Z)), leq(X,Z)))
//	transitive2 = ForAll((X,Y,Z), Or(Not(leq(X,Y)), Not(leq(Y,Z)), leq(X,Z)))
//	transitive3 = Not(Exists((X,Y,Z), And(leq(X,Y), leq(Y,Z), Not(leq(X,Z)))))
//
//	z3_implies(transitive1, transitive2) == True
//	z3_implies(transitive2, transitive3) == True
//	z3_implies(transitive3, transitive1) == True
func TestZ3UtilsTransitivityEquivalences(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	leqXY, leqYZ, leqXZ, X, Y, Z, _, _ := buildTransitivityFixtures(t)
	vars := []*logic.Variable{X, Y, Z}

	// transitive1: ForAll(X,Y,Z, Implies(And(leq(X,Y), leq(Y,Z)), leq(X,Z)))
	andPremise := &logic.And{Terms: []logic.Expr{leqXY, leqYZ}}
	imp := &logic.Implies{T1: andPremise, T2: leqXZ}
	transitive1 := mustForAll(t, vars, imp)

	// transitive2: ForAll(X,Y,Z, Or(Not(leq(X,Y)), Not(leq(Y,Z)), leq(X,Z)))
	notXY := &logic.Not{Body: leqXY}
	notYZ := &logic.Not{Body: leqYZ}
	orBody := &logic.Or{Terms: []logic.Expr{notXY, notYZ, leqXZ}}
	transitive2 := mustForAll(t, vars, orBody)

	// transitive3: Not(Exists(X,Y,Z, And(leq(X,Y), leq(Y,Z), Not(leq(X,Z)))))
	notXZ := &logic.Not{Body: leqXZ}
	andInner := &logic.And{Terms: []logic.Expr{leqXY, leqYZ, notXZ}}
	existsInner := mustExists(t, vars, andInner)
	transitive3 := &logic.Not{Body: existsInner}

	// t1 => t2
	r, err := u.Z3Implies(transitive1, transitive2, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Error("transitive1 should imply transitive2")
	}

	// t2 => t3
	r, err = u.Z3Implies(transitive2, transitive3, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Error("transitive2 should imply transitive3")
	}

	// t3 => t1
	r, err = u.Z3Implies(transitive3, transitive1, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Error("transitive3 should imply transitive1")
	}
}

// TestZ3UtilsTransitivityNotAntisymmetric ports Python __main__ line 187:
//
//	z3_implies(transitive3, antisymmetric) == False
func TestZ3UtilsTransitivityNotAntisymmetric(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	leqXY, leqYZ, leqXZ, X, Y, Z, _, S := buildTransitivityFixtures(t)
	vars := []*logic.Variable{X, Y, Z}

	// transitive3
	notXZ := &logic.Not{Body: leqXZ}
	andInner := &logic.And{Terms: []logic.Expr{leqXY, leqYZ, notXZ}}
	existsInner := mustExists(t, vars, andInner)
	transitive3 := &logic.Not{Body: existsInner}

	// antisymmetric: ForAll(X,Y, Implies(And(leq(X,Y), leq(Y,X), true), Eq(Y,X)))
	binRel := mustFS(t, S, S, logic.Boolean)
	leq := logic.NewConst("leq", binRel)
	leqYX := mustApply(t, leq, Y, X)
	trueVal := &logic.And{Terms: []logic.Expr{}} // logic.True
	andBody := &logic.And{Terms: []logic.Expr{leqXY, leqYX, trueVal}}
	eqYX := mustEq(t, Y, X)
	impBody := &logic.Implies{T1: andBody, T2: eqYX}
	antisymmetric := mustForAll(t, []*logic.Variable{X, Y}, impBody)

	r, err := u.Z3Implies(transitive3, antisymmetric, false)
	if err != nil {
		t.Fatal(err)
	}
	if r {
		t.Error("transitivity should NOT imply antisymmetry")
	}
}

// TestZ3UtilsIffEquivalence ports Python __main__ line 190:
//
//	z3_implies(true, Iff(transitive1, transitive2)) == True
func TestZ3UtilsIffEquivalence(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	leqXY, leqYZ, leqXZ, X, Y, Z, _, _ := buildTransitivityFixtures(t)
	vars := []*logic.Variable{X, Y, Z}

	// transitive1
	andPremise := &logic.And{Terms: []logic.Expr{leqXY, leqYZ}}
	imp := &logic.Implies{T1: andPremise, T2: leqXZ}
	transitive1 := mustForAll(t, vars, imp)

	// transitive2
	notXY := &logic.Not{Body: leqXY}
	notYZ := &logic.Not{Body: leqYZ}
	orBody := &logic.Or{Terms: []logic.Expr{notXY, notYZ, leqXZ}}
	transitive2 := mustForAll(t, vars, orBody)

	trueVal := &logic.And{Terms: []logic.Expr{}}
	iff := &logic.Iff{T1: transitive1, T2: transitive2}

	r, err := u.Z3Implies(trueVal, iff, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Error("true should imply Iff(transitive1, transitive2)")
	}
}

// TestZ3UtilsIteImplications ports Python __main__ lines 193-197:
//
//	b ⊨ Eq(Ite(b, x, y), x)           → true
//	¬b ⊨ Eq(Ite(b, x, y), y)          → true
//	¬Eq(x,y) ⊨ Iff(Eq(Ite(b,x,y),x), b) → true
func TestZ3UtilsIteImplications(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	S := &logic.UninterpretedSort{Name: "S"}
	x := logic.NewConst("x", S)
	y := logic.NewConst("y", S)
	b := logic.NewConst("b", logic.Boolean)
	ite := mustIte(t, b, x, y) // Ite(b, x, y)

	// Test 1: b ⊨ Eq(Ite(b, x, y), x)
	eqIteX := mustEq(t, ite, x)
	r, err := u.Z3Implies(b, eqIteX, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Error("b should imply Eq(Ite(b,x,y), x)")
	}

	// Test 2: ¬b ⊨ Eq(Ite(b, x, y), y)
	notB := &logic.Not{Body: b}
	eqIteY := mustEq(t, ite, y)
	r, err = u.Z3Implies(notB, eqIteY, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Error("Not(b) should imply Eq(Ite(b,x,y), y)")
	}

	// Test 3: ¬Eq(x,y) ⊨ Iff(Eq(Ite(b,x,y),x), b)
	eqXY := mustEq(t, x, y)
	notEqXY := &logic.Not{Body: eqXY}
	iffEqB := &logic.Iff{T1: eqIteX, T2: b}
	r, err = u.Z3Implies(notEqXY, iffEqB, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Error("Not(Eq(x,y)) should imply Iff(Eq(Ite(b,x,y),x), b)")
	}
}

// --- Group 10: Free Variable Sharing ---

func TestZ3UtilsFreeVarsShared(t *testing.T) {
	u := NewZ3Utils()
	defer u.Close()

	S := &logic.UninterpretedSort{Name: "S"}
	X := mustVar(t, "X", S)
	fs := mustFS(t, S, logic.Boolean)
	r := logic.NewConst("r", fs)

	// premise: r(X)
	premiseApp := mustApply(t, r, X)
	// formula: r(X) — same X should be shared
	formulaApp := mustApply(t, r, X)

	result, err := u.Z3Implies(premiseApp, formulaApp, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("r(X) should imply r(X) with shared free variable X")
	}
}
