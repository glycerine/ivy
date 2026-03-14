package updr

import (
	"container/heap"
	"fmt"
	"math/rand"
	"testing"

	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/z3bridge"
)

// helper: create a fresh context + bool variable pair (x0, xn)
func makeOneBit(t *testing.T) (*z3bridge.Context, z3bridge.Expr, z3bridge.Expr) {
	t.Helper()
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	xn := ctx.Const("xn", bs)
	return ctx, x, xn
}

// helper: create N bool variable pairs
func makeNBits(t *testing.T, n int) (*z3bridge.Context, []z3bridge.Expr, []z3bridge.Expr) {
	t.Helper()
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	x0 := make([]z3bridge.Expr, n)
	xn := make([]z3bridge.Expr, n)
	for i := 0; i < n; i++ {
		x0[i] = ctx.Const(fmt.Sprintf("x%d", i), bs)
		xn[i] = ctx.Const(fmt.Sprintf("xn%d", i), bs)
	}
	return ctx, x0, xn
}

// ---------- Test 1: Trivial safe — bad = false ----------

func TestPDR_TrivialSafe(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)                   // init: x = false
	trans := ctx.Iff(xn, x)              // trans: x' = x (identity)
	bad := ctx.BoolVal(false)             // bad: never
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (bad=false), got invalid")
	}
}

// ---------- Test 2: Trivial unsafe — init ∧ bad satisfiable ----------

func TestPDR_TrivialUnsafe(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)
	trans := ctx.Iff(xn, x)
	bad := ctx.Not(x) // bad = ¬x, same as init
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	result := pdr.Run()
	if result.Valid {
		t.Fatal("expected invalid (init ∧ bad is SAT), got valid")
	}
	if len(result.Trace) == 0 {
		t.Fatal("expected non-empty trace")
	}
}

// ---------- Test 3: One-bit flip, bad = true ----------
// init: x=false, trans: x'=¬x, bad: x=true
// After one step x becomes true => counterexample of length 2.

func TestPDR_OneBitFlip(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)
	trans := ctx.Iff(xn, ctx.Not(x)) // flip
	bad := x                           // bad when x=true
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	result := pdr.Run()
	if result.Valid {
		t.Fatal("expected counterexample (bit flips to bad in 1 step)")
	}
}

// ---------- Test 4: One-bit stays false, bad = true ----------
// init: x=false, trans: x'=false, bad: x=true
// x never becomes true => invariant exists.

func TestPDR_OneBitStaysFalse(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)
	trans := ctx.Not(xn)   // x' always false
	bad := x               // bad when x=true
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (x stays false)")
	}
}

// ---------- Test 5: Two-bit counter reaches bad state ----------
// State: (a, b) encoding 0-3.
// init: (0,0). trans: increment mod 4. bad: (1,1)=3.
// 0->1->2->3 so bad is reachable.

func TestPDR_TwoBitCounterUnsafe(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)
	an := ctx.Const("an", bs)
	bn := ctx.Const("bn", bs)

	// init: a=0, b=0
	init := ctx.And(ctx.Not(a), ctx.Not(b))

	// trans: (an, bn) = (a,b)+1 mod 4
	// b flips every step; a flips when b is true
	trans := ctx.And(
		ctx.Iff(bn, ctx.Not(b)),
		ctx.Iff(an, ctx.Iff(a, b)), // a XOR b => a flips when b=1
	)

	// bad: a=1, b=1 (value 3)
	bad := ctx.And(a, b)

	x0 := []z3bridge.Expr{a, b}
	xn := []z3bridge.Expr{an, bn}

	pdr := NewPDR(ctx, init, trans, bad, x0, nil, xn)
	result := pdr.Run()
	if result.Valid {
		t.Fatal("expected counterexample (counter reaches 3)")
	}
}

// ---------- Test 6: Two-bit, bad=3 unreachable (counter wraps at 2) ----------

