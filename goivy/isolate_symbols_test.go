package goivy

import (
	"math/rand"
	"testing"
)

// test helpers

func tSort() Sort {
	return &UninterpretedSort{Name: "S"}
}

func tConst(name string) *Const {
	return NewConst(name, tSort())
}

func tApply(fn Expr, terms ...Expr) *Apply {
	// Intentional: bypass sort check for test convenience (fn may have non-function sort)
	return &Apply{Func: fn, Terms: terms}
}

func tForAll(body Expr) *ForAll {
	v := &Variable{Name: "V", VSort: tSort()}
	return &ForAll{Variables: []*Variable{v}, Body: body}
}

func tLambda(body Expr) *Lambda {
	v := &Variable{Name: "V", VSort: tSort()}
	return &Lambda{Variables: []*Variable{v}, Body: body}
}

func symSetNames(syms *InsMap[NodeKey, Expr]) map[string]bool {
	names := make(map[string]bool, syms.Len())
	for _, expr := range syms.All() {
		if c, ok := expr.(*Const); ok {
			names[c.Name] = true
		}
	}
	return names
}

// --- collectSymbolsInto tests ---

func TestCollectSymbolsInto_SimpleConst(t *testing.T) {
	c := tConst("c")
	syms := NewInsMap[NodeKey, Expr]()
	collectSymbolsInto("test", c, syms)
	names := symSetNames(syms)
	if !names["c"] {
		t.Error("should contain c")
	}
	if syms.Len() != 1 {
		t.Errorf("expected 1 symbol, got %d", syms.Len())
	}
}

func TestCollectSymbolsInto_Apply(t *testing.T) {
	f := tConst("f")
	a := tConst("a")
	b := tConst("b")
	app := tApply(f, a, b)
	syms := NewInsMap[NodeKey, Expr]()
	collectSymbolsInto("test", app, syms)
	names := symSetNames(syms)
	for _, name := range []string{"f", "a", "b"} {
		if !names[name] {
			t.Errorf("should contain %s", name)
		}
	}
}

func TestCollectSymbolsInto_BinderFunc(t *testing.T) {
	// Apply{Func: Lambda{body: f(x)}, Terms: [a]}
	// The key test: Lambda is a binder, so SymbolsIluAst expands into its body
	// rather than yielding the Lambda itself. Should find f, x, a.
	f := tConst("f")
	x := tConst("x")
	a := tConst("a")
	fOfX := tApply(f, x)
	lam := tLambda(fOfX)
	app := tApply(lam, a)

	syms := NewInsMap[NodeKey, Expr]()
	collectSymbolsInto("test", app, syms)
	names := symSetNames(syms)
	for _, name := range []string{"f", "x", "a"} {
		if !names[name] {
			t.Errorf("should contain %s, got %v", name, names)
		}
	}
}

func TestCollectSymbolsInto_Nil(t *testing.T) {
	syms := NewInsMap[NodeKey, Expr]()
	collectSymbolsInto("test", nil, syms) // must not panic
	if syms.Len() != 0 {
		t.Errorf("expected empty map for nil node, got %d entries", syms.Len())
	}
}

func TestCollectSymbolsInto_MergesIntoExisting(t *testing.T) {
	pre := tConst("pre")
	syms := NewInsMap[NodeKey, Expr]()
	syms.Set(Key(pre), pre)

	f := tConst("f")
	a := tConst("a")
	collectSymbolsInto("test", tApply(f, a), syms)

	if syms.Len() != 3 {
		t.Errorf("expected 3 symbols (pre + f + a), got %d", syms.Len())
	}
	names := symSetNames(syms)
	for _, name := range []string{"pre", "f", "a"} {
		if !names[name] {
			t.Errorf("should contain %s", name)
		}
	}
}

