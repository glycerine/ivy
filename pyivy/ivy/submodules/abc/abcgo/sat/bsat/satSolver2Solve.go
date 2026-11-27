package bsat

func sat_solver2_solve(s *sat_solver2, assumptions []lit, conflicts, propagations, globalConflicts, globalPropagations int64) int32 {
	for _, l := range assumptions {
		if lit_check(l, s.size) == 0 {
			panic("solver2 assumption outside solver")
		}
	}
	if s.fSolved != 0 {
		veci_resize(&s.conf_final, 0)
		return l_False
	}
	s.hLearntLast, s.hProofLast = -1, -1
	s.nConfLimit, s.nInsLimit = 0, 0
	if conflicts != 0 {
		s.nConfLimit = s.stats.conflicts + conflicts
	}
	if propagations != 0 {
		s.nInsLimit = s.stats.propagations + propagations
	}
	if globalConflicts != 0 && (s.nConfLimit == 0 || globalConflicts < s.nConfLimit) {
		s.nConfLimit = globalConflicts
	}
	if globalPropagations != 0 && (s.nInsLimit == 0 || globalPropagations < s.nInsLimit) {
		s.nInsLimit = globalPropagations
	}
	// Root conflicts must be discovered before assumption levels are opened;
	// otherwise a contradictory assumption can hide pending propagation.
	s.root_level = 0
	veci_resize(&s.conf_final, 0)
	if solver2_propagate(s) != nil {
		s.fSolved = 1
		return l_False
	}
	s.root_level = int32(len(assumptions))
	for _, l := range assumptions {
		veci_push(&s.trail_lim, s.qtail)
		if solver2_enqueue(s, l, 0) == 0 {
			reason := clause2_read(s, s.reasons[lit_var(l)])
			if reason != nil {
				s.hProofLast = solver2_analyze_final(s, reason, 1)
				veci_push(&s.conf_final, lit_neg(l))
			} else {
				s.hProofLast = -1
				if s.levels[lit_var(l)] == 0 && s.units[lit_var(l)] != 0 {
					proof_chain_start(s, clause2_read(s, s.units[lit_var(l)]))
					s.hProofLast = proof_chain_stop(s)
				}
				veci_resize(&s.conf_final, 0)
				veci_push(&s.conf_final, lit_neg(l))
				if s.levels[lit_var(l)] > 0 {
					veci_push(&s.conf_final, l)
				}
			}
			solver2_canceluntil(s, 0)
			return l_False
		}
		if conflict := solver2_propagate(s); conflict != nil {
			s.hProofLast = solver2_analyze_final(s, conflict, 0)
			solver2_canceluntil(s, 0)
			return l_False
		}
	}
	var status lbool
	for restart := int32(0); status == l_Undef; restart++ {
		status = solver2_search(s, int64(100*luby(2, restart)))
		if status != l_Undef {
			break
		}
		if s.nConfLimit != 0 && s.stats.conflicts > s.nConfLimit || s.nInsLimit != 0 && s.stats.propagations > s.nInsLimit || s.runtimeExpired() {
			break
		}
	}
	solver2_canceluntil(s, 0)
	if status == l_False && len(assumptions) == 0 {
		s.fSolved = 1
	}
	return int32(status)
}
func sat_solver2_simplify(s *sat_solver2) int32 { return solver2Bool(solver2_propagate(s) == nil) }

// sat_solver2_solve_c is satSolver2.c's sat_solver2_solve line by line:
// pending root-level implications are propagated under the first
// assumption's level, learned clauses are reduced between restarts, and
// there is no "solved" state. sat_solver2_solve above keeps the Solver2
// wrappers' stronger contract; Sat_solver2_solve exports this one.
func sat_solver2_solve_c(s *sat_solver2, assumptions []lit, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal int64) int32 {
	restart_iter := int32(0)
	var nof_conflicts int64
	status := lbool(l_Undef)
	var proof_id int32

	s.hLearntLast = -1
	s.hProofLast = -1

	// set the external limits
	s.nConfLimit = 0
	s.nInsLimit = 0
	if nConfLimit != 0 {
		s.nConfLimit = s.stats.conflicts + nConfLimit
	}
	if nInsLimit != 0 {
		s.nInsLimit = s.stats.propagations + nInsLimit
	}
	if nConfLimitGlobal != 0 && (s.nConfLimit == 0 || s.nConfLimit > nConfLimitGlobal) {
		s.nConfLimit = nConfLimitGlobal
	}
	if nInsLimitGlobal != 0 && (s.nInsLimit == 0 || s.nInsLimit > nInsLimitGlobal) {
		s.nInsLimit = nInsLimitGlobal
	}

	// Perform assumptions:
	s.root_level = int32(len(assumptions))
	for _, p := range assumptions {
		if lit_check(p, s.size) == 0 {
			panic("solver2 assumption outside solver")
		}
		veci_push(&s.trail_lim, s.qtail)
		if solver2_enqueue(s, p, 0) == 0 {
			r := clause2_read(s, s.reasons[lit_var(p)])
			if r != nil {
				confl := r
				proof_id = solver2_analyze_final(s, confl, 1)
				veci_push(&s.conf_final, lit_neg(p))
			} else {
				proof_id = -1 // the only case when ProofId is not assigned (conflicting assumptions)
				veci_resize(&s.conf_final, 0)
				veci_push(&s.conf_final, lit_neg(p))
				if s.levels[lit_var(p)] > 0 {
					veci_push(&s.conf_final, p)
				}
			}
			s.hProofLast = proof_id
			solver2_canceluntil(s, 0)
			return l_False
		} else {
			confl := solver2_propagate(s)
			if confl != nil {
				proof_id = solver2_analyze_final(s, confl, 0)
				s.hProofLast = proof_id
				solver2_canceluntil(s, 0)
				return l_False
			}
		}
	}

	for status == l_Undef {
		if s.nRuntimeLimit != 0 && s.runtimeExpired() {
			break
		}
		// reduce the set of learnt clauses
		if s.nLearntMax != 0 && s.act_clas.size >= s.nLearntMax && s.pPrf2 == nil {
			sat_solver2_reducedb(s)
		}
		// perform next run
		nof_conflicts = int64(100 * luby2(2, restart_iter))
		restart_iter++
		status = solver2_search(s, nof_conflicts)
		// quit the loop if reached an external limit
		if s.nConfLimit != 0 && s.stats.conflicts > s.nConfLimit {
			break
		}
		if s.nInsLimit != 0 && s.stats.propagations > s.nInsLimit {
			break
		}
	}
	solver2_canceluntil(s, 0)
	return int32(status)
}
