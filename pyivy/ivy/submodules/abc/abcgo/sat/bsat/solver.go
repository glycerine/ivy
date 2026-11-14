package bsat

import "time"

// Literal is ABC's zero-based variable/sign encoding: 2*v is positive, 2*v+1
// negative. It is distinct from signed one-based DIMACS numbers.
type Literal = lit

func Positive(variable int32) Literal { return toLit(variable) }
func Negative(variable int32) Literal { return lit_neg(toLit(variable)) }

// Solver is an incremental CDCL solver. Construct it with NewSolver. A solver
// owns its clauses and is intended for one goroutine at a time.
type Solver = sat_solver_t

func NewSolver() *Solver                   { return sat_solver_new() }
func (s *sat_solver_t) NewVariable() int32 { s.lastStatus = Unknown; return sat_solver_addvar(s) }
func (s *sat_solver_t) NumVars() int32     { return s.size }

// Close releases the solver's internal clauses and vectors.
func (s *sat_solver_t) Close() { sat_solver_delete(s) }

// RestartSearch applies ABC's sat_solver_restart operation: it clears all
// variables, clauses, learned clauses, and results so the solver can be reused.
func (s *sat_solver_t) RestartSearch() {
	sat_solver_restart(s)
	s.fSolved = 0
	s.lastStatus = Unknown
}

// SetNumVars ensures that variable IDs below n exist, preserving existing
// assignments and clauses.
func (s *sat_solver_t) SetNumVars(n int32) {
	if n < 0 {
		panic("negative SAT variable count")
	}
	s.lastStatus = Unknown
	sat_solver_setnvars(s, n)
}

// AddClause copies literals and automatically introduces referenced variables.
// False means the formula is already inconsistent; later Solve calls return UNSAT.
func (s *sat_solver_t) AddClause(literals ...Literal) bool {
	s.lastStatus = Unknown
	if s.fSolved != 0 {
		return false
	}
	if sat_solver_addclause(s, literals) == 0 {
		s.fSolved = 1
		return false
	}
	return true
}

// AddXor constrains output to a XOR b, optionally complemented.
func (s *sat_solver_t) AddXor(output, a, b int32, complement bool) bool {
	return sat_solver_add_xor(s, output, a, b, boolInt32(complement)) != 0
}

func (s *sat_solver_t) Push(literal Literal) bool { return sat_solver_push(s, literal) != 0 }
func (s *sat_solver_t) Pop()                      { sat_solver_pop(s) }
func (s *sat_solver_t) Delete()                   { sat_solver_delete(s) }

// SetUserVariables configures the legacy user-variable trace used by ACB's
// optional optimization diagnostics.
func (s *sat_solver_t) SetUserVariables(variables []int32) {
	veci_resize(&s.user_vars, 0)
	for _, v := range variables {
		veci_push(&s.user_vars, v)
	}
}
func (s *sat_solver_t) ClearUserVariables()   { veci_resize(&s.user_vars, 0) }
func (s *sat_solver_t) UserValues() []int32   { return veci_begin(&s.user_values) }
func (s *sat_solver_t) UserValueCount() int32 { return veci_size(&s.user_values) }
func (s *sat_solver_t) ClearUserValues()      { veci_resize(&s.user_values, 0) }

func boolInt32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

// MinimizeAssumptions shortens an UNSAT assumption prefix in place and returns
// the number of retained assumptions.
func (s *sat_solver_t) MinimizeAssumptions(literals []Literal, budget int32) int32 {
	return sat_solver_minimize_assumptions(s, literals, int32(len(literals)), budget)
}

// Solve preserves clauses and learned information between calls. Assumptions
// apply only to this call; a contradiction among assumptions does not poison the
// underlying formula. Resource exhaustion returns Unknown.
func (s *sat_solver_t) Solve(assumptions ...Literal) Status {
	s.lastStatus = Status(sat_solver_solve(s, assumptions, s.limitConflicts, s.limitPropagations, 0, 0))
	return s.lastStatus
}

