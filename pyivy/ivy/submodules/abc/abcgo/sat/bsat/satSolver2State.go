package bsat

func sat_solver2_new() *sat_solver2 {
	s := &sat_solver2{fProofLogging: 1, hLearntLast: -1, hProofLast: -1, tempInter: -1}
	s.satAssignments = newSatAssignments(0)
	Sat_MemAlloc_(&s.Mem, 14)
	s.var_inc, s.cla_inc = 1<<5, 1<<11
	s.random_seed = 91648253
	s.fSkipSimplify = 1
	s.nLearntStart, s.nLearntDelta, s.nLearntRatio = LEARNT_MAX_START_DEFAULT, LEARNT_MAX_INCRE_DEFAULT, LEARNT_MAX_RATIO_DEFAULT
	s.nLearntMax = s.nLearntStart
	s.hProofPivot = 1
	for _, v := range []*veci{&s.order, &s.trail_lim, &s.tagged, &s.stack, &s.temp_clause, &s.temp_proof, &s.conf_final, &s.mark_levels, &s.min_lit_order, &s.min_step_order, &s.act_clas, &s.claProofs} {
		veci_new(v)
	}
	return s
}
func sat_solver2_setnvars(s *sat_solver2, n int32) {
	old := s.size
	sat_solver_setnvars(&s.sat_solver_t, n)
	s.units = satResize(s.units, int(s.size), int(s.cap))
	s.variableA = satResize(s.variableA, int(s.size), int(s.cap))
	s.markedLevels = satResize(s.markedLevels, int(s.size)+1, int(s.cap)+1)
	for v := old; v < s.size; v++ {
		s.units[v] = 0
		s.variableA[v] = 0
	}
}
func solver2_enqueue(s *sat_solver2, l lit, from cla) int32 { return solver2Bool(s.enqueue(l, from)) }
func solver2_assume(s *sat_solver2, l lit) int32            { return solver2Bool(s.decision(l)) }
func solver2_canceluntil(s *sat_solver2, level int32) {
	if s.trail_lim.size <= level {
		return
	}
	bound := s.trail_lim.ptr[level]
	for i := bound; i < s.qtail; i++ {
		s.units[lit_var(s.trail[i])] = 0
	}
	sat_solver_canceluntil(&s.sat_solver_t, level)
}
func solver2_canceluntil_rollback(s *sat_solver2, bound int32) {
	for i := bound; i < s.qtail; i++ {
		s.units[lit_var(s.trail[i])] = 0
	}
	sat_solver_canceluntil_rollback(&s.sat_solver_t, bound)
}
func solver2_dlevel(s *sat_solver2) int32       { return s.trail_lim.size }
func solver2_wlist(s *sat_solver2, l lit) *veci { return &s.wlists[l] }
