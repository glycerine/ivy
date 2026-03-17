// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from mini_ic3.py.

// Package updr implements the IC3/PDR (Property-Directed Reachability)
// model checking algorithm and the UPDR orchestration layer for Ivy.
//
// IC3/PDR works by maintaining a sequence of over-approximating frames
// F0, F1, ..., Fn where each Fi represents a set of reachable states
// at step i. The algorithm refines these frames by blocking bad cubes
// (conjunctions of literals that can reach the error state) and
// propagating clauses forward until either a fixpoint (inductive
// invariant) is found or a concrete counterexample is discovered.
package updr

import (
	"container/heap"
	"fmt"
	"runtime"
	"strings"

	"github.com/glycerine/goivy/z3bridge"
)

// PDRResult holds the outcome of a PDR run.
type PDRResult struct {
	Valid     bool          // true if the property holds
	Invariant z3bridge.Expr // if Valid, the inductive invariant
	Trace     []*Goal       // if !Valid, counterexample trace (root first)
}

// Frame represents one level in the PDR frame sequence.
// Each frame is a conjunction of clauses that over-approximates the
// states reachable in at most that many steps.
type Frame struct {
	clauses map[string]z3bridge.Expr // keyed by String() for dedup
	solver  *z3bridge.Solver
}

// Goal represents a proof obligation: a cube (conjunction of literals)
// that must be blocked at a given level.
type Goal struct {
	cube   []z3bridge.Expr
	parent *Goal
	level  int
}

// GoalHeap is a min-heap of goals ordered by level (lowest first).
type GoalHeap []*Goal

func (h GoalHeap) Len() int            { return len(h) }
func (h GoalHeap) Less(i, j int) bool  { return h[i].level < h[j].level }
func (h GoalHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *GoalHeap) Push(x interface{}) { *h = append(*h, x.(*Goal)) }
func (h *GoalHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}

// PDR implements the IC3/Property-Directed Reachability algorithm.
//
// The algorithm maintains frames F0 ... Fn where:
//   - F0 = init
//   - Fi+1 is implied by the image of Fi under the transition relation
//   - Each Fi over-approximates the states reachable in <= i steps
//
// It terminates when either:
//   - Some Fi = Fi+1 (fixpoint => invariant found => property holds)
//   - A concrete path from init to bad is found (counterexample)
type PDR struct {
	ctx    *z3bridge.Context
	x0     []z3bridge.Expr // current-state variables
	xn     []z3bridge.Expr // next-state variables
	inputs []z3bridge.Expr // input variables

	init  z3bridge.Expr // initial state formula (over x0)
	trans z3bridge.Expr // transition relation (over x0, xn, inputs)
	bad   z3bridge.Expr // bad state formula (over x0)

	frames []*Frame  // F0, F1, ..., Fn
	goals  *GoalHeap // priority queue of proof obligations

	// Statistics
	N              int // number of frames (len(frames)-1)
	IterationCount int
	SATQueryCount  int
}

// NewPDR creates a new PDR instance.
//
// Parameters:
//   - ctx: Z3 context
//   - init: initial state formula (over x0 variables)
//   - trans: transition relation (over x0, xn, inputs)
//   - bad: bad/error state formula (over x0 variables)
//   - x0: current-state variables
//   - inputs: input variables (may be nil)
//   - xn: next-state variables (must correspond 1-to-1 with x0)
func NewPDR(ctx *z3bridge.Context, init, trans, bad z3bridge.Expr,
	x0, inputs, xn []z3bridge.Expr) *PDR {

	p := &PDR{
		ctx:    ctx,
		x0:     x0,
		xn:     xn,
		inputs: inputs,
		init:   init,
		trans:  trans,
		bad:    bad,
	}

	gh := &GoalHeap{}
	heap.Init(gh)
	p.goals = gh

	// Create F0: solver with init asserted
	f0 := &Frame{
		clauses: make(map[string]z3bridge.Expr),
		solver:  ctx.NewSolver(),
	}
	f0.solver.Assert(init)
	p.frames = []*Frame{f0}
	p.N = 0

	return p
}

