package bsat

import "time"

// Solver2 exposes ABC's proof-logging solver variant without leaking its mutable
// clause/proof arenas. Root clause IDs returned by ProofCore are one-based.
type Solver2 struct {
	solver *sat_solver2
	nextID int32
}

func NewSolver2() *Solver2 {
	s := sat_solver2_new()
	s.pPrf1 = NewProofGraph()
	return &Solver2{solver: s, nextID: 1}
}
func (p *Solver2) NewVariable() int32 {
	p.solver.lastStatus = Unknown
	v := p.solver.size
	sat_solver2_setnvars(p.solver, v+1)
	return v
}
func (p *Solver2) NumVars() int32 { return p.solver.size }

// Simplify propagates root-level assignments before search. As with the C
// solver, a false result ends the current solve: do not call Solve afterward.
func (p *Solver2) Simplify() bool {
	p.solver.lastStatus = Unknown
	if p.solver.fSolved != 0 {
		return false
	}
	return sat_solver2_simplify(p.solver) != 0
}
func (p *Solver2) AddClause(literals ...Literal) bool {
	id := p.nextID
	return p.addClause(id, literals)
}

func (p *Solver2) addClause(id int32, literals []Literal) bool {
	s := p.solver
	s.lastStatus = Unknown
	if s.fSolved != 0 {
		return false
	}
	maxVar := int32(-1)
	for _, l := range literals {
		if l < 0 {
			panic("negative encoded literal")
		}
		maxVar = max(maxVar, lit_var(l))
	}
	if maxVar >= 0 {
		sat_solver2_setnvars(s, maxVar+1)
	}
	satisfied := false
	unassigned := false
	seen := map[lit]bool{}
	for _, l := range literals {
		if seen[lit_neg(l)] || s.assigns[lit_var(l)] == lit_sign(l) {
			satisfied = true
		}
		seen[l] = true
		if s.assigns[lit_var(l)] == satUnassigned {
			unassigned = true
		}
	}
	p.nextID++
	if satisfied || unassigned {
		sat_solver2_addclause(s, literals, id)
		return true
	}
	// A root clause can contradict existing root units. Preserve its identity and
	// explicitly resolve their proofs rather than relying on the C API assertion.
	unique := make([]lit, 0, len(seen))
	for _, l := range literals {
		if seen[l] {
			unique = append(unique, l)
			delete(seen, l)
		}
	}
	h := clause2_create_new(s, unique, 0, 0)
	clause2_set_id(s, h, id)
	proof_chain_start(s, clause2_read(s, h))
	for _, l := range unique {
		proof_chain_resolve(s, nil, lit_var(l))
	}
	s.hProofLast = proof_chain_stop(s)
	s.fSolved = 1
	return false
}
func (p *Solver2) Solve(assumptions ...Literal) Status {
	p.solver.lastStatus = Status(sat_solver2_solve(p.solver, assumptions, p.solver.limitConflicts, p.solver.limitPropagations, 0, 0))
	return p.solver.lastStatus
}

func (p *Solver2) ProofCore() []int32 {
	if p.solver.lastStatus != Unsatisfiable {
		return nil
	}
	return append([]int32(nil), Sat_ProofCore(p.solver)...)
}
func (p *Solver2) Value(variable int32) (bool, bool) { return p.solver.sat_solver_t.Value(variable) }
func (p *Solver2) Model() []bool                     { return p.solver.sat_solver_t.Model() }
func (p *Solver2) UnsatCore() []Literal              { return p.solver.sat_solver_t.UnsatCore() }
func (p *Solver2) Statistics() Statistics            { return p.solver.sat_solver_t.Statistics() }
func (p *Solver2) SetLimits(conflicts, propagations int64) {
	p.solver.sat_solver_t.SetLimits(conflicts, propagations)
}
func (p *Solver2) SetDeadline(deadline time.Time) { p.solver.sat_solver_t.SetDeadline(deadline) }
func (p *Solver2) SetRuntimeDeadline(clock func() int64, deadline int64) {
	p.solver.sat_solver_t.SetRuntimeDeadline(clock, deadline)
}
func (p *Solver2) Reset() { *p = *NewSolver2() }
