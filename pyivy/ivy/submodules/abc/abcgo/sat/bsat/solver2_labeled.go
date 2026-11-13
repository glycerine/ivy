package bsat

import (
	"slices"
	"time"
)

// LabeledSolver2 associates each proof root with the caller's problem-object
// ID. Multiple clauses may share an ID; -1 marks a background clause that is
// omitted from the reported core. Resolution proofs remain in Solver2's graph.
type LabeledSolver2 struct {
	solver     *Solver2
	labels     map[int32]int32
	bookmarkID int32
	bookmarked bool
}

func NewLabeledSolver2() *LabeledSolver2 {
	return &LabeledSolver2{solver: NewSolver2(), labels: make(map[int32]int32)}
}

func (p *LabeledSolver2) NewVariable() int32 { return p.solver.NewVariable() }
func (p *LabeledSolver2) NumVars() int32     { return p.solver.NumVars() }

// AddClause associates the clause with id. ABC's solver reserves zero and
// uses -1 for a clause which is excluded from abstraction proof cores.
func (p *LabeledSolver2) AddClause(id int32, literals ...Literal) bool {
	if id == 0 || id < -1 {
		panic("invalid labeled clause ID")
	}
	root := p.solver.nextID
	result := p.solver.AddClause(literals...)
	if p.solver.nextID != root {
		p.labels[root] = id
	}
	return result
}

func (p *LabeledSolver2) Solve(assumptions ...Literal) Status {
	return p.solver.Solve(assumptions...)
}
func (p *LabeledSolver2) Value(variable int32) (bool, bool) { return p.solver.Value(variable) }
func (p *LabeledSolver2) Model() []bool                     { return p.solver.Model() }
func (p *LabeledSolver2) UnsatCore() []Literal              { return p.solver.UnsatCore() }
func (p *LabeledSolver2) Statistics() Statistics            { return p.solver.Statistics() }
func (p *LabeledSolver2) SetLimits(conflicts, propagations int64) {
	p.solver.SetLimits(conflicts, propagations)
}
func (p *LabeledSolver2) SetDeadline(deadline time.Time) { p.solver.SetDeadline(deadline) }

// ProofCore returns caller IDs that contributed to the UNSAT result. After a
// bookmark, only clauses added since that bookmark are reported; earlier
// clauses form the fixed background, as in ABC's restarted bitset logger.
func (p *LabeledSolver2) ProofCore() []int32 {
	seen := make(map[int32]bool)
	var result []int32
	for _, root := range p.solver.ProofCore() {
		if p.bookmarked && root < p.bookmarkID {
			continue
		}
		id, ok := p.labels[root]
		if ok && id >= 0 && !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	slices.Sort(result)
	slices.Reverse(result)
	return result
}

// Bookmark requires root-level propagation to have completed, matching ABC.
func (p *LabeledSolver2) Bookmark() {
	sat_solver2_bookmark(p.solver.solver)
	p.bookmarkID = p.solver.nextID
	p.bookmarked = true
}

// Rollback restores the solver and label table to the latest bookmark.
func (p *LabeledSolver2) Rollback() {
	if !p.bookmarked {
		panic("rollback without bookmark")
	}
	sat_solver2_rollback(p.solver.solver)
	for root := range p.labels {
		if root >= p.bookmarkID {
			delete(p.labels, root)
		}
	}
	p.solver.nextID = p.bookmarkID
	p.solver.solver.lastStatus = Unknown
	p.solver.solver.hProofLast = -1
	veci_resize(&p.solver.solver.conf_final, 0)
	p.bookmarked = false
}

func (p *LabeledSolver2) Reset() {
	p.solver.Reset()
	clear(p.labels)
	p.bookmarked = false
	p.bookmarkID = 0
}
