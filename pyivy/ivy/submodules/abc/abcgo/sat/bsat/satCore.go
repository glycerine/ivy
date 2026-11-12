package bsat

func sat_solver_enqueue(s *sat_solver, l lit, from int32) int32 {
	s.pFreqs[lit_var(l)] = 1
	if s.satAssignments.enqueue(l, from) {
		return 1
	}
	return 0
}
func sat_solver_decision(s *sat_solver, l lit) int32 {
	if s.qtail != s.qhead || var_value(s, lit_var(l)) != varX {
		panic("invalid solver decision")
	}
	veci_push(&s.trail_lim, s.qtail)
	return sat_solver_enqueue(s, l, 0)
}
func sat_solver_canceluntil(s *sat_solver, level int32) {
	for _, v := range s.satAssignments.cancelUntil(level) {
		order_unassigned(s, v)
	}
}
func sat_solver_canceluntil_rollback(s *sat_solver, bound int32) {
	for _, v := range s.satAssignments.rollback(bound) {
		order_unassigned(s, v)
	}
}

func sat_solver_propagate(s *sat_solver) int32 {
	return satPropagate(&s.satAssignments, &s.Mem, s.wlists, s.hBinary, s.binary, s.pFreqs, &s.stats)
}