func TestPDR_TwoBitCounterSafe(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)
	an := ctx.Const("an", bs)
	bn := ctx.Const("bn", bs)

	// init: (0,0)
	init := ctx.And(ctx.Not(a), ctx.Not(b))

	// trans: b toggles, a stays false (so we only visit 00 and 01)
	trans := ctx.And(
		ctx.Iff(bn, ctx.Not(b)),
		ctx.Not(an), // a' always false
	)

	// bad: a=1
	bad := a

	x0 := []z3bridge.Expr{a, b}
	xn := []z3bridge.Expr{an, bn}

	pdr := NewPDR(ctx, init, trans, bad, x0, nil, xn)
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (a stays false)")
	}
}

// ---------- Test 7: Mutex property ----------
// Two processes with tokens t0, t1. Initially both false (idle).
// trans: at most one can become true. bad: both true.

func TestPDR_MutualExclusion(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	t0 := ctx.Const("t0", bs)
	t1 := ctx.Const("t1", bs)
	t0n := ctx.Const("t0n", bs)
	t1n := ctx.Const("t1n", bs)

	// init: both idle
	init := ctx.And(ctx.Not(t0), ctx.Not(t1))

	// trans: at most one token can be held
	// Nondeterministic: either process can acquire if other doesn't hold
	// But constraint: not both true simultaneously in next state
	mutex := ctx.Not(ctx.And(t0n, t1n))
	trans := mutex // only constraint on next state

	// bad: both hold token
	bad := ctx.And(t0, t1)

	x0 := []z3bridge.Expr{t0, t1}
	xn := []z3bridge.Expr{t0n, t1n}

	pdr := NewPDR(ctx, init, trans, bad, x0, nil, xn)
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (mutex maintained)")
	}
}

// ---------- Test 8: Multi-step reachability (3+ frames needed) ----------

func TestPDR_MultiStep(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)
	c := ctx.Const("c", bs)
	an := ctx.Const("an", bs)
	bn := ctx.Const("bn", bs)
	cn := ctx.Const("cn", bs)

	// init: a=1, b=0, c=0
	init := ctx.And(a, ctx.Not(b), ctx.Not(c))

	// trans: a infects b, b infects c (chain propagation)
	// an = a, bn = a|b, cn = b|c
	trans := ctx.And(
		ctx.Iff(an, a),
		ctx.Iff(bn, ctx.Or(a, b)),
		ctx.Iff(cn, ctx.Or(b, c)),
	)

	// bad: c=true. Reachable after 2 steps: a->b->c
	bad := c

	x0 := []z3bridge.Expr{a, b, c}
	xn := []z3bridge.Expr{an, bn, cn}

	pdr := NewPDR(ctx, init, trans, bad, x0, nil, xn)
	result := pdr.Run()
	if result.Valid {
		t.Fatal("expected counterexample (infection reaches c in 2 steps)")
	}
}

// ---------- Test 9: Multi-step safe (infection blocked) ----------

func TestPDR_MultiStepSafe(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)
	c := ctx.Const("c", bs)
	an := ctx.Const("an", bs)
	bn := ctx.Const("bn", bs)
	cn := ctx.Const("cn", bs)

	// init: a=1, b=0, c=0
	init := ctx.And(a, ctx.Not(b), ctx.Not(c))

	// trans: a stays, b can become a, c stays false (blocked)
	trans := ctx.And(
		ctx.Iff(an, a),
		ctx.Iff(bn, ctx.Or(a, b)),
		ctx.Not(cn), // c always false
	)

	// bad: c=true
	bad := c

	x0 := []z3bridge.Expr{a, b, c}
	xn := []z3bridge.Expr{an, bn, cn}

	pdr := NewPDR(ctx, init, trans, bad, x0, nil, xn)
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (c always stays false)")
	}
}

// ---------- Test 10: Identity transition (safe) ----------

func TestPDR_Identity(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)
	trans := ctx.Iff(xn, x) // identity
	bad := x
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (identity preserves init)")
	}
}

// ---------- Test 11: Nondeterministic choice (unsafe) ----------

