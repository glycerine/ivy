package bsat

// SetLearntReduction sets the optional learnt-clause thresholds used by FRA.
// Zero values retain the solver defaults, as in the original C caller.
func (s *sat_solver_t) SetLearntReduction(start, delta, percentage int32) {
	if start != 0 {
		s.nLearntStart, s.nLearntMax = start, start
	}
	if delta != 0 {
		s.nLearntDelta = delta
	}
	if percentage != 0 {
		s.nLearntRatio = percentage
	}
}
