package bsat

// Exported one-to-one wrappers of satSolver2.h/satSolver2.c for ports that
// drive sat_solver2 directly, as absGla.c's Ga2_Man_t does: the solver, its
// proof logger (pPrf2) and the counters the C code prints.

// Sat_solver2 is C's sat_solver2.
type Sat_solver2 = sat_solver2

func Sat_solver2_new() *Sat_solver2                { return sat_solver2_new() }
func Sat_solver2_delete(s *Sat_solver2)            { sat_solver2_delete(s) }
func Sat_solver2_setnvars(s *Sat_solver2, n int32) { sat_solver2_setnvars(s, n) }

// Sat_solver2_addclause adds the clause lits with the given clause ID.
func Sat_solver2_addclause(s *Sat_solver2, lits []int32, id int32) int32 {
	return sat_solver2_addclause(s, lits, id)
}

// Sat_solver2_solve returns l_True (1), l_False (-1) or l_Undef (0).
func Sat_solver2_solve(s *Sat_solver2, lits []int32, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal int64) int32 {
	return sat_solver2_solve_c(s, lits, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal)
}
func Sat_solver2_simplify(s *Sat_solver2) int32           { return sat_solver2_simplify(s) }
func Sat_solver2_nvars(s *Sat_solver2) int32              { return sat_solver2_nvars(s) }
func Sat_solver2_nclauses(s *Sat_solver2) int32           { return sat_solver2_nclauses(s) }
func Sat_solver2_nconflicts(s *Sat_solver2) int32         { return sat_solver2_nconflicts(s) }
func Sat_solver2_nlearnts(s *Sat_solver2) int32           { return sat_solver2_nlearnts(s) }
func Sat_solver2_var_value(s *Sat_solver2, v int32) int32 { return sat_solver2_var_value(s, v) }

// Var_is_assigned is satSolver2.h's var_is_assigned.
func Var_is_assigned(s *Sat_solver2, v int32) int32 { return var_is_assigned(s, v) }
func Sat_solver2_bookmark(s *Sat_solver2)           { sat_solver2_bookmark(s) }
func Sat_solver2_rollback(s *Sat_solver2)           { sat_solver2_rollback(s) }
func Sat_solver2_memory(s *Sat_solver2, all int32) float64 {
	return sat_solver2_memory(s, all)
}
func Sat_solver2_memory_proof(s *Sat_solver2) float64 { return sat_solver2_memory_proof(s) }

// Sat_solver2_set_runtime_limit sets s->nRuntimeLimit; clock is the
// Abc_Clock the limit is measured against (nil: wall clock).
func Sat_solver2_set_runtime_limit(s *Sat_solver2, clock func() int64, limit int64) int64 {
	old := s.nRuntimeLimit
	s.runtimeClock = clock
	s.nRuntimeLimit = limit
	return old
}

// Sat_solver2_RuntimeLimit reads s->nRuntimeLimit.
func Sat_solver2_RuntimeLimit(s *Sat_solver2) int64 { return s.nRuntimeLimit }

// Sat_solver2_SetLearntLimits assigns nLearntStart, nLearntDelta,
// nLearntRatio and nLearntMax = nLearntStart, as the abstraction engines do.
func Sat_solver2_SetLearntLimits(s *Sat_solver2, start, delta, ratio int32) {
	s.nLearntStart = start
	s.nLearntDelta = delta
	s.nLearntRatio = ratio
	s.nLearntMax = s.nLearntStart
}

// Sat_solver2_NDBreduces reads s->nDBreduces.
func Sat_solver2_NDBreduces(s *Sat_solver2) int32 { return s.nDBreduces }

// Sat_solver2_Size reads s->size.
func Sat_solver2_Size(s *Sat_solver2) int32 { return s.size }

// Sat_solver2_Prf2 and Sat_solver2_SetPrf2 access s->pPrf2.
func Sat_solver2_Prf2(s *Sat_solver2) *Prf_Man_t       { return s.pPrf2 }
func Sat_solver2_SetPrf2(s *Sat_solver2, p *Prf_Man_t) { s.pPrf2 = p }
func Sat_solver2_StopPrf2(s *Sat_solver2)              { Prf_ManStopP(&s.pPrf2) }

// Sat_solver2_root_value returns a variable's current assignment (true for
// var1) and whether it is assigned; call it at decision level 0.
func Sat_solver2_root_value(s *Sat_solver2, v int32) (value, assigned bool) {
	if s.assigns[v] == satUnassigned {
		return false, false
	}
	return s.assigns[v] == 0, true
}

// Sat_solver2_add_const and Sat_solver2_add_buffer are satSolver2.h's
// inline gate encoders.
func Sat_solver2_add_const(s *Sat_solver2, iVar, fCompl, fMark, id int32) int32 {
	return sat_solver2_add_const(s, iVar, fCompl, fMark, id)
}
func Sat_solver2_add_buffer(s *Sat_solver2, iVarA, iVarB, fCompl, fMark, id int32) int32 {
	return sat_solver2_add_buffer(s, iVarA, iVarB, fCompl, fMark, id)
}

// Sat_solver2_StartPrf1 is p->pSat->pPrf1 = Vec_SetAlloc( 20 ): full
// resolution-proof logging.
func Sat_solver2_StartPrf1(s *Sat_solver2) { s.pPrf1 = NewProofGraph() }

// Sat_solver2_HasPrf1 reports p->pSat->pPrf1 != NULL.
func Sat_solver2_HasPrf1(s *Sat_solver2) bool { return s.pPrf1 != nil }
