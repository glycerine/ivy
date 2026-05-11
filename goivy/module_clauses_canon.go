package goivy

import "io"

// Canon returns a canonical s-expression for the Clauses.
// Format: (clauses fmlas:[sexp1 sexp2 ...] defs:[sexp1 sexp2 ...])
// Both Go and Python must produce identical output for identical state.
func (c *Clauses) Canon() Canonical {
	return canonString(func(w io.Writer) { writeClausesCanon(w, c) })
}
