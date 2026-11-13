package bsat

// SolveWithSourceLimits forwards signed limits using ABC's original arithmetic.
// Local limits are added to the current counters, global limits are absolute,
// and an effective zero disables the corresponding limit. Unlike the safer
// SolveWithResourceLimits API, negative values are accepted unchanged.
func (s *sat_solver_t) SolveWithSourceLimits(assumptions []Literal, conflicts, propagations, globalConflicts, globalPropagations int64) Status {
	s.lastStatus = Status(sat_solver_solve(s, assumptions, conflicts, propagations, globalConflicts, globalPropagations))
	return s.lastStatus
}

// SolveWithSourceLimits applies the same source-compatible signed budgets to
// the proof-logging solver2.
func (p *Solver2) SolveWithSourceLimits(assumptions []Literal, conflicts, propagations, globalConflicts, globalPropagations int64) Status {
	p.solver.lastStatus = Status(sat_solver2_solve(p.solver, assumptions, conflicts, propagations, globalConflicts, globalPropagations))
	return p.solver.lastStatus
}