func TestCollectSymbolsInto_ForAllBody(t *testing.T) {
	// ForAll(V, f(a) & g(b))
	f := tConst("f")
	g := tConst("g")
	a := tConst("a")
	b := tConst("b")
	body := &And{Terms: []Expr{tApply(f, a), tApply(g, b)}}
	fa := tForAll(body)

	syms := NewInsMap[NodeKey, Expr]()
	collectSymbolsInto("test", fa, syms)
	names := symSetNames(syms)
	for _, name := range []string{"f", "a", "g", "b"} {
		if !names[name] {
			t.Errorf("should contain %s", name)
		}
	}
}

func TestCollectSymbolsInto_ApplyWithForAllFunc(t *testing.T) {
	// Apply{Func: ForAll{body: g(c)}, Terms: [a]}
	// ForAll is a binder → expand into body. Should find g, c, a.
	g := tConst("g")
	c := tConst("c")
	a := tConst("a")
	fa := tForAll(tApply(g, c))
	app := tApply(fa, a)

	syms := NewInsMap[NodeKey, Expr]()
	collectSymbolsInto("test", app, syms)
	names := symSetNames(syms)
	for _, name := range []string{"g", "c", "a"} {
		if !names[name] {
			t.Errorf("should contain %s, got %v", name, names)
		}
	}
}

// --- Fuzz/random test ---

// randAST generates a random AST tree of the given depth with the given constants.
func randAST(rng *rand.Rand, consts []*Const, depth int) Expr {
	if depth <= 0 || rng.Intn(3) == 0 {
		return consts[rng.Intn(len(consts))]
	}
	switch rng.Intn(5) {
	case 0: // bare const
		return consts[rng.Intn(len(consts))]
	case 1: // Apply with const func
		nArgs := rng.Intn(3)
		args := make([]Expr, nArgs)
		for i := range args {
			args[i] = randAST(rng, consts, depth-1)
		}
		return tApply(consts[rng.Intn(len(consts))], args...)
	case 2: // And
		nTerms := 1 + rng.Intn(3)
		terms := make([]Expr, nTerms)
		for i := range terms {
			terms[i] = randAST(rng, consts, depth-1)
		}
		return &And{Terms: terms}
	case 3: // ForAll
		return tForAll(randAST(rng, consts, depth-1))
	case 4: // Apply with Lambda func (binder case)
		body := randAST(rng, consts, depth-1)
		lam := tLambda(body)
		nArgs := rng.Intn(2)
		args := make([]Expr, nArgs)
		for i := range args {
			args[i] = randAST(rng, consts, depth-1)
		}
		return tApply(lam, args...)
	}
	return consts[0]
}

func TestRandomized_NoPanic(t *testing.T) {
	// Verify collectSymbolsInto never panics on random ASTs
	// and always finds at least one symbol (since randAST always includes Consts).
	consts := []*Const{
		tConst("f"), tConst("g"), tConst("a"), tConst("b"),
		tConst("c"), tConst("h"), tConst("p"), tConst("q"),
	}
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 500; i++ {
		node := randAST(rng, consts, 4)
		syms := NewInsMap[NodeKey, Expr]()
		collectSymbolsInto("test", node, syms)
		if syms.Len() == 0 {
			t.Errorf("iter %d: expected at least one symbol", i)
		}
	}
}

func FuzzCollectSymbolsInto(f *testing.F) {
	f.Add(int64(1), 3)
	f.Add(int64(42), 5)
	f.Add(int64(999), 2)

	consts := []*Const{
		tConst("f"), tConst("g"), tConst("a"), tConst("b"),
		tConst("c"), tConst("h"),
	}

	f.Fuzz(func(t *testing.T, seed int64, depth int) {
		if depth < 0 {
			depth = 0
		}
		if depth > 6 {
			depth = 6
		}
		rng := rand.New(rand.NewSource(seed))
		node := randAST(rng, consts, depth)

		// must not panic
		syms := NewInsMap[NodeKey, Expr]()
		collectSymbolsInto("test", node, syms)

		// every random AST contains at least one Const leaf
		if syms.Len() == 0 {
			t.Error("expected at least one symbol")
		}
	})
}