// SolveWithResourceLimits uses per-call conflict and propagation budgets and
// absolute cumulative conflict and propagation caps, as in ABC's
// sat_solver_solve. Zero means unlimited; the tighter nonzero limit applies.
// The original nInsLimit arguments count propagations, not inspections.
// ABC checks limits using > at search boundaries, so a decisive result may
// precede a limit check and a counter can exceed its limit. These arguments
// apply only to this call and do not change the limits configured by SetLimits.
func (s *sat_solver_t) SolveWithResourceLimits(assumptions []Literal, conflicts, propagations, globalConflicts, globalPropagations int64) Status {
	if conflicts < 0 || propagations < 0 || globalConflicts < 0 || globalPropagations < 0 {
		panic("negative SAT resource budget")
	}
	s.lastStatus = Status(sat_solver_solve(s, assumptions, conflicts, propagations, globalConflicts, globalPropagations))
	return s.lastStatus
}
func (s *sat_solver_t) Value(variable int32) (bool, bool) {
	if s.lastStatus != Satisfiable || variable < 0 || variable >= s.size {
		return false, false
	}
	return s.model[variable] == l_True, true
}
func (s *sat_solver_t) Model() []bool {
	if s.lastStatus != Satisfiable {
		return nil
	}
	model := make([]bool, s.size)
	for v := range model {
		model[v] = s.model[v] == l_True
	}
	return model
}

// UnsatCore returns a copy of the conflict literals (the negations of the
// contributing assumptions), or nil when the previous result was not UNSAT.
func (s *sat_solver_t) UnsatCore() []Literal {
	if s.lastStatus != Unsatisfiable {
		return nil
	}
	return append([]Literal(nil), veci_begin(&s.conf_final)...)
}

// Limits returns the conflict and propagation budgets used by Solve.
func (s *sat_solver_t) Limits() (conflicts, propagations int64) {
	return s.limitConflicts, s.limitPropagations
}

// SetLimits sets per-call conflict and propagation budgets; zero is unlimited.
func (s *sat_solver_t) SetLimits(conflicts, propagations int64) {
	if conflicts < 0 || propagations < 0 {
		panic("negative SAT resource budget")
	}
	s.limitConflicts = conflicts
	s.limitPropagations = propagations
}

// Deadline returns the configured absolute runtime deadline, or zero if unset.
func (s *sat_solver_t) Deadline() time.Time {
	if s.nRuntimeLimit == 0 {
		return time.Time{}
	}
	return time.Unix(0, s.nRuntimeLimit)
}

func (s *sat_solver_t) SetDeadline(deadline time.Time) {
	s.runtimeClock = nil
	if deadline.IsZero() {
		s.nRuntimeLimit = 0
	} else {
		s.nRuntimeLimit = deadline.UnixNano()
	}
}

// SetRuntimeDeadline selects an absolute deadline in the supplied clock's
// units. The caller must ensure that every reading uses the same clock domain.
// A nil clock restores the wall-clock interpretation of existing deadlines.
func (s *sat_solver_t) SetRuntimeDeadline(clock func() int64, deadline int64) {
	s.runtimeClock = clock
	s.nRuntimeLimit = deadline
}

// RuntimeDeadline returns the clock and absolute deadline used by the solver.
// A nil clock denotes the wall-clock domain used by SetDeadline.
func (s *sat_solver_t) RuntimeDeadline() (func() int64, int64) {
	return s.runtimeClock, s.nRuntimeLimit
}

func (s *sat_solver_t) runtimeExpired() bool {
	if s.nRuntimeLimit == 0 {
		return false
	}
	if s.runtimeClock != nil {
		return s.runtimeClock() > s.nRuntimeLimit
	}
	return time.Now().UnixNano() > s.nRuntimeLimit
}

// Reset discards the formula and its proof records while leaving proof logging
// enabled if it was enabled before. Release the store first to retain it.
func (s *sat_solver_t) Reset() {
	s.ClearActivityFactors()
	if s.pStore != nil {
		s.pStore = Sto_ManAlloc()
	}
	sat_solver_restart(s)
	for _, pending := range []*veci{&s.unit_lits, &s.user_vars, &s.user_values, &s.pivot_vars} {
		veci_resize(pending, 0)
	}
	s.fSolved = 0
	s.lastStatus = Unknown
	veci_resize(&s.conf_final, 0)
}

type Statistics struct {
	Variables                                       int32
	Clauses, Learned                                uint32
	Decisions, Conflicts, Propagations, Inspections int64
}

func (s *sat_solver_t) Statistics() Statistics {
	return Statistics{Variables: s.size, Clauses: s.stats.clauses, Learned: s.stats.learnts, Decisions: s.stats.decisions, Conflicts: s.stats.conflicts, Propagations: s.stats.propagations, Inspections: s.stats.inspects}
}
