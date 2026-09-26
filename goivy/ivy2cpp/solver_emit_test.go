package ivy2cpp

import (
	"os"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func TestEmitSetSolverLargeFunctionEmitsForall(t *testing.T) {
	idx := &goivy.RangeSort{
		Name: "idx",
		Lb:   goivy.NumeralBound{Value: "0"},
		Ub:   goivy.NumeralBound{Value: "1024"},
	}
	fnSort, err := goivy.NewFunctionSort(idx, idx)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	mod := goivy.New()
	if err := mod.Sig.AddSort(idx); err != nil {
		t.Fatalf("AddSort(idx): %v", err)
	}
	mod.Functions.Set(goivy.FunctionKey("bigf", fnSort), fnSort)
	g := &Generator{Mod: mod, ClassName: "bigf", Config: Config{Target: "test", ClassName: "bigf"}}
	sym := stateSymbol{Name: "bigf", Sort: fnSort}
	if !g.isLargeType(sym.Sort) {
		t.Fatalf("expected idx cardinality 1025 to exceed largeThresh=%d", largeThresh)
	}

	var w cppWriter
	g.emitSetSolver(&w, sym, "obj")
	body := w.String()
	for _, want := range []string{
		"std::vector<z3::expr> __quants;;",
		`__quants.push_back(ctx.constant("X__0",sort("idx")));;`,
		`slvr.add(forall(__quants,__to_solver(*this,apply("bigf", ctx.constant("X__0", sort("idx"))),obj.bigf)));`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in large-function emit_set:\n%s", want, body)
		}
	}
	if strings.Contains(body, "for (") {
		t.Fatalf("large-function emit_set should not unroll the domain:\n%s", body)
	}
}

func TestEmitSetSolverLargeRelationUsesVectorApplyForAritySix(t *testing.T) {
	idx := &goivy.RangeSort{
		Name: "idx",
		Lb:   goivy.NumeralBound{Value: "0"},
		Ub:   goivy.NumeralBound{Value: "3"},
	}
	fnSort, err := goivy.NewFunctionSort(idx, idx, idx, idx, idx, idx, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	mod := goivy.New()
	if err := mod.Sig.AddSort(idx); err != nil {
		t.Fatalf("AddSort(idx): %v", err)
	}
	mod.Functions.Set(goivy.FunctionKey("sixrel", fnSort), fnSort)
	g := &Generator{Mod: mod, ClassName: "sixrel", Config: Config{Target: "test", ClassName: "sixrel"}}
	sym := stateSymbol{Name: "sixrel", Sort: fnSort}
	if !g.isLargeType(sym.Sort) {
		t.Fatalf("expected idx^6 cardinality to exceed largeThresh=%d", largeThresh)
	}

	var w cppWriter
	g.emitSetSolver(&w, sym, "obj")
	body := w.String()
	for _, want := range []string{
		"std::vector<z3::expr> __ivy_apply_args;",
		`__ivy_apply_args.push_back(ctx.constant("X__5", sort("idx")));`,
		`return apply("sixrel", __ivy_apply_args);`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in large relation emit_set:\n%s", want, body)
		}
	}
	if strings.Contains(body, `apply("sixrel", ctx.constant("X__0", sort("idx")), ctx.constant("X__1", sort("idx")), ctx.constant("X__2", sort("idx")), ctx.constant("X__3", sort("idx")), ctx.constant("X__4", sort("idx")), ctx.constant("X__5", sort("idx")))`) {
		t.Fatalf("arity-six apply should not use unsupported fixed-arity overload:\n%s", body)
	}
}

func TestEmitSetSolverLargeFunctionUsesThunkToZ3(t *testing.T) {
	out := generateHashThunkSolverFixture(t, "thunkz3")
	for _, want := range []string{
		"template<typename R> class to_solver_class<hash_thunk<int,R> > {",
		"dynamic_cast<z3_thunk<int,R> *>(val.fun)->to_z3(g, v)",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in generated impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestEmitSetSolverHashThunkSpecializationEmittedOnce(t *testing.T) {
	out := generateHashThunkSolverFixture(t, "dedup")
	expectCount(t, out.Impl, "template<typename R> class to_solver_class<hash_thunk<int,R> > {", 1)
	expectCount(t, out.Impl, "template<typename R> z3::expr __to_solver(gen &g, const z3::expr &v, hash_thunk<int,R> &val) {", 0)
	expectCount(t, out.Impl, "template<typename R> z3::expr __to_solver(gen &g, const z3::expr &v, const hash_thunk<int,R> &val) {", 0)
}

func TestEmitSetSolverHashThunkCTupleSpecialization(t *testing.T) {
	out := generateHashThunkSolverFixture(t, "tuplez3")
	for _, want := range []string{
		"template<typename R> class to_solver_class<hash_thunk<tuplez3::__tup__int__int,R> > {",
		"auto __key = it->first;",
		"z3::expr cond = __to_solver(g, v.arg(0), __key.arg0) && __to_solver(g, v.arg(1), __key.arg1);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in generated impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestSolverEmitCommentNotStale(t *testing.T) {
	b, err := os.ReadFile("solver_emit.go")
	if err != nil {
		t.Fatalf("read solver_emit.go: %v", err)
	}
	src := string(b)
	for _, bad := range []string{"Deferred to milestone 5", "uses make_thunk infrastructure"} {
		if strings.Contains(src, bad) {
			t.Fatalf("solver_emit.go still contains stale comment %q", bad)
		}
	}
}

func generateHashThunkSolverFixture(t *testing.T, className string) *Output {
	t.Helper()
	mod := compileIvySource(t, `#lang ivy1.7
type key
relation seen(K:key)
relation touched(K:key)
relation pair(K:key, L:key)
after init {
    seen(K) := false;
    touched(K) := false;
    pair(K,L) := false
}
action mark = {}
export mark
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: className})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return out
}

func expectCount(t *testing.T, text, needle string, want int) {
	t.Helper()
	if got := strings.Count(text, needle); got != want {
		t.Fatalf("count(%q) = %d, want %d:\n%s", needle, got, want, text)
	}
}
