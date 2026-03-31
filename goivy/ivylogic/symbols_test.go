package ivylogic

import (
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// helpers to build AST nodes for tests

func testSort() lg.Sort {
	return &lg.UninterpretedSort{Name: "S"}
}

func testConst(name string) *lg.Const {
	return lg.NewConst(name, testSort())
}

func testVar(name string) *lg.Variable {
	v, _ := lg.NewVariable(name, testSort())
	if v == nil {
		// Variable names must start uppercase in ivy logic.
		// If the caller passed lowercase, build directly.
		return &lg.Variable{Name: name, VSort: testSort()}
	}
	return v
}

func testApply(fn lg.Expr, terms ...lg.Expr) *lg.Apply {
	return &lg.Apply{Func: fn, Terms: terms}
}

func testForAll(vars []*lg.Variable, body lg.Expr) *lg.ForAll {
	// Build directly to avoid sort-checking in NewForAll.
	return &lg.ForAll{Variables: vars, Body: body}
}

func testLambda(vars []*lg.Variable, body lg.Expr) *lg.Lambda {
	return &lg.Lambda{Variables: vars, Body: body}
}

func testExists(vars []*lg.Variable, body lg.Expr) *lg.Exists {
	return &lg.Exists{Variables: vars, Body: body}
}

// collectSyms collects all yielded symbols from SymbolsIluAst into a slice.
func collectSyms(node lg.Expr) []lg.Expr {
	var result []lg.Expr
	for sym := range SymbolsIluAst(node) {
		result = append(result, sym)
	}
	return result
}

// symNames extracts names from a slice of Expr (assumes all are *lg.Const).
func symNames(syms []lg.Expr) []string {
	names := make([]string, len(syms))
	for i, s := range syms {
		if c, ok := s.(*lg.Const); ok {
			names[i] = c.Name
		} else {
			names[i] = s.String()
		}
	}
	return names
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- NodeRep tests ---

func TestNodeRep_Apply(t *testing.T) {
	f := testConst("f")
	a := testConst("a")
	app := testApply(f, a)
	rep := NodeRep(app)
	if rep != f {
		t.Errorf("NodeRep(Apply) should return Func, got %v", rep)
	}
}

func TestNodeRep_Const(t *testing.T) {
	c := testConst("c")
	rep := NodeRep(c)
	if rep != c {
		t.Errorf("NodeRep(Const) should return self, got %v", rep)
	}
}

func TestNodeRep_NamedBinder0Vars(t *testing.T) {
	nb := &lg.NamedBinder{
		Name:      "nb",
		Variables: nil,
		Body:      testConst("body"),
	}
	rep := NodeRep(nb)
	if rep != nb {
		t.Errorf("NodeRep(NamedBinder with 0 vars) should return self, got %v", rep)
	}
}

func TestNodeRep_NamedBinderWithVars(t *testing.T) {
	v := testVar("x")
	nb := &lg.NamedBinder{
		Name:      "nb",
		Variables: []*lg.Variable{v},
		Body:      testConst("body"),
	}
	rep := NodeRep(nb)
	if rep != nil {
		t.Errorf("NodeRep(NamedBinder with vars) should return nil, got %v", rep)
	}
}

func TestNodeRep_Variable(t *testing.T) {
	v := testVar("x")
	rep := NodeRep(v)
	if rep != nil {
		t.Errorf("NodeRep(Variable) should return nil, got %v", rep)
	}
}

// --- SymbolsIluAst tests ---

func TestSymbolsIluAst_SimpleConst(t *testing.T) {
	// A bare constant is an app; its rep is itself.
	c := testConst("c")
	syms := collectSyms(c)
	names := symNames(syms)
	expected := []string{"c"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_SimpleApply(t *testing.T) {
	// f(a, b) should yield f, a, b
	f := testConst("f")
	a := testConst("a")
	b := testConst("b")
	app := testApply(f, a, b)
	syms := collectSyms(app)
	names := symNames(syms)
	expected := []string{"f", "a", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_NestedApply(t *testing.T) {
	// f(g(a), b) should yield f, g, a, b
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	inner := testApply(g, a)
	outer := testApply(f, inner, b)
	syms := collectSyms(outer)
	names := symNames(syms)
	expected := []string{"f", "g", "a", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_BinderFunc(t *testing.T) {
	// Apply with Lambda as Func: Apply{Func: Lambda{vars:[x], body: f(x)}, Terms: [a]}
	// Should NOT yield the Lambda. Should recurse into Lambda.Body → yield f, x.
	// Then recurse into Terms → yield a.
	// Result: f, x, a
	f := testConst("f")
	x := testConst("x")
	a := testConst("a")
	fOfX := testApply(f, x)
	v := testVar("v")
	lam := testLambda([]*lg.Variable{v}, fOfX)
	app := testApply(lam, a)

	syms := collectSyms(app)
	names := symNames(syms)
	expected := []string{"f", "x", "a"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_ForAllFormula(t *testing.T) {
	// ForAll([v], f(a) & g(b))
	// ForAll is not an app, so no rep yielded.
	// NodeArgs(ForAll) = [body] = [And(f(a), g(b))]
	// And is not an app → recurse into And's args: f(a), g(b)
	// f(a): yield f, then arg a → yield a
	// g(b): yield g, then arg b → yield b
	// Result: f, a, g, b
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	v := testVar("x")
	fOfA := testApply(f, a)
	gOfB := testApply(g, b)
	body := &lg.And{Terms: []lg.Expr{fOfA, gOfB}}
	fa := testForAll([]*lg.Variable{v}, body)

	syms := collectSyms(fa)
	names := symNames(syms)
	expected := []string{"f", "a", "g", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_NestedBinders(t *testing.T) {
	// ForAll([v], Apply{Func: Lambda{vars:[w], body: f(c)}, Terms: [a]})
	// ForAll args = [body] where body = Apply{Lambda, [a]}
	// Apply is an app. rep = Lambda. IsBinder(Lambda) → true.
	// Recurse into Lambda.Body = f(c): yield f, c.
	// Then Apply args = [a]: yield a.
	// Result: f, c, a
	f := testConst("f")
	c := testConst("c")
	a := testConst("a")
	v := testVar("v")
	w := testVar("w")
	fOfC := testApply(f, c)
	lam := testLambda([]*lg.Variable{w}, fOfC)
	app := testApply(lam, a)
	fa := testForAll([]*lg.Variable{v}, app)

	syms := collectSyms(fa)
	names := symNames(syms)
	expected := []string{"f", "c", "a"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_Variable(t *testing.T) {
	// Variable is not an app and has no args. Should yield nothing.
	v := testVar("x")
	syms := collectSyms(v)
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols from Variable, got %d: %v", len(syms), symNames(syms))
	}
}

func TestSymbolsIluAst_And(t *testing.T) {
	// And(f(a), g(b)) — And is not an app, args are [f(a), g(b)]
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	and := &lg.And{Terms: []lg.Expr{testApply(f, a), testApply(g, b)}}

	syms := collectSyms(and)
	names := symNames(syms)
	expected := []string{"f", "a", "g", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_Eq(t *testing.T) {
	// Eq(a, b) — Eq is not an app. Args = [a, b]. Each is a Const (app), yield each.
	a := testConst("a")
	b := testConst("b")
	eq := &lg.Eq{T1: a, T2: b}

	syms := collectSyms(eq)
	names := symNames(syms)
	expected := []string{"a", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_NamedBinder0VarsAsApp(t *testing.T) {
	// NamedBinder with 0 vars is treated as an app by IsApp.
	// Its rep is itself. IsBinder(NamedBinder) → true.
	// So recurse into body instead of yielding the binder.
	// body = f(a). yield f, a.
	// NodeArgs(NamedBinder) = [body], so recurse into body again → f, a again.
	// Total: f, a, f, a (the Python does this double-visit too).
	f := testConst("f")
	a := testConst("a")
	fOfA := testApply(f, a)
	nb := &lg.NamedBinder{
		Name:      "nb",
		Variables: nil,
		Body:      fOfA,
	}

	syms := collectSyms(nb)
	names := symNames(syms)
	expected := []string{"f", "a", "f", "a"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_Not(t *testing.T) {
	// Not(f(a)) — Not is not an app. Args = [body] = [f(a)].
	f := testConst("f")
	a := testConst("a")
	not := &lg.Not{Body: testApply(f, a)}

	syms := collectSyms(not)
	names := symNames(syms)
	expected := []string{"f", "a"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_Implies(t *testing.T) {
	// Implies(f(a), g(b))
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	imp := &lg.Implies{T1: testApply(f, a), T2: testApply(g, b)}

	syms := collectSyms(imp)
	names := symNames(syms)
	expected := []string{"f", "a", "g", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_Ite(t *testing.T) {
	// Ite(p, f(a), g(b))
	p := testConst("p")
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	ite := &lg.Ite{Cond: p, Then: testApply(f, a), Else: testApply(g, b)}

	syms := collectSyms(ite)
	names := symNames(syms)
	expected := []string{"p", "f", "a", "g", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_ExistsFormula(t *testing.T) {
	// Exists([v], f(a)) → recurse into body f(a), yield f, a
	f := testConst("f")
	a := testConst("a")
	v := testVar("x")
	ex := testExists([]*lg.Variable{v}, testApply(f, a))

	syms := collectSyms(ex)
	names := symNames(syms)
	expected := []string{"f", "a"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_ApplyWithForAllFunc(t *testing.T) {
	// Apply{Func: ForAll{vars:[v], body: g(c)}, Terms: [a]}
	// Apply is an app. rep = ForAll. IsBinder(ForAll) → true.
	// Recurse into ForAll.Body = g(c): yield g, c.
	// Then Apply.Terms = [a]: yield a.
	// Result: g, c, a
	g := testConst("g")
	c := testConst("c")
	a := testConst("a")
	v := testVar("v")
	gOfC := testApply(g, c)
	fa := testForAll([]*lg.Variable{v}, gOfC)
	app := testApply(fa, a)

	syms := collectSyms(app)
	names := symNames(syms)
	expected := []string{"g", "c", "a"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestSymbolsIluAst_EarlyStop(t *testing.T) {
	// Verify that the iterator can be stopped early.
	f := testConst("f")
	a := testConst("a")
	b := testConst("b")
	app := testApply(f, a, b)

	var first lg.Expr
	for sym := range SymbolsIluAst(app) {
		first = sym
		break // stop after first
	}
	if c, ok := first.(*lg.Const); !ok || c.Name != "f" {
		t.Errorf("first symbol should be f, got %v", first)
	}
}

// --- UsedSymbolsAst tests ---

func TestUsedSymbolsAst_Dedup(t *testing.T) {
	// f(a, a) → set should have {f, a} (2 entries, not 3)
	f := testConst("f")
	a := testConst("a")
	app := testApply(f, a, a)

	syms := UsedSymbolsAst(app)
	if len(syms) != 2 {
		t.Errorf("expected 2 unique symbols, got %d", len(syms))
	}
	if _, ok := syms[lg.Key(f)]; !ok {
		t.Error("set should contain f")
	}
	if _, ok := syms[lg.Key(a)]; !ok {
		t.Error("set should contain a")
	}
}

// --- UsedSymbolsAsts tests ---

func TestUsedSymbolsAsts_MultipleNodes(t *testing.T) {
	// Two ASTs: f(a) and g(b). Union set = {f, a, g, b}.
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	nodes := []lg.Expr{testApply(f, a), testApply(g, b)}

	syms := UsedSymbolsAsts(nodes)
	if len(syms) != 4 {
		t.Errorf("expected 4 unique symbols, got %d", len(syms))
	}
	for _, c := range []*lg.Const{f, g, a, b} {
		if _, ok := syms[lg.Key(c)]; !ok {
			t.Errorf("set should contain %s", c.Name)
		}
	}
}

func TestUsedSymbolsAsts_OverlapDedup(t *testing.T) {
	// f(a) and g(a). Union set = {f, a, g} (a appears in both).
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	nodes := []lg.Expr{testApply(f, a), testApply(g, a)}

	syms := UsedSymbolsAsts(nodes)
	if len(syms) != 3 {
		t.Errorf("expected 3 unique symbols, got %d", len(syms))
	}
}

// --- SymbolsAsts tests ---

func TestSymbolsAsts_PreservesOrder(t *testing.T) {
	// [f(a), g(b)] → f, a, g, b in order
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	nodes := []lg.Expr{testApply(f, a), testApply(g, b)}

	var names []string
	for sym := range SymbolsAsts(nodes) {
		if c, ok := sym.(*lg.Const); ok {
			names = append(names, c.Name)
		}
	}
	expected := []string{"f", "a", "g", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

// --- UsedSymbolsInOrderAst tests ---

func TestUsedSymbolsInOrderAst(t *testing.T) {
	// f(a, g(a, b)) → unique in order: f, a, g, b
	f := testConst("f")
	g := testConst("g")
	a := testConst("a")
	b := testConst("b")
	inner := testApply(g, a, b)
	outer := testApply(f, a, inner)

	syms := UsedSymbolsInOrderAst(outer)
	names := symNames(syms)
	expected := []string{"f", "a", "g", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestUsedSymbolsInOrderAst_DuplicatesRemoved(t *testing.T) {
	// f(a, a, f(b)) → unique in order: f, a, b
	f := testConst("f")
	a := testConst("a")
	b := testConst("b")
	inner := testApply(f, b)
	outer := testApply(f, a, a, inner)

	syms := UsedSymbolsInOrderAst(outer)
	names := symNames(syms)
	expected := []string{"f", "a", "b"}
	if !sliceEqual(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}
