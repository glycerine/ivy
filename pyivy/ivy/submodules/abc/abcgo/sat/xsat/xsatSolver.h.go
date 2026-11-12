// Go-owned solver state corresponding to xsatSolver.h.
package xsat

type Solver struct {
	Config               Options
	Stats                Stats
	memory               *xSAT_Mem_t
	clauses, learnts     []uint32
	watches, binary      *xSAT_VecWatchList_t
	activity             []uint32
	order                *xSAT_Heap_t
	levels               []int
	reasons              []uint32
	assigns, polarity    []byte
	trail, limits        []int
	qhead                int
	varInc, claInc       uint32
	lbdQueue, trailQueue *xSAT_BQueue_t
	sumLBD               float64
	reduceAt, rc2        int64
	simplifyAssigned     int
	simplifyRemaining    int64
	okay                 bool
	model                []int
}
type xSAT_Solver_t = Solver

func xSAT_Var2Lit(v, c int) int              { return 2*v + boolInt(c != 0) }
func xSAT_NegLit(lit int) int                { return lit ^ 1 }
func xSAT_Lit2Var(lit int) int               { return lit >> 1 }
func xSAT_LitSign(lit int) int               { return lit & 1 }
func xSAT_SolverDecisionLevel(s *Solver) int { return len(s.limits) }
func xSAT_SolverReadClause(s *Solver, ref uint32) *xSAT_Clause_t {
	return xSAT_MemClauseHand(s.memory, ref)
}
func (s *Solver) litValue(lit int) int {
	value := s.assigns[lit>>1]
	if value == VarX {
		return LBoolUndef
	}
	if int(value) == lit&1 {
		return LBoolTrue
	}
	return LBoolFalse
}
func xSAT_SolverIsClauseSatisfied(s *Solver, c *xSAT_Clause_t) int {
	for _, lit := range c.lits {
		if s.litValue(lit) == LBoolTrue {
			return 1
		}
	}
	return 0
}
func (s *Solver) ModelValue(variable int) int {
	if variable < 0 || variable >= len(s.model) {
		return -1
	}
	return s.model[variable]
}
func (s *Solver) VariableCount() int { return len(s.assigns) }
func (s *Solver) ClauseCount() int   { return len(s.clauses) }
func (s *Solver) LearntCount() int   { return len(s.learnts) }