func TestPDR_NondeterministicUnsafe(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)
	trans := ctx.BoolVal(true) // xn can be anything
	bad := x
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	result := pdr.Run()
	if result.Valid {
		t.Fatal("expected counterexample (nondeterministic can reach bad)")
	}
}

// ---------- Test 12: With input variables ----------

func TestPDR_WithInputs(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	xn := ctx.Const("xn", bs)
	inp := ctx.Const("inp", bs)

	init := ctx.Not(x)
	// trans: x' = x AND inp (x can only become true if input is true AND x is true)
	trans := ctx.Iff(xn, ctx.And(x, inp))
	bad := x

	pdr := NewPDR(ctx, init, trans, bad,
		[]z3bridge.Expr{x}, []z3bridge.Expr{inp}, []z3bridge.Expr{xn})
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (x starts false and AND with input keeps it false)")
	}
}

// ---------- Test 13: Input can cause bad (unsafe) ----------

func TestPDR_InputCausesUnsafe(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	xn := ctx.Const("xn", bs)
	inp := ctx.Const("inp", bs)

	init := ctx.Not(x)
	// trans: x' = inp (input directly sets next state)
	trans := ctx.Iff(xn, inp)
	bad := x

	pdr := NewPDR(ctx, init, trans, bad,
		[]z3bridge.Expr{x}, []z3bridge.Expr{inp}, []z3bridge.Expr{xn})
	result := pdr.Run()
	if result.Valid {
		t.Fatal("expected counterexample (input can set x to true)")
	}
}

// ---------- Test 14: Statistics tracking ----------

func TestPDR_Statistics(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)
	trans := ctx.Iff(xn, x)
	bad := ctx.BoolVal(false)
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	pdr.Run()

	if pdr.SATQueryCount == 0 {
		t.Fatal("expected at least one SAT query")
	}
	if pdr.IterationCount == 0 {
		t.Fatal("expected at least one iteration")
	}
}

// ---------- Test 15: PDR String representation ----------

func TestPDR_String(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	init := ctx.Not(x)
	trans := ctx.Iff(xn, x)
	bad := ctx.BoolVal(false)
	pdr := NewPDR(ctx, init, trans, bad, []z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})
	s := pdr.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
	if !contains(s, "PDR") {
		t.Fatalf("expected 'PDR' in string: %s", s)
	}
}

// ---------- Test 16: cube2clause helper ----------

func TestCube2Clause(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)

	pdr := &PDR{ctx: ctx}

	// Empty cube
	c := pdr.cube2clause(nil)
	if !c.IsFalse() {
		t.Fatal("cube2clause of empty should be false")
	}

	// Single literal
	c = pdr.cube2clause([]z3bridge.Expr{a})
	// Should be (not a)
	s := c.String()
	if !contains(s, "not") {
		t.Fatalf("expected negation, got: %s", s)
	}

	// Two literals
	c = pdr.cube2clause([]z3bridge.Expr{a, b})
	s = c.String()
	if !contains(s, "or") {
		t.Fatalf("expected disjunction, got: %s", s)
	}
}

// ---------- Test 17: checkDisjoint helper ----------

func TestCheckDisjoint(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)

	pdr := &PDR{ctx: ctx}

	// a and ¬a should be disjoint
	if !pdr.checkDisjoint(a, ctx.Not(a)) {
		t.Fatal("a and ¬a should be disjoint")
	}

	// a and a should NOT be disjoint
	if pdr.checkDisjoint(a, a) {
		t.Fatal("a and a should not be disjoint")
	}

	// true and false should be disjoint
	if !pdr.checkDisjoint(ctx.BoolVal(true), ctx.BoolVal(false)) {
		t.Fatal("true and false should be disjoint")
	}
}

// ---------- Test 18: GoalHeap ordering ----------

