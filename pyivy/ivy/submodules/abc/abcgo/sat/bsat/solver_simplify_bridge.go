package bsat

// Simplify propagates root assignments and removes satisfied clauses. It
// reports false when root propagation proves the formula inconsistent.
func (s *sat_solver_t) Simplify() bool {
	s.lastStatus = Unknown
	if s.fSolved != 0 {
		return false
	}
	if sat_solver_simplify(s) == 0 {
		s.fSolved = 1
		return false
	}
	return true
}
