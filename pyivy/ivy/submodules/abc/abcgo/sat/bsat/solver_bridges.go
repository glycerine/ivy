package bsat

import "math"

// ClearActivityFactors removes the active cone used by restart activity bumps.
func (s *sat_solver_t) ClearActivityFactors() {
	for _, v := range veci_begin(&s.act_vars) {
		if int(v) < len(s.factors) {
			s.factors[v] = 0
		}
	}
	veci_resize(&s.act_vars, 0)
	clear(s.activityFactorActive)
}

// SetActivityFactor activates a variable for the existing cone activity bump.
// Updating a factor replaces its value without adding a second bump.
func (s *sat_solver_t) SetActivityFactor(variable int32, factor float64) {
	if variable < 0 || variable >= s.size {
		panic("activity variable outside solver")
	}
	if factor < 0 || math.IsNaN(factor) || math.IsInf(factor, 0) {
		panic("invalid activity factor")
	}
	if s.factors == nil {
		s.factors = make([]float64, s.size, s.cap)
	}
	if s.activityFactorActive == nil {
		s.activityFactorActive = make(map[int32]bool)
	}
	s.factors[variable] = factor
	if !s.activityFactorActive[variable] {
		veci_push(&s.act_vars, variable)
		s.activityFactorActive[variable] = true
	}
}

// SolveWithGlobalLimits additionally applies absolute cumulative counter limits.
// Propagations and inspected watches are different counters: the original
// nInsLimit solver parameter currently limits propagations, as in ABC.
func (s *sat_solver_t) SolveWithGlobalLimits(assumptions []Literal, conflicts, propagations int64) Status {
	if conflicts < 0 || propagations < 0 {
		panic("negative global SAT limit")
	}
	s.lastStatus = Status(sat_solver_solve(s, assumptions, s.limitConflicts, s.limitPropagations, conflicts, propagations))
	return s.lastStatus
}

// EnableClauseStore begins recording original and learned clauses. Call before
// adding input clauses; existing clauses cannot be recovered into a proof log.
func (s *sat_solver_t) EnableClauseStore()            { sat_solver_store_alloc(s) }
func (s *sat_solver_t) MarkClauseStoreRoots()         { sat_solver_store_mark_roots(s) }
func (s *sat_solver_t) MarkClauseStorePartitionA()    { sat_solver_store_mark_clauses_a(s) }
func (s *sat_solver_t) ChangeLastStoredClause() int32 { return sat_solver_store_change_last(s) }

// ReleaseClauseStore transfers ownership; the solver stops appending to it.
func (s *sat_solver_t) ReleaseClauseStore() *Sto_Man_t {
	p := solverClauseStore(s)
	s.pStore = nil
	return p
}

type StoredClause struct {
	ID               int32
	Literals         []Literal
	Root, PartitionA bool
}

func (p *Sto_Man_t) NumVariables() int32 { return p.nVars }
func (p *Sto_Man_t) NumClauses() int32   { return p.nClauses }
func (p *Sto_Man_t) NumRoots() int32     { return p.nRoots }

// Clauses returns owned snapshots; callers cannot corrupt proof-log storage.
func (p *Sto_Man_t) Clauses() []StoredClause {
	result := make([]StoredClause, 0, p.nClauses)
	for c := p.pHead; c != nil; c = c.pNext {
		result = append(result, StoredClause{ID: c.Id, Literals: append([]Literal(nil), c.pLits...), Root: c.fRoot != 0, PartitionA: c.fA != 0})
	}
	return result
}

// VariableActivity returns the actual heuristic score in its configured
// representation; it does not apply an extra cone factor or change ordering.
func (s *sat_solver_t) VariableActivity(variable int32) float64 {
	if variable < 0 || variable >= s.size {
		panic("activity variable outside solver")
	}
	switch s.VarActType {
	case 0:
		return float64(s.activity[variable])
	case 1:
		return math.Float64frombits(s.activity[variable])
	case 2:
		return Xdbl_ToDouble(s.activity[variable])
	default:
		panic("invalid activity representation")
	}
}

func (s *sat_solver_t) NeedsPropagation() bool  { return s.qhead != s.qtail }
func (s *sat_solver_t) ConflictCount() int64    { return s.stats.conflicts }
func (s *sat_solver_t) PropagationCount() int64 { return s.stats.propagations }

// TakeUnitLiterals transfers a copy of newly learned units and clears the queue.
func (s *sat_solver_t) TakeUnitLiterals() []Literal {
	result := append([]Literal(nil), veci_begin(&s.unit_lits)...)
	veci_resize(&s.unit_lits, 0)
	return result
}

// PropagateUnits performs level-zero propagation. False records a permanent
// conflict so later Solve calls remain UNSAT even after consuming the queue.
func (s *sat_solver_t) PropagateUnits() bool {
	if s.fSolved != 0 {
		return false
	}
	if sat_solver_simplify(s) == 0 {
		s.fSolved = 1
		s.lastStatus = Unknown
		return false
	}
	return true
}
func (s *sat_solver_t) Compress() bool {
	if s.NeedsPropagation() {
		return s.PropagateUnits()
	}
	return s.fSolved == 0
}

// SolveLexicographic updates literals in place to the least ordered satisfying
// assignment. The first literal is most significant. Begin with all-negative
// literals; successive calls may reuse the result after blocking that assignment
// and all earlier ones. Arbitrary mixed-sign lower bounds are not supported by
// ABC's lex-SAT algorithm. Variables must be distinct and already introduced.
// Per-call limits cover the entire lexicographic search, including its subsolves.
func (s *sat_solver_t) SolveLexicographic(literals []Literal) Status {
	seen := make(map[int32]bool, len(literals))
	for _, l := range literals {
		v := lit_var(l)
		if lit_check(l, s.size) == 0 || seen[v] {
			panic("invalid lexicographic variable order")
		}
		seen[v] = true
	}
	if len(literals) == 0 || s.fSolved != 0 {
		return s.Solve()
	}
	if s.root_level != 0 || s.trail_lim.size != 0 {
		panic("lexicographic search with pending assumptions")
	}
	if !satSolverPropagateRoot(s) {
		s.lastStatus = Unsatisfiable
		return Unsatisfiable
	}
	sat_solver_set_resource_limits(s, s.limitConflicts, s.limitPropagations, 0, 0)
	veci_resize(&s.conf_final, 0)
	s.lastStatus = Status(sat_solver_solve_lexsat(s, literals, int32(len(literals))))
	return s.lastStatus
}

// SetPivotVariables records the preferred proof pivot variables. The solver
// owns a copy, so the caller may reuse its slice after this call.
func (s *sat_solver_t) SetPivotVariables(vars []int32) {
	for _, v := range vars {
		if v < 0 || v >= s.NumVars() {
			panic("pivot variable outside solver")
		}
	}
	sat_solver_set_pivot_variables(s, vars, int32(len(vars)))
}