func TestGoalHeap(t *testing.T) {
	gh := &GoalHeap{}
	heap.Init(gh)

	heap.Push(gh, &Goal{level: 5})
	heap.Push(gh, &Goal{level: 1})
	heap.Push(gh, &Goal{level: 3})

	g := heap.Pop(gh).(*Goal)
	if g.level != 1 {
		t.Fatalf("expected level 1, got %d", g.level)
	}
	g = heap.Pop(gh).(*Goal)
	if g.level != 3 {
		t.Fatalf("expected level 3, got %d", g.level)
	}
	g = heap.Pop(gh).(*Goal)
	if g.level != 5 {
		t.Fatalf("expected level 5, got %d", g.level)
	}
}

// ---------- Test 19: next/prev renaming ----------

func TestNextPrev(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	pdr := &PDR{ctx: ctx, x0: []z3bridge.Expr{x}, xn: []z3bridge.Expr{xn}}

	// next(x) should produce xn
	nx := pdr.next(x)
	// Check that next(x) equals xn by checking equivalence
	s := ctx.NewSolver()
	s.Assert(ctx.Not(ctx.Iff(nx, xn)))
	if s.Check() != z3bridge.Unsat {
		t.Fatal("next(x) should equal xn")
	}

	// prev(xn) should produce x
	px := pdr.prev(xn)
	s2 := ctx.NewSolver()
	s2.Assert(ctx.Not(ctx.Iff(px, x)))
	if s2.Check() != z3bridge.Unsat {
		t.Fatal("prev(xn) should equal x")
	}
}

// ---------- Test 20: ForwardClauses ----------

func TestForwardClauses(t *testing.T) {
	ctx, x, xn := makeOneBit(t)
	clause := ctx.Not(x) // ¬x

	forwarded := ForwardClauses(ctx, clause, []z3bridge.Expr{x}, []z3bridge.Expr{xn})

	// forwarded should be equivalent to ¬xn
	s := ctx.NewSolver()
	s.Assert(ctx.Not(ctx.Iff(forwarded, ctx.Not(xn))))
	if s.Check() != z3bridge.Unsat {
		t.Fatal("ForwardClauses should rename x to xn")
	}
}

// ---------- Test 21: NumClauses ----------

func TestNumClauses(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)

	n := NumClauses(ctx.BoolVal(true))
	if n != 0 {
		t.Fatalf("NumClauses(true) = %d, want 0", n)
	}

	n = NumClauses(a)
	if n != 1 {
		t.Fatalf("NumClauses(a) = %d, want 1", n)
	}

	conj := ctx.And(a, b)
	n = NumClauses(conj)
	if n < 2 {
		t.Fatalf("NumClauses(a ∧ b) = %d, want >= 2", n)
	}
}

// ---------- Test 22: CheckModule placeholder ----------

func TestCheckModule(t *testing.T) {
	mod := module.New()
	result, err := CheckModule(mod)
	if err != nil {
		t.Fatalf("CheckModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected trivially valid for empty module")
	}
}

// ---------- Test 23: CheckModule nil module ----------

func TestCheckModuleNil(t *testing.T) {
	_, err := CheckModule(nil)
	if err == nil {
		t.Fatal("expected error for nil module")
	}
}

// ---------- Test 24: Two variables, only one matters ----------

func TestPDR_IrrelevantVariable(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	y := ctx.Const("y", bs)
	xn := ctx.Const("xn", bs)
	yn := ctx.Const("yn", bs)

	// init: x=false (y irrelevant)
	init := ctx.Not(x)
	// trans: x'=false, y can be anything
	trans := ctx.Not(xn) // only constrains x
	// bad: x=true
	bad := x

	x0 := []z3bridge.Expr{x, y}
	xnv := []z3bridge.Expr{xn, yn}
	_ = y // suppress unused

	pdr := NewPDR(ctx, init, trans, bad, x0, nil, xnv)
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid (x stays false regardless of y)")
	}
}

// ---------- Test 25: Frame equality ----------

