# SAT clause encoder review

The SAT equality and two-input AND encoders now use the public `bsat.Solver`
API and value-based `bsat.Literal` clauses. The implementation does not expose
BSAT's unexported solver or clause-storage types.

The separate bug review exhaustively checks every input/output assignment and
both complement phases. It found and avoids assuming phases are exactly the
integer 0 or 1 when deriving their inverse; any nonzero phase is treated as
complemented, matching C's logical `!` behavior.

The helpers preserve the source convention that a successful clause addition
returns 1 and an inconsistent addition returns 0. `go test -race` on the
explicit encoder files is the current focused validation; full `opt/res`
integration remains blocked by other generated C-shaped files and missing
cross-package imports.
