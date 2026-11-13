package minisatp

import "time"

// Solver is MiniSat-p 1.14's no-assumptions proof-producing solver.
// Literals use the MiniSat encoding 2*variable+sign.
type Solver struct{ s *solver }

// New allocates a proof-producing solver with n variables.
func New(n int) *Solver {
	if n < 0 {
		panic("negative variable count")
	}
	return &Solver{s: newSolver(n)}
}

// AddClause adds a problem clause and reports whether the solver remains
// consistent. Input literals must refer to allocated variables.
func (s *Solver) AddClause(lits []int32) bool {
	for _, lit := range lits {
		if lit < 0 || int(lit>>1) >= len(s.s.assign) {
			panic("literal out of range")
		}
	}
	return s.s.addClause(lits)
}

// Solve returns 1 for satisfiable, 0 for unsatisfiable, and -1 when stopped.
func (s *Solver) Solve() int32 { return int32(s.s.solve()) }

// Interrupted reports whether a configured limit stopped this solver.
func (s *Solver) Interrupted() bool { return s.s.expired() }

// SetLimits sets a total conflict budget and wall-clock deadline. A zero
// budget or zero deadline means unlimited. A stopped solver returns -1 and
// cannot be resumed; create a new solver to retry with different limits.
func (s *Solver) SetLimits(conflicts int32, deadline time.Time) {
	s.s.conflictLimit, s.s.deadline = conflicts, deadline
}

// Conflicts reports the number of conflicts encountered by search.
func (s *Solver) Conflicts() int32 { return s.s.conflicts }

// Proof returns owned roots, resolution chains, and the zero-based input-clause
// index of each root. MiniSat omits satisfied and tautological input clauses.
// Root-first numbering permits consumers to enumerate roots before chains even
// when upstream recorded a derived clause between two input roots.
func (s *Solver) Proof() ([][]int32, []Chain, []int) {
	var roots [][]int32
	for _, event := range s.s.proof.events {
		if s.s.expired() {
			return nil, nil, nil
		}
		if !event.deleted && len(event.chain.Clauses) == 0 {
			roots = append(roots, append([]int32(nil), event.root...))
		}
	}
	ids := make(map[ClauseID]ClauseID)
	rootCount := ClauseID(0)
	for _, event := range s.s.proof.events {
		if s.s.expired() {
			return nil, nil, nil
		}
		if !event.deleted && len(event.chain.Clauses) == 0 {
			ids[event.id] = rootCount
			rootCount++
		}
	}
	var chains []Chain
	for _, event := range s.s.proof.events {
		if s.s.expired() {
			return nil, nil, nil
		}
		if event.deleted || len(event.chain.Clauses) == 0 {
			continue
		}
		chain := Chain{Clauses: make([]ClauseID, len(event.chain.Clauses)), Pivots: append([]int32(nil), event.chain.Pivots...)}
		for i, id := range event.chain.Clauses {
			if s.s.expired() {
				return nil, nil, nil
			}
			chain.Clauses[i] = ids[id]
		}
		ids[event.id] = rootCount + ClauseID(len(chains))
		chains = append(chains, chain)
	}
	if len(s.s.proof.events) > 0 && len(chains) > 0 {
		last := s.s.proof.events[len(s.s.proof.events)-1]
		if !last.deleted && len(last.chain.Clauses) == 0 {
			chains = append(chains, Chain{Clauses: []ClauseID{ids[last.id]}})
		}
	}
	return roots, chains, append([]int(nil), s.s.rootInputs...)
}