func TestFramesEqual(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	clause := ctx.Not(a)
	key := clause.String()

	pdr := &PDR{ctx: ctx}

	f1 := &Frame{clauses: map[string]z3bridge.Expr{key: clause}, solver: ctx.NewSolver()}
	f2 := &Frame{clauses: map[string]z3bridge.Expr{key: clause}, solver: ctx.NewSolver()}
	f3 := &Frame{clauses: map[string]z3bridge.Expr{}, solver: ctx.NewSolver()}

	pdr.frames = []*Frame{f1, f2, f3}

	if !pdr.framesEqual(0, 1) {
		t.Fatal("frames with same clauses should be equal")
	}
	if pdr.framesEqual(0, 2) {
		t.Fatal("frames with different clauses should not be equal")
	}
}

// ---------- Test 26: frameToExpr ----------

func TestFrameToExpr(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)

	pdr := &PDR{ctx: ctx}
	f := &Frame{clauses: map[string]z3bridge.Expr{}, solver: ctx.NewSolver()}
	pdr.frames = []*Frame{f}

	// Empty frame should be true
	e := pdr.frameToExpr(0)
	if !e.IsTrue() {
		t.Fatal("empty frame should be true")
	}

	// Single clause
	clause := ctx.Not(a)
	f.clauses[clause.String()] = clause
	e = pdr.frameToExpr(0)
	s := e.String()
	if !contains(s, "not") {
		t.Fatalf("expected negation in frame expr, got: %s", s)
	}
}

// ---------- Test 27: extractTrace ----------

func TestExtractTrace(t *testing.T) {
	pdr := &PDR{}
	g0 := &Goal{level: 0}
	g1 := &Goal{level: 1, parent: g0}
	g2 := &Goal{level: 2, parent: g1}

	trace := pdr.extractTrace(g2)
	if len(trace) != 3 {
		t.Fatalf("expected trace length 3, got %d", len(trace))
	}
	if trace[0].level != 0 {
		t.Fatal("first trace element should be level 0")
	}
	if trace[2].level != 2 {
		t.Fatal("last trace element should be level 2")
	}
}

// ---------- Test 28: prune removes subsumed clauses ----------

func TestPrune(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)

	pdr := &PDR{ctx: ctx}

	// (¬a) subsumes (¬a ∨ ¬b) because ¬a => (¬a ∨ ¬b)
	c1 := ctx.Not(a)
	c2 := ctx.Or(ctx.Not(a), ctx.Not(b))

	f := &Frame{
		clauses: map[string]z3bridge.Expr{
			c1.String(): c1,
			c2.String(): c2,
		},
		solver: ctx.NewSolver(),
	}

	pdr.prune(f)
	if len(f.clauses) != 1 {
		t.Fatalf("expected 1 clause after pruning, got %d", len(f.clauses))
	}
	// The remaining clause should be ¬a (the stronger one)
	if _, ok := f.clauses[c1.String()]; !ok {
		t.Fatal("expected ¬a to survive pruning")
	}
}

// ---------- Test 29: minimizeCube ----------

func TestMinimizeCube(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)

	pdr := &PDR{ctx: ctx}

	// Trivial case: cube with no inputs/lits
	result := pdr.minimizeCube([]z3bridge.Expr{a, b}, nil, nil)
	if len(result) == 0 {
		t.Fatal("minimizeCube should return non-empty result")
	}
}

// ---------- Test 30: NumUnivClauses ----------

func TestNumUnivClauses(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)

	n := NumUnivClauses(a)
	if n < 1 {
		t.Fatalf("NumUnivClauses(a) = %d, want >= 1", n)
	}
}

// ---------- Test 31: Large nondeterministic system ----------

func TestPDR_ThreeBitNondeterministic(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()

	// 3 bits, all start false, transitions keep them false
	var x0, xn []z3bridge.Expr
	var transParts []z3bridge.Expr
	for i := 0; i < 3; i++ {
		xi := ctx.Const(fmt.Sprintf("x%d", i), bs)
		xni := ctx.Const(fmt.Sprintf("xn%d", i), bs)
		x0 = append(x0, xi)
		xn = append(xn, xni)
		transParts = append(transParts, ctx.Not(xni)) // all stay false
	}

	init := ctx.And(ctx.Not(x0[0]), ctx.Not(x0[1]), ctx.Not(x0[2]))
	trans := ctx.And(transParts...)
	bad := x0[0] // bad: first bit true

	pdr := NewPDR(ctx, init, trans, bad, x0, nil, xn)
	result := pdr.Run()
	if !result.Valid {
		t.Fatal("expected valid")
	}
}

