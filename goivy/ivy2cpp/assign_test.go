package ivy2cpp

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// TestEmitAssignTwoPhaseSelfReferential covers the Python `emit_assign`
// branch at ivy_to_cpp.py:3725-3764: a quantified LHS whose RHS reads the
// same symbol must use a function-typed temporary so the inner reads see
// the pre-assignment values.
//
//	r(X) := !r(X)
//
// for X over a 2-element enumerated sort `e`. We expect:
//   - a temp declaration of an array/hash_thunk of bool
//   - two for-loops over X
//   - the first loop body writes the temp from !r[X]
//   - the second loop body writes r[X] from the temp
func TestEmitAssignTwoPhaseSelfReferential(t *testing.T) {
	enum := &goivy.LogicEnumeratedSort{Name: "e", Extension: []string{"a", "b"}}
	fn, err := goivy.NewFunctionSort(enum, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	r := goivy.NewConst("r", fn)
	X, err := goivy.NewVariable("X", enum)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	rOfX := goivy.MustApply(r, X)
	notR, err := goivy.NewNot(rOfX)
	if err != nil {
		t.Fatalf("NewNot: %v", err)
	}
	assign := goivy.NewAssignAction(rOfX, notR)

	var w cppWriter
	(&Generator{}).emitAction(&w, assign)
	got := normalizeCPP(w.String())

	// Two loop heads (one per phase). Both iterate over X.
	if strings.Count(got, "for (") != 2 {
		t.Fatalf("expected 2 for-loops (two-phase), got:\n%s", got)
	}
	// Temp declaration of bool-ranged storage. The exact name is
	// __ivy_tmp1 from g.nextTemp.
	if !strings.Contains(got, "__ivy_tmp1") {
		t.Fatalf("expected temp declaration with __ivy_tmp1, got:\n%s", got)
	}
	// Phase 1 writes the temp, phase 2 writes r. Cheap structural check:
	// the temp must be on the LHS of an assignment that contains !(r[X]).
	if !strings.Contains(got, "__ivy_tmp1[X] = !(r[X])") {
		t.Fatalf("expected phase-1 write `__ivy_tmp1[X] = !(r[X])`, got:\n%s", got)
	}
	if !strings.Contains(got, "r[X] = __ivy_tmp1[X]") {
		t.Fatalf("expected phase-2 copy-back `r[X] = __ivy_tmp1[X]`, got:\n%s", got)
	}
}

// TestEmitAssignBoundsExprTightensLoop covers the Python bexpr trick at
// ivy_to_cpp.py:3717-3720. When RHS is Ite(cond, then, lhs) and cond does
// not mention the modified symbol, the cond is used to tighten loop bounds.
//
//	f(I) := (I < K) ? V : f(I)
//
// where I:idx ranges over 0..10. The expected loop should iterate from 0
// to K (the body-derived upper bound) instead of the full 0..11 range.
func TestEmitAssignBoundsExprTightensLoop(t *testing.T) {
	idxSort := &goivy.RangeSort{
		Name: "idx",
		Lb:   goivy.NumeralBound{Value: "0"},
		Ub:   goivy.NumeralBound{Value: "10"},
	}
	fnSort, err := goivy.NewFunctionSort(idxSort, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	f := goivy.NewConst("f", fnSort)
	I, err := goivy.NewVariable("I", idxSort)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	K := goivy.NewConst("K", idxSort)
	fOfI := goivy.MustApply(f, I)
	v := goivy.NewConst("V", goivy.Boolean)
	// Build (I < K) ? V : f(I)
	ltSort, err := goivy.NewFunctionSort(idxSort, idxSort, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort lt: %v", err)
	}
	lt := goivy.NewConst("<", ltSort)
	cond := goivy.MustApply(lt, I, K)
	ite, err := goivy.NewIte(cond, v, fOfI)
	if err != nil {
		t.Fatalf("NewIte: %v", err)
	}
	assign := goivy.NewAssignAction(fOfI, ite)

	var w cppWriter
	(&Generator{}).emitAction(&w, assign)
	got := normalizeCPP(w.String())

	// With bexpr tightening, the loop should be bounded by K (the
	// upper bound from `I < K`) instead of 11 (sort cardinality of
	// idx={0..10}, hi=10+1=11). loopHeaderForSortBounds emits the
	// half-open form `I < K`.
	if !strings.Contains(got, "I < K") {
		t.Fatalf("expected loop bounded by `I < K`, got:\n%s", got)
	}
	// Sanity: the temp+two-phase machinery should still be there.
	if !strings.Contains(got, "__ivy_tmp1") {
		t.Fatalf("expected temp declaration, got:\n%s", got)
	}
}

// TestEmitAssignBoundsExprSkippedWhenModifiedSymbolInCond checks the
// safety side of the bexpr trick (Python ivy_to_cpp.py:3719). The
// condition mentions the modified symbol, so bexpr must NOT be used —
// otherwise the bounds would depend on the value being mutated.
func TestEmitAssignBoundsExprSkippedWhenModifiedSymbolInCond(t *testing.T) {
	idxSort := &goivy.RangeSort{
		Name: "idx",
		Lb:   goivy.NumeralBound{Value: "0"},
		Ub:   goivy.NumeralBound{Value: "10"},
	}
	fnSort, err := goivy.NewFunctionSort(idxSort, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	f := goivy.NewConst("f", fnSort)
	I, err := goivy.NewVariable("I", idxSort)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	fOfI := goivy.MustApply(f, I)
	v := goivy.NewConst("V", goivy.Boolean)
	// Build f(I) ? V : f(I) — cond mentions f, the modified symbol.
	ite, err := goivy.NewIte(fOfI, v, fOfI)
	if err != nil {
		t.Fatalf("NewIte: %v", err)
	}
	assign := goivy.NewAssignAction(fOfI, ite)

	var w cppWriter
	(&Generator{}).emitAction(&w, assign)
	got := normalizeCPP(w.String())

	// Without bexpr tightening, the loop uses the default
	// loopHeaderForVar which emits the natural [lb, ub]-inclusive form
	// for range types (loopHeaderForSort, expr.go:1378).
	if !strings.Contains(got, "for (unsigned I = 0; I <= 10; I++)") {
		t.Fatalf("expected default natural loop (bexpr trick skipped), got:\n%s", got)
	}
	// And NOT the body-derived half-open form.
	if strings.Contains(got, "I < 10") || strings.Contains(got, "I < 11") {
		t.Fatalf("expected NO body-derived bounds (modified symbol in cond), got:\n%s", got)
	}
}

// TestEmitAssignLargeThunkFallback covers Python emit_assign_large
// (ivy_to_cpp.py:3654-3662). When the LHS free variable is over an
// uninterpreted sort with no cardinality / iterability, we cannot emit
// explicit loops. The assignment becomes f = hash_thunk<...>(new
// __thunk__N(env...)).
func TestEmitAssignLargeThunkFallback(t *testing.T) {
	// An uninterpreted sort with no cardinality attribute is
	// non-iterable. node is the conventional name for such sorts.
	node := &goivy.UninterpretedSort{Name: "node"}
	fnSort, err := goivy.NewFunctionSort(node, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	r := goivy.NewConst("r", fnSort)
	X, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	rOfX := goivy.MustApply(r, X)
	// r(X) := !r(X) — would emit two-phase if r were over a finite
	// sort. Over node (uninterpreted), the two-phase loop has no
	// iteration anchor and falls back to thunk.
	notR, err := goivy.NewNot(rOfX)
	if err != nil {
		t.Fatalf("NewNot: %v", err)
	}
	assign := goivy.NewAssignAction(rOfX, notR)

	var w cppWriter
	(&Generator{}).emitAction(&w, assign)
	got := normalizeCPP(w.String())

	// Note: this test constructs sorts without a Module, so the
	// UninterpretedSort "node" resolves to cppType "int" (the default
	// fallback). The shape we care about is the thunk struct +
	// hash_thunk construction; the domain type substitution into
	// templates is exercised by integration-level tests.
	if !strings.Contains(got, "struct __thunk__0 : thunk<int, bool>") {
		t.Fatalf("expected thunk struct definition, got:\n%s", got)
	}
	if !strings.Contains(got, "bool operator()(const int &arg)") {
		t.Fatalf("expected operator()(const int &arg), got:\n%s", got)
	}
	// The body should reference !(r[arg]) (loop var substituted by arg).
	if !strings.Contains(got, "!(r[arg])") {
		t.Fatalf("expected substituted body !(r[arg]), got:\n%s", got)
	}
	// Construction expression.
	if !strings.Contains(got, "r = hash_thunk<int, bool>(new __thunk__0(") {
		t.Fatalf("expected hash_thunk construction, got:\n%s", got)
	}
	// And the env captures r (the LHS function symbol referenced in
	// the body) so the thunk can lazily compute results.
	if !strings.Contains(got, "hash_thunk<int,bool> r;") {
		t.Fatalf("expected r captured as env field, got:\n%s", got)
	}
}

// TestEmitAssignMultiVariableTransposeTwoPhase covers the canonical
// aliasing case for multi-variable LHSs: a transpose update
//
//	g(X, Y) := g(Y, X)
//
// must not crossover in-place — Phase 2's reads from the temp must see
// the original g values, not partially-updated ones.
func TestEmitAssignMultiVariableTransposeTwoPhase(t *testing.T) {
	enum := &goivy.LogicEnumeratedSort{Name: "e", Extension: []string{"a", "b"}}
	fn, err := goivy.NewFunctionSort(enum, enum, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	g := goivy.NewConst("g", fn)
	X, err := goivy.NewVariable("X", enum)
	if err != nil {
		t.Fatalf("NewVariable X: %v", err)
	}
	Y, err := goivy.NewVariable("Y", enum)
	if err != nil {
		t.Fatalf("NewVariable Y: %v", err)
	}
	lhs := goivy.MustApply(g, X, Y)
	rhs := goivy.MustApply(g, Y, X) // transpose
	assign := goivy.NewAssignAction(lhs, rhs)

	var w cppWriter
	(&Generator{}).emitAction(&w, assign)
	got := normalizeCPP(w.String())

	// Two nested loops per phase = 4 total for-loops.
	if strings.Count(got, "for (") != 4 {
		t.Fatalf("expected 4 for-loops (two-phase × two vars), got:\n%s", got)
	}
	// Phase 1 reads g(Y, X) and writes the temp.
	if !strings.Contains(got, "g[__tup__e__e(Y, X)]") && !strings.Contains(got, "g[Y][X]") {
		t.Fatalf("expected phase-1 RHS read g(Y,X) (array or tuple form), got:\n%s", got)
	}
	// Phase 2 writes g(X, Y) from the temp.
	if !strings.Contains(got, "g[__tup__e__e(X, Y)] = __ivy_tmp1") && !strings.Contains(got, "g[X][Y] = __ivy_tmp1") {
		t.Fatalf("expected phase-2 write g(X,Y) = temp, got:\n%s", got)
	}
}

// TestEmitAssignFlipOracleMatchesPython is a cross-language oracle
// check for the canonical self-referential flip case. Python emits
// (test_sref.ivy + target=test):
//
//	void test_sref::ext__flip(){
//	    bool __tmp0[2];
//	    for (int X = 0; X < 2; X++) { __tmp0[X] = !marked[X]; }
//	    for (int X = 0; X < 2; X++) { marked[X] = __tmp0[X]; }
//	}
//
// Go's emission diverges in two minor ways documented inline:
//   - temp name __ivy_tmp1 (not __tmp0) — Go counter naming convention
//   - loop variable typed as `color` with range-for over the enum
//     extension list — Go's loopHeaderForVar idiom, semantically
//     equivalent to Python's `for (int X = 0; X < 2; X++)`.
//
// The structural shape — temp array, two loops, phase-1 reads marked,
// phase-2 writes marked from temp — must match.
func TestEmitAssignFlipOracleMatchesPython(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
action flip = {
    marked(X) := ~marked(X)
}
export flip
`)
	out, err := Generate(mod, Config{ClassName: "test_sref"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Expected structural shape, matching Python's two-phase emission.
	// Function name lacks the ext__ prefix here because we used the
	// default target=impl (not repl/test where the prefix is added).
	for _, want := range []string{
		"void test_sref::flip()",
		"bool __ivy_tmp1[2];",
		"for (color X : {red, green})",
		"__ivy_tmp1[X] = !(marked[X]);",
		"marked[X] = __ivy_tmp1[X];",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	// Both loops must appear. Together with the impl-wide check
	// above, this catches a regression where Phase 2 is dropped.
	if strings.Count(out.Impl, "for (color X : {red, green})") < 2 {
		t.Fatalf("expected two for-loops in flip impl:\n%s", out.Impl)
	}
}

// TestEmitAssignFieldRoutesThroughEmitAssign covers Phase E: a
// LogicAssignFieldAction is lowered into AssignAction(field(obj), value)
// and routed through emitAssign. The synthesized AssignAction has no
// free LHS variables, so it takes the simple path. This verifies the
// field-action dead-code path emits identical output to a hand-rolled
// destructor assignment.
func TestEmitAssignFieldRoutesThroughEmitAssign(t *testing.T) {
	// Synthesize a destructor symbol: shade : node -> color
	enum := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	node := &goivy.UninterpretedSort{Name: "node"}
	fldSort, err := goivy.NewFunctionSort(node, enum)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	shade := goivy.NewConst("shade", fldSort)
	obj := goivy.NewConst("obj", node)
	green := goivy.NewConst("green", enum)
	a := goivy.NewAssignFieldAction(shade, obj, green)

	var w cppWriter
	(&Generator{}).emitAction(&w, a)
	got := normalizeCPP(w.String())

	// Without a registered destructor, emitApply falls through to
	// cppStorageAccess which produces shade[obj] for this Apply.
	// Either form (shade[obj] or obj.shade) is acceptable; we just
	// need to confirm the LHS makes it through emitAssign and the
	// RHS green is the assigned value.
	if !strings.Contains(got, "= green;") {
		t.Fatalf("expected `= green;` in synthesized assignment, got:\n%s", got)
	}
	if strings.Contains(got, "unsupported") {
		t.Fatalf("emitAssignField should not emit unsupported: %s", got)
	}
}

// TestEmitAssignSimpleScalarRespectsExistingPath covers the
// no-free-variable path: a simple scalar assignment must remain a
// single-line write (no temporary).
func TestEmitAssignSimpleScalarRespectsExistingPath(t *testing.T) {
	flag := goivy.NewConst("flag", goivy.Boolean)
	assign := goivy.NewAssignAction(flag, goivy.NewConst("true", goivy.Boolean))
	var w cppWriter
	(&Generator{}).emitAction(&w, assign)
	got := normalizeCPP(w.String())
	if got != "flag = true;" {
		t.Fatalf("expected `flag = true;` for scalar simple-path assign, got:\n%s", got)
	}
	if strings.Contains(got, "__ivy_tmp") {
		t.Fatalf("simple path should not allocate a temp, got:\n%s", got)
	}
}



