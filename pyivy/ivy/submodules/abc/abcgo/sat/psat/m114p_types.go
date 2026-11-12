// Go representation of the opaque solver handle declared in m114p_types.h.
package psat

// M114p_Solver_t is a Go-owned opaque solver, replacing the external library's
// integer handle. No registry or machine-width pointer casts are needed.
type M114p_Solver_t = *Solver
type Literal = int
type lit = int

const (
	Unsatisfiable = 0
	Satisfiable   = 1
	Undecided     = -1
)
