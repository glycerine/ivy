package bsat

import "time"

func sat_solver2_solve(s *sat_solver2, assumptions []lit, conflicts, propagations, globalConflicts, globalPropagations int64) int32 {
	if s.fSolved != 0 {
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
		if lit_check(l, s.size) == 0 {
			panic("solver2 assumption outside solver")
		}
		veci_push(&s.trail_lim, s.qtail)
		if solver2_enqueue(s, l, 0) == 0 {
			reason := clause2_read(s, s.reasons[lit_var(l)])
			if reason != nil {
				s.hProofLast = solver2_analyze_final(s, reason, 1)
				veci_push(&s.conf_final, lit_neg(l))
			} else {
				s.hProofLast = -1
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
		if s.nConfLimit != 0 && s.stats.conflicts > s.nConfLimit || s.nInsLimit != 0 && s.stats.propagations > s.nInsLimit || s.nRuntimeLimit != 0 && time.Now().UnixNano() > s.nRuntimeLimit {
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
