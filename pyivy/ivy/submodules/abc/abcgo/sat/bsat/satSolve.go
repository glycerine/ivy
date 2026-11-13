package bsat

func sat_solver_var_value(s *sat_solver, v int32) int32 {
	if v < 0 || v >= s.size {
		panic("model variable outside solver")
	}
	if s.model[v] == l_True {
		return 1
	}
	return 0
}
func sat_solver_var_literal(s *sat_solver, v int32) lit {
	if sat_solver_var_value(s, v) != 0 {
		return toLit(v)
	}
	return lit_neg(toLit(v))
}
func sat_solver_push(s *sat_solver, p lit) int32 {
	if lit_check(p, s.size) == 0 {
		panic("assumption outside solver")
	}
	veci_push(&s.trail_lim, s.qtail)
	s.root_level++
	if sat_solver_enqueue(s, p, 0) == 0 {
		reason := s.reasons[lit_var(p)]
		if reason != 0 {
			if clause_is_lit(reason) != 0 {
				s.binary.lits[1] = lit_neg(p)
				s.binary.lits[0] = clause_read_lit(reason)
				reason = s.hBinary
			}
			sat_solver_analyze_final(s, reason, 1)
			veci_push(&s.conf_final, lit_neg(p))
		} else {
			veci_resize(&s.conf_final, 0)
			veci_push(&s.conf_final, lit_neg(p))
			if s.levels[lit_var(p)] > 0 {
				veci_push(&s.conf_final, p)
			}
		}
		return 0
	}
	if conflict := sat_solver_propagate(s); conflict != 0 {
		sat_solver_analyze_final(s, conflict, 0)
		return 0
	}
	return 1
}
func sat_solver_pop(s *sat_solver) {
	if s.root_level <= 0 || sat_solver_dl(s) <= 0 {
		panic("pop without assumption")
	}
	s.root_level--
	sat_solver_canceluntil(s, s.root_level)
}
func sat_solver_set_resource_limits(s *sat_solver, conflicts, propagations, globalConflicts, globalPropagations int64) {
	s.nRestarts = 0
	s.nConfLimit = 0
	s.nInsLimit = 0
	if conflicts != 0 {
		s.nConfLimit = s.stats.conflicts + conflicts
	}
	if propagations != 0 {
		s.nInsLimit = s.stats.propagations + propagations
	}
	if globalConflicts != 0 && (s.nConfLimit == 0 || s.nConfLimit > globalConflicts) {
		s.nConfLimit = globalConflicts
	}
	if globalPropagations != 0 && (s.nInsLimit == 0 || s.nInsLimit > globalPropagations) {
		s.nInsLimit = globalPropagations
	}
}

// satSolverPropagateRoot records a permanent contradiction before any
// temporary assumptions are applied.
func satSolverPropagateRoot(s *sat_solver) bool {
	if sat_solver_propagate(s) == 0 {
		return true
	}
	s.fSolved = 1
	if p := solverClauseStore(s); p != nil && p.pEmpty == nil {
		Sto_ManAddClause(p, nil)
	}
	return false
}
func sat_solver_solve(s *sat_solver, assumptions []lit, conflicts, propagations, globalConflicts, globalPropagations int64) int32 {
	for _, l := range assumptions {
		if lit_check(l, s.size) == 0 {
			panic("assumption outside solver")
		}
	}
	if s.fSolved != 0 {
		veci_resize(&s.conf_final, 0)
		if p := solverClauseStore(s); p != nil && p.pEmpty == nil {
			Sto_ManAddClause(p, nil)
		}
		return l_False
	}
	sat_solver_set_resource_limits(s, conflicts, propagations, globalConflicts, globalPropagations)
	s.root_level = 0
	veci_resize(&s.conf_final, 0)
	// ABC pushes assumptions before propagating pending root units. This can
	// lose a pre-existing root conflict after a failed assumption; callers must
	// simplify before assumptions when using potentially inconsistent clauses.
	for _, l := range assumptions {
		if sat_solver_push(s, l) == 0 {
			sat_solver_canceluntil(s, 0)
			s.root_level = 0
			return l_False
		}
	}
	status := sat_solver_solve_internal(s)
	sat_solver_canceluntil(s, 0)
	s.root_level = 0
	if status == l_False && len(assumptions) == 0 {
		if p := solverClauseStore(s); p != nil && p.pEmpty == nil {
			Sto_ManAddClause(p, nil)
		}
	}
	return status
}
