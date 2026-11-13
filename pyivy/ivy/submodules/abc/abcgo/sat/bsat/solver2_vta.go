package bsat

// Simplify propagates the solver's root assignments before bookmarking.
func (p *LabeledSolver2) Simplify() bool {
	return sat_solver2_simplify(p.solver.solver) != 0
}

// SetLearnedPolicy configures ABC's learned-clause reduction policy.
func (p *LabeledSolver2) SetLearnedPolicy(start, delta, ratio int32) {
	s := p.solver.solver
	s.nLearntStart, s.nLearntMax = start, start
	s.nLearntDelta, s.nLearntRatio = delta, ratio
}

// Reductions returns the number of learned-clause database reductions.
func (p *LabeledSolver2) Reductions() int32 { return p.solver.solver.nDBreduces }

// Memory reports ABC's estimates for solver storage and proof storage in bytes.
func (p *LabeledSolver2) Memory(all bool) (float64, float64) {
	var flag int32
	if all {
		flag = 1
	}
	return sat_solver2_memory(p.solver.solver, flag), sat_solver2_memory_proof(p.solver.solver)
}
