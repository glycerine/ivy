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

// SetnvarsC is sat_solver_setnvars.
func (s *sat_solver_t) SetnvarsC(n int32) { sat_solver_setnvars(s, n) }

// NVarsC is sat_solver_nvars.
func (s *sat_solver_t) NVarsC() int32 { return sat_solver_nvars(s) }

// SolveC is sat_solver_solve with C's limits.
func (s *sat_solver_t) SolveC(assumptions []Literal, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal int64) int32 {
	status := sat_solver_solve(s, assumptions, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal)
	s.lastStatus = Status(status)
	return status
}

// CompressC is sat_solver_compress.
func (s *sat_solver_t) CompressC() { sat_solver_compress(s) }

// VarValueC is sat_solver_var_value.
func (s *sat_solver_t) VarValueC(v int32) int32 { return sat_solver_var_value(s, v) }

// MemoryC is sat_solver_memory.
func (s *sat_solver_t) MemoryC() float64 { return sat_solver_memory(s) }

// CountUsedVarsC is sat_solver_count_usedvars.
func (s *sat_solver_t) CountUsedVarsC() int32 { return sat_solver_count_usedvars(s) }

// StatsC reads s->stats.clauses, s->stats.conflicts and s->stats.learnts.
func (s *sat_solver_t) StatsC() (clauses, conflicts, learnts float64) {
	return float64(s.stats.clauses), float64(s.stats.conflicts), float64(s.stats.learnts)
}

// LearntLimitsC reads nLearntStart, nLearntDelta, nLearntRatio and nDBreduces.
func (s *sat_solver_t) LearntLimitsC() (start, delta, ratio, reduces int32) {
	return s.nLearntStart, s.nLearntDelta, s.nLearntRatio, s.nDBreduces
}

// SetSolvedC is pSat->fSolved = 1.
func (s *sat_solver_t) SetSolvedC() { s.fSolved = 1 }

// SetGlobalVarsC is pSat->pGlobalVars = pGlobalVars (nil clears it).
func (s *sat_solver_t) SetGlobalVarsC(pGlobalVars []int32) { s.pGlobalVars = pGlobalVars }

// SetRuntimeLimitC is sat_solver_set_runtime_limit with an Abc_Clock deadline.
func (s *sat_solver_t) SetRuntimeLimitC(clock func() int64, nRuntimeLimit int64) int64 {
	old := s.nRuntimeLimit
	s.SetRuntimeDeadline(clock, nRuntimeLimit)
	return old
}

// StoreAllocC is sat_solver_store_alloc.
func (s *sat_solver_t) StoreAllocC() { sat_solver_store_alloc(s) }

// StoreMarkClausesAC is sat_solver_store_mark_clauses_a.
func (s *sat_solver_t) StoreMarkClausesAC() { sat_solver_store_mark_clauses_a(s) }

// StoreMarkRootsC is sat_solver_store_mark_roots.
func (s *sat_solver_t) StoreMarkRootsC() { sat_solver_store_mark_roots(s) }

// StoreReleaseC is sat_solver_store_release.
func (s *sat_solver_t) StoreReleaseC() *Sto_Man_t { return s.ReleaseClauseStore() }

// SetVerbosityC is pSat->verbosity = verbosity.
func (s *sat_solver_t) SetVerbosityC(verbosity int32) { s.verbosity = verbosity }
