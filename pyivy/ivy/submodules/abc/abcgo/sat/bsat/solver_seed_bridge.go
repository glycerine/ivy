package bsat

// NewSolverSeed creates a solver using the zsat_solver_new_seed initialization.
func NewSolverSeed(seed float64) *Solver { return zsat_solver_new_seed(seed) }

// RestartSeed performs ABC's zsat_solver_restart_seed in place. Like C, it
// resets the clause database and counters while preserving other solver state.
func (s *sat_solver_t) RestartSeed(seed float64) {
	zsat_solver_restart_seed(s, seed)
	s.lastStatus = Unknown
}
