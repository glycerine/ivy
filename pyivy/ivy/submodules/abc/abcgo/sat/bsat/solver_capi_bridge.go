package bsat

// C-API bridges for line-by-line ports that call ABC's sat_solver_* routines
// directly. Unlike AddClause and Simplify, these never mark the solver as
// permanently UNSAT: C's sat_solver has no such flag (fSolved is never set),
// so a failed sat_solver_addclause leaves later sat_solver_solve calls free
// to search, and the ports must observe the same behaviour.

// AddClauseC is sat_solver_addclause.
func (s *sat_solver_t) AddClauseC(literals []Literal) int32 {
	s.lastStatus = Unknown
	return sat_solver_addclause(s, literals)
}

// AddAndC is sat_solver_add_and.
func (s *sat_solver_t) AddAndC(iVar, iVar0, iVar1, fCompl0, fCompl1, fCompl int32) int32 {
	s.lastStatus = Unknown
	return sat_solver_add_and(s, iVar, iVar0, iVar1, fCompl0, fCompl1, fCompl)
}

// AddBufferC is sat_solver_add_buffer.
func (s *sat_solver_t) AddBufferC(iVarA, iVarB, fCompl int32) int32 {
	s.lastStatus = Unknown
	return sat_solver_add_buffer(s, iVarA, iVarB, fCompl)
}

// SimplifyC is sat_solver_simplify.
func (s *sat_solver_t) SimplifyC() int32 { return sat_solver_simplify(s) }

// NClauses is sat_solver_nclauses.
func (s *sat_solver_t) NClauses() int32 { return sat_solver_nclauses(s) }

// NConflicts is sat_solver_nconflicts.
func (s *sat_solver_t) NConflicts() int32 { return sat_solver_nconflicts(s) }

// PushC is sat_solver_push: the assumption stays on the trail until popped.
func (s *sat_solver_t) PushC(literal Literal) int32 { return sat_solver_push(s, literal) }

// SolveInternalC is sat_solver_solve_internal (solves under the pushed
// assumptions without pushing or popping any).
func (s *sat_solver_t) SolveInternalC() int32 {
	status := sat_solver_solve_internal(s)
	s.lastStatus = Status(status)
	return status
}

// SetRandomC is sat_solver_set_random.
func (s *sat_solver_t) SetRandomC(fNotUseRandom int32) int32 {
	return sat_solver_set_random(s, fNotUseRandom)
}

// SetPolarityC is sat_solver_set_polarity.
func (s *sat_solver_t) SetPolarityC(vars []int32) { sat_solver_set_polarity(s, vars, int32(len(vars))) }

// BookmarkC is sat_solver_bookmark.
func (s *sat_solver_t) BookmarkC() { sat_solver_bookmark(s) }

// SetResourceLimitsC is sat_solver_set_resource_limits.
func (s *sat_solver_t) SetResourceLimitsC(conflicts, propagations, globalConflicts, globalPropagations int64) {
	sat_solver_set_resource_limits(s, conflicts, propagations, globalConflicts, globalPropagations)
}

// RollbackC is sat_solver_rollback.
func (s *sat_solver_t) RollbackC() {
	s.lastStatus = Unknown
	sat_solver_rollback(s)
}