// Run executes the PDR algorithm and returns the result.
func (p *PDR) Run() PDRResult {
	// Safety check: init and bad must be disjoint
	if !p.checkDisjoint(p.init, p.bad) {
		// Init intersects bad — immediate counterexample
		return PDRResult{
			Valid: false,
			Trace: []*Goal{{cube: nil, level: 0}},
		}
	}

	for {
		p.IterationCount++

		// Phase 1: Unfold — check if bad is reachable from frame N
		res, cube := p.unfold()
		if res == z3bridge.Sat {
			// Bad state reachable from current frontier — try to block
			g := &Goal{cube: cube, level: p.N}
			heap.Push(p.goals, g)

			// Phase 2: Block all goals
			ok := p.blockGoals()
			if !ok {
				// Found a real counterexample
				// Find the goal at level 0 to extract trace
				trace := p.extractTrace(g)
				return PDRResult{Valid: false, Trace: trace}
			}
		} else {
			// No bad state reachable — add a new frame
			p.addFrame()

			// Phase 3: Propagate clauses forward
			p.propagate()

			// Phase 4: Check for fixpoint
			if inv := p.isValid(); inv != nil {
				return PDRResult{Valid: true, Invariant: *inv}
			}
		}
	}
}

// addFrame adds a new empty frame.
func (p *PDR) addFrame() {
	f := &Frame{
		clauses: make(map[string]z3bridge.Expr),
		solver:  p.ctx.NewSolver(),
	}
	p.frames = append(p.frames, f)
	p.N = len(p.frames) - 1
}

// unfold checks whether the bad state is reachable from frame N.
// Returns (Sat, cube) if a predecessor of bad exists in frame N,
// or (Unsat, nil) otherwise.
func (p *PDR) unfold() (z3bridge.CheckResult, []z3bridge.Expr) {
	s := p.frames[p.N].solver
	s.Push()
	defer s.Pop()

	// Assert trans and bad(xn) — look for a state in F_N that can
	// transition to a bad state
	s.Assert(p.trans)
	s.Assert(p.next(p.bad))

	p.SATQueryCount++
	res := s.Check()
	if res == z3bridge.Sat {
		cube := p.extractCube(s)
		return z3bridge.Sat, cube
	}
	return z3bridge.Unsat, nil
}

// blockGoals processes all queued proof obligations.
// Returns false if a counterexample at level 0 is found (cannot block).
func (p *PDR) blockGoals() bool {
	for p.goals.Len() > 0 {
		g := heap.Pop(p.goals).(*Goal)

		if g.level == 0 {
			// Cannot block at level 0 — counterexample found
			return false
		}

		// Check if cube is already blocked
		if !p.isSAT(p.frames[g.level].solver, g.cube) {
			continue
		}

		// Try to find a predecessor in frame g.level-1
		core, _, res := p.isInductive(g.level-1, g.cube)
		if res == z3bridge.Unsat {
			// Cube is inductive relative to frame g.level-1
			// Block it at levels up to the generalized level
			genCube, genLevel := p.generalize(core, g.level)
			p.blockCube(genLevel, genCube)
		} else {
			// Found a predecessor — push it as a new goal
			predGoal := &Goal{
				cube:   core,
				parent: g,
				level:  g.level - 1,
			}
			heap.Push(p.goals, predGoal)
			// Re-queue original goal
			heap.Push(p.goals, g)
		}
	}
	return true
}

// isInductive checks whether a cube has a predecessor in the given frame.
//
// It checks: F_level ∧ T ∧ cube' is SAT?
// If SAT, returns (predecessorCube, 0, Sat) — the cube is reachable.
// If UNSAT, returns (cube, maxLevel, Unsat) — the cube is blocked.
func (p *PDR) isInductive(level int, cube []z3bridge.Expr) ([]z3bridge.Expr, int, z3bridge.CheckResult) {
	s := p.frames[level].solver
	s.Push()
	defer s.Pop()

	// Assert transition relation
	s.Assert(p.trans)

	// Assert cube in next state: we want to find a state in F_level
	// that can transition into the cube
	for _, lit := range cube {
		s.Assert(p.nextExpr(lit))
	}

	p.SATQueryCount++
	res := s.Check()

	if res == z3bridge.Unsat {
		// No predecessor in this frame — cube is blocked
		maxLevel := p.N
		return cube, maxLevel, z3bridge.Unsat
	}

	// SAT — extract predecessor cube from the model
	pred := p.extractCube(s)
	return pred, 0, z3bridge.Sat
}

// blockCube adds ¬cube as a clause to frames 1..level.
func (p *PDR) blockCube(level int, cube []z3bridge.Expr) {
	clause := p.cube2clause(cube)
	key := clause.String()

	for i := 1; i <= level && i < len(p.frames); i++ {
		if _, exists := p.frames[i].clauses[key]; !exists {
			p.frames[i].clauses[key] = clause
			p.frames[i].solver.Assert(clause)
		}
	}
}

