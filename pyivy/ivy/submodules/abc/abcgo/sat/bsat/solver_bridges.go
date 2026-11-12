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
