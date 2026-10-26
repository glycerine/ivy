package abc

// Phase hints from sat/bsat/satSolver.h:sat_solver_set_polarity.

// SetPreferredVariables clears saved phases and prefers the supplied variables
// true. Duplicate entries are harmless. Invalid input leaves all hints intact.
func (s *SatSolver) SetPreferredVariables(vars []SatVar) error {
	if s == nil {
		return ErrDimension
	}
	if s.solving {
		return ErrInvariant
	}
	for _, v := range vars {
		if uint64(v) >= uint64(len(s.polarity)) {
			return ErrDimension
		}
	}
	clear(s.polarity)
	for _, v := range vars {
		s.polarity[v] = true
	}
	return nil
}
