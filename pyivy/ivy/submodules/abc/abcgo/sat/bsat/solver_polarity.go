package bsat

// SetPolarity sets the preferred truth value for a future decision on variable.
// This is a phase hint, not an assumption or a constraint; clauses and solving
// may choose the other value. The variable must already exist in the solver.
func (s *sat_solver_t) SetPolarity(variable int32, preferredValue bool) {
	if variable < 0 || variable >= s.size {
		panic("polarity variable outside solver")
	}
	s.polarity[variable] = 0
	if preferredValue {
		s.polarity[variable] = 1
	}
}