// ---------- Test 32: Eventually reaches bad (delayed) ----------

func TestPDR_DelayedBad(t *testing.T) {
	ctx := z3bridge.NewContext()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)
	an := ctx.Const("an", bs)
	bn := ctx.Const("bn", bs)

	// init: a=0, b=0
	init := ctx.And(ctx.Not(a), ctx.Not(b))
	// trans: a'=¬a, b'=a (b gets old value of a)
	trans := ctx.And(
		ctx.Iff(an, ctx.Not(a)),
		ctx.Iff(bn, a),
	)
	// bad: b=1 (happens at step 2: a flips to 1 at step 1, b gets 1 at step 2)
	bad := b

	pdr := NewPDR(ctx, init, trans, bad,
		[]z3bridge.Expr{a, b}, nil, []z3bridge.Expr{an, bn})
	result := pdr.Run()
	if result.Valid {
		t.Fatal("expected counterexample (b becomes true after 2 steps)")
	}
}

// ---------- Test 33: Fuzz test ----------

func FuzzPDR_RandomOneBit(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint8(0))
	f.Add(uint8(1), uint8(1), uint8(1))
	f.Add(uint8(0), uint8(1), uint8(0))

	f.Fuzz(func(t *testing.T, initVal, transFlip, badVal uint8) {
		ctx := z3bridge.NewContext()
		bs := ctx.BoolSort()
		x := ctx.Const("x", bs)
		xn := ctx.Const("xn", bs)

		var init z3bridge.Expr
		if initVal%2 == 0 {
			init = ctx.Not(x)
		} else {
			init = x
		}

		var trans z3bridge.Expr
		if transFlip%2 == 0 {
			trans = ctx.Iff(xn, x) // identity
		} else {
			trans = ctx.Iff(xn, ctx.Not(x)) // flip
		}

		var bad z3bridge.Expr
		if badVal%2 == 0 {
			bad = x
		} else {
			bad = ctx.Not(x)
		}

		pdr := NewPDR(ctx, init, trans, bad,
			[]z3bridge.Expr{x}, nil, []z3bridge.Expr{xn})

		// Just verify it terminates without panic
		result := pdr.Run()
		_ = result
	})
}

// ---------- Test 34: Random multi-bit systems ----------

func TestPDR_RandomSmall(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	for trial := 0; trial < 10; trial++ {
		ctx := z3bridge.NewContext()
		bs := ctx.BoolSort()
		n := 2 + rng.Intn(2) // 2 or 3 bits

		x0 := make([]z3bridge.Expr, n)
		xn := make([]z3bridge.Expr, n)
		for i := 0; i < n; i++ {
			x0[i] = ctx.Const(fmt.Sprintf("r%d_%d", trial, i), bs)
			xn[i] = ctx.Const(fmt.Sprintf("rn%d_%d", trial, i), bs)
		}

		// init: all false
		initParts := make([]z3bridge.Expr, n)
		for i := 0; i < n; i++ {
			initParts[i] = ctx.Not(x0[i])
		}
		init := ctx.And(initParts...)

		// trans: each bit stays false (safe by construction)
		transParts := make([]z3bridge.Expr, n)
		for i := 0; i < n; i++ {
			transParts[i] = ctx.Not(xn[i])
		}
		trans := ctx.And(transParts...)

		// bad: first bit true
		bad := x0[0]

		pdr := NewPDR(ctx, init, trans, bad, x0, nil, xn)
		result := pdr.Run()
		if !result.Valid {
			t.Fatalf("trial %d: expected valid (all bits stay false)", trial)
		}
	}
}

// ---------- helpers ----------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsImpl(s, substr)
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
