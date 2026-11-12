// Go port of xsatSolverAPI.c by Bruno Schmitt (2016).
package xsat

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func New() *Solver {
	s := &Solver{Config: DefaultConfig, memory: xSAT_MemAlloc(0), watches: xSAT_VecWatchListAlloc(0), binary: xSAT_VecWatchListAlloc(0), varInc: 1 << 5, claInc: 1 << 11, okay: true}
	s.order = xSAT_HeapAlloc(&s.activity)
	s.lbdQueue = xSAT_BQueueNew(s.Config.LBDWindow)
	s.trailQueue = xSAT_BQueueNew(s.Config.TrailWindow)
	s.reduceAt = int64(s.Config.FirstReduce)
	s.rc2 = int64(s.Config.FirstReduce)
	return s
}
func xSAT_SolverCreate() *xSAT_Solver_t { return New() }
func xSAT_SolverDestroy(s *Solver) {
	if s != nil {
		*s = Solver{}
	}
}
func (s *Solver) AddVariable(sign int) int {
	v := len(s.assigns)
	s.assigns = append(s.assigns, VarX)
	s.polarity = append(s.polarity, byte(boolInt(sign != 0)))
	s.activity = append(s.activity, 0)
	s.levels = append(s.levels, 0)
	s.reasons = append(s.reasons, CRefUndef)
	xSAT_VecWatchListPush(s.watches)
	xSAT_VecWatchListPush(s.watches)
	xSAT_VecWatchListPush(s.binary)
	xSAT_VecWatchListPush(s.binary)
	xSAT_HeapInsert(s.order, v)
	s.model = nil
	return v
}
func xSAT_SolverAddVariable(s *Solver, sign int) { s.AddVariable(sign) }

// AddClause preserves the caller's Go slice. The compatibility entry point
// below additionally retains the original C vector's sorting/shrinking behavior.
func (s *Solver) AddClause(literals []int) bool {
	result, _ := s.addMutableClause(append([]int(nil), literals...))
	return result
}
func (s *Solver) addMutableClause(lits []int) (bool, int) {
	for _, lit := range lits {
		if lit < 0 {
			panic("negative encoded literal")
		}
	}
	s.model = nil
	if !s.okay {
		return false, -1
	}
	xSAT_SolverCancelUntil(s, 0)
	for _, lit := range lits {
		for lit>>1 >= len(s.assigns) {
			s.AddVariable(1)
		}
	}
	sort.Ints(lits)
	previous := LitUndef
	filtered := lits[:0]
	for _, lit := range lits {
		if lit == previous {
			continue
		}
		if lit == previous^1 {
			return true, -1
		}
		previous = lit
		value := s.litValue(lit)
		if value == LBoolTrue {
			return true, -1
		}
		if value == LBoolUndef {
			filtered = append(filtered, lit)
		}
	}
	if len(filtered) == 0 {
		s.okay = false
		return false, 0
	}
	if len(filtered) == 1 {
		if !s.enqueue(filtered[0], CRefUndef) || s.propagate() != CRefUndef {
			s.okay = false
			return false, 1
		}
		return true, 1
	}
	s.newClause(filtered, false)
	return true, len(filtered)
}
func xSAT_SolverAddClause(s *Solver, lits *vec.Vec_Int_t) int {
	values := vec.Vec_IntArray(lits)
	native := make([]int, len(values))
	for i, value := range values {
		native[i] = int(value)
	}
	result, size := s.addMutableClause(native)
	for i, value := range native {
		vec.Vec_IntWriteEntry(lits, i, int32(value))
	}
	if size >= 0 {
		vec.Vec_IntShrink(lits, size)
	}
	return boolInt(result)
}
func (s *Solver) Simplify() bool {
	if len(s.limits) != 0 {
		panic("simplification requires root level")
	}
	if !s.okay {
		return false
	}
	if s.propagate() != CRefUndef {
		s.okay = false
		return false
	}
	if s.simplifyAssigned == len(s.trail) || s.simplifyRemaining > 0 {
		return true
	}
	kept := s.clauses[:0]
	for _, ref := range s.clauses {
		c := s.memory.clauses[ref]
		if xSAT_SolverIsClauseSatisfied(s, c) != 0 {
			c.marked = true
			s.Stats.ClauseLiterals -= int64(len(c.lits))
			s.unwatch(ref)
		} else {
			kept = append(kept, ref)
		}
	}
	s.clauses = kept
	xSAT_SolverRebuildOrderHeap(s)
	s.simplifyAssigned = len(s.trail)
	s.simplifyRemaining = s.Stats.ClauseLiterals + s.Stats.LearntLiterals
	return true
}
func xSAT_SolverSimplify(s *Solver) int { return boolInt(s.Simplify()) }
func xSAT_SolverPrintStats(s *Solver) {
	fmt.Printf("starts        : %10d\nconflicts     : %10d\ndecisions     : %10d\npropagations  : %10d\n", s.Stats.Starts, s.Stats.Conflicts, s.Stats.Decisions, s.Stats.Propagations)
}
func xSAT_SolverSolve(s *Solver) int { return s.Solve() }