// generalize tries to minimize a cube while keeping it inductive.
// Returns (minimizedCube, maxLevel).
func (p *PDR) generalize(cube []z3bridge.Expr, level int) ([]z3bridge.Expr, int) {
	// Try dropping each literal
	result := make([]z3bridge.Expr, len(cube))
	copy(result, cube)

	for i := 0; i < len(result); {
		// Try without literal i
		candidate := make([]z3bridge.Expr, 0, len(result)-1)
		candidate = append(candidate, result[:i]...)
		candidate = append(candidate, result[i+1:]...)

		if len(candidate) == 0 {
			break
		}

		// Check if the reduced cube is still inductive
		_, _, res := p.isInductive(level-1, candidate)
		if res == z3bridge.Unsat {
			result = candidate
			// Don't increment i — the next element slid into position i
		} else {
			i++
		}
	}

	return result, level
}

// propagate pushes clauses from lower frames to higher frames.
func (p *PDR) propagate() {
	for i := 1; i < p.N; i++ {
		for key, clause := range p.frames[i].clauses {
			// Check if clause holds in frame i+1
			if i+1 < len(p.frames) {
				if _, exists := p.frames[i+1].clauses[key]; exists {
					continue // already there
				}
				// Check: F_i ∧ T ∧ ¬clause' is UNSAT?
				// If so, clause is inductive and can be pushed
				s := p.frames[i].solver
				s.Push()
				s.Assert(p.trans)
				s.Assert(p.next(p.ctx.Not(clause)))
				p.SATQueryCount++
				res := s.Check()
				s.Pop()

				if res == z3bridge.Unsat {
					p.frames[i+1].clauses[key] = clause
					p.frames[i+1].solver.Assert(clause)
				}
			}
		}
	}
}

// isValid checks for a fixpoint: whether any frame Fi equals Fi+1.
// If found, returns the conjunction of clauses as the inductive invariant.
func (p *PDR) isValid() *z3bridge.Expr {
	for i := 1; i < p.N; i++ {
		if p.framesEqual(i, i+1) {
			inv := p.frameToExpr(i)
			return &inv
		}
	}
	return nil
}

// framesEqual checks if frames[i] and frames[j] have the same clause set.
func (p *PDR) framesEqual(i, j int) bool {
	if i >= len(p.frames) || j >= len(p.frames) {
		return false
	}
	fi := p.frames[i]
	fj := p.frames[j]
	if len(fi.clauses) != len(fj.clauses) {
		return false
	}
	for key := range fi.clauses {
		if _, ok := fj.clauses[key]; !ok {
			return false
		}
	}
	return true
}

// frameToExpr returns the conjunction of all clauses in a frame.
func (p *PDR) frameToExpr(i int) z3bridge.Expr {
	f := p.frames[i]
	if len(f.clauses) == 0 {
		return p.ctx.BoolVal(true)
	}
	args := make([]z3bridge.Expr, 0, len(f.clauses))
	for _, c := range f.clauses {
		args = append(args, c)
	}
	if len(args) == 1 {
		return args[0]
	}
	return p.ctx.And(args...)
}

// extractCube extracts a cube (conjunction of literals) from a SAT model.
// The cube is over the current-state variables x0.
func (p *PDR) extractCube(s *z3bridge.Solver) []z3bridge.Expr {
	m := s.Model()
	if m == nil {
		return nil
	}
	var cube []z3bridge.Expr
	for _, v := range p.x0 {
		val, ok := m.Eval(v, true)
		if ok {
			if val.IsTrue() {
				cube = append(cube, v)
			} else if val.IsFalse() {
				cube = append(cube, p.ctx.Not(v))
			}
			// Skip variables with non-boolean values
		}
	}
	runtime.KeepAlive(m)
	return cube
}

// extractTrace walks parent pointers from a goal back to level 0.
func (p *PDR) extractTrace(g *Goal) []*Goal {
	// Walk up to find the root
	var trace []*Goal
	for cur := g; cur != nil; cur = cur.parent {
		trace = append(trace, cur)
	}
	// Reverse so root (level 0) is first
	for i, j := 0, len(trace)-1; i < j; i, j = i+1, j-1 {
		trace[i], trace[j] = trace[j], trace[i]
	}
	return trace
}

// isSAT checks if cube is satisfiable with the solver's current assertions.
func (p *PDR) isSAT(s *z3bridge.Solver, cube []z3bridge.Expr) bool {
	s.Push()
	defer s.Pop()
	for _, lit := range cube {
		s.Assert(lit)
	}
	p.SATQueryCount++
	return s.Check() == z3bridge.Sat
}

