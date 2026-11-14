package bsat

// SetLearnedPolicy configures the source solver's learned-clause reduction
// thresholds. Zero values retain the solver defaults.
func (s *sat_solver_t) SetLearnedPolicy(start, delta, ratio int32) {
	if start < 0 || delta < 0 || ratio < 0 {
		panic("negative learned-clause policy")
	}
	s.nLearntStart, s.nLearntMax = start, start
	s.nLearntDelta, s.nLearntRatio = delta, ratio
}

// SetNoRestarts controls periodic CDCL restarts.
func (s *sat_solver_t) SetNoRestarts(disable bool) {
	if disable {
		s.fNoRestarts = 1
	} else {
		s.fNoRestarts = 0
	}
}
