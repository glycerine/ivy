package bsat

// NewSolverSeed creates a solver using the zsat_solver_new_seed initialization.
func NewSolverSeed(seed float64) *Solver { return zsat_solver_new_seed(seed) }