// --- Variable renaming helpers ---

// next renames an expression from current-state (x0) to next-state (xn).
func (p *PDR) next(e z3bridge.Expr) z3bridge.Expr {
	if len(p.x0) == 0 {
		return e
	}
	return p.ctx.Substitute(e, p.x0, p.xn)
}

// nextExpr renames a single expression to next-state variables.
func (p *PDR) nextExpr(e z3bridge.Expr) z3bridge.Expr {
	return p.next(e)
}

// prev renames an expression from next-state (xn) to current-state (x0).
func (p *PDR) prev(e z3bridge.Expr) z3bridge.Expr {
	if len(p.xn) == 0 {
		return e
	}
	return p.ctx.Substitute(e, p.xn, p.x0)
}

// cube2clause negates a cube to produce a clause.
// A cube is a conjunction of literals; its negation is a disjunction (clause).
// cube = [l1, l2, ..., ln] => clause = (¬l1 ∨ ¬l2 ∨ ... ∨ ¬ln)
func (p *PDR) cube2clause(cube []z3bridge.Expr) z3bridge.Expr {
	if len(cube) == 0 {
		return p.ctx.BoolVal(false) // empty cube = true, negation = false
	}
	negated := make([]z3bridge.Expr, len(cube))
	for i, lit := range cube {
		negated[i] = p.ctx.Not(lit)
	}
	if len(negated) == 1 {
		return negated[0]
	}
	return p.ctx.Or(negated...)
}

// checkDisjoint returns true if a ∧ b is UNSAT (the formulas are disjoint).
func (p *PDR) checkDisjoint(a, b z3bridge.Expr) bool {
	s := p.ctx.NewSolver()
	s.Assert(a)
	s.Assert(b)
	p.SATQueryCount++
	return s.Check() == z3bridge.Unsat
}

// minimizeCube removes literals from a cube that are not needed for
// the UNSAT core, producing a smaller blocking clause.
func (p *PDR) minimizeCube(cube, inputs, lits []z3bridge.Expr) []z3bridge.Expr {
	s := p.ctx.NewSolver()
	for _, e := range inputs {
		s.Assert(e)
	}
	for _, e := range lits {
		s.Assert(e)
	}
	res := s.CheckAssumptions(cube)
	if res == z3bridge.Unsat {
		core := s.UnsatCore()
		if len(core) > 0 {
			return core
		}
	}
	return cube
}

// prune removes subsumed clauses from a frame's clause set.
// A clause c1 subsumes c2 if c1 => c2 (i.e., c1 is stronger).
//
// had a bug: the root cause was not Z3 at all -- it was
// Go map iteration order nondeterminism. The prune function
// only checked one direction of subsumption (ci => cj), so
// when the map happened to yield the weaker clause first,
// the stronger clause's subsumption check was never reached.
//
// The fix adds a reverse check: if ci doesn't subsume cj, also
// check if cj subsumes ci.
func (p *PDR) prune(f *Frame) {
	// Simple approach: for each pair, check if one subsumes the other.
	// This is O(n^2) but frames are typically small.
	keys := make([]string, 0, len(f.clauses))
	for k := range f.clauses {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			ci := f.clauses[keys[i]]
			cj := f.clauses[keys[j]]
			// Check if ci => cj (ci subsumes cj, so remove cj)
			s := p.ctx.NewSolver()
			s.Assert(ci)
			s.Assert(p.ctx.Not(cj))
			if s.Check() == z3bridge.Unsat {
				delete(f.clauses, keys[j])
				keys = append(keys[:j], keys[j+1:]...)
				j--
				continue
			}
			// Check if cj => ci (cj subsumes ci, so remove ci)
			s2 := p.ctx.NewSolver()
			s2.Assert(cj)
			s2.Assert(p.ctx.Not(ci))
			if s2.Check() == z3bridge.Unsat {
				delete(f.clauses, keys[i])
				keys = append(keys[:i], keys[i+1:]...)
				i--
				break
			}
		}
	}
}

// String returns a human-readable summary of the PDR state.
func (p *PDR) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "PDR: %d frames, %d iterations, %d SAT queries\n",
		len(p.frames), p.IterationCount, p.SATQueryCount)
	for i, f := range p.frames {
		fmt.Fprintf(&b, "  F%d: %d clauses\n", i, len(f.clauses))
	}
	return b.String()
}
