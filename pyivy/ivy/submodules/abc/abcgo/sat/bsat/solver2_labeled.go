package bsat

import (
	"slices"
	"time"
)

// LabeledSolver2 associates each proof root with the caller's problem-object
// ID. Multiple clauses may share an ID; -1 marks a background clause that is
// omitted from the reported core. Resolution proofs remain in Solver2's graph.
type LabeledSolver2 struct {
	solver      *Solver2
	labels      map[int32]int32
	bookmarkID  int32
	bookmarked  bool
	sourceMode  bool
	incremental bool
	proofIDs    []int32
	proofCount  int32
}

func NewLabeledSolver2() *LabeledSolver2 {
	return &LabeledSolver2{solver: NewSolver2(), labels: make(map[int32]int32)}
}

// NewLabeledSolver2ProofMode uses the two proof modes of ABC's old GLA.
// In incremental mode, proof logging starts only when explicitly requested.
func NewLabeledSolver2ProofMode(full bool) *LabeledSolver2 {
	if full {
		p := NewLabeledSolver2()
		p.sourceMode = true
		return p
	}
	return &LabeledSolver2{
		solver: &Solver2{solver: sat_solver2_new(), nextID: 1},
		labels: make(map[int32]int32), sourceMode: true, incremental: true,
	}
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
	var result bool
	if p.incremental {
		result = p.solver.addClause(id, literals)
	} else {
		result = p.solver.AddClause(literals...)
	}
	if !p.incremental && p.solver.nextID != root {
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
func (p *LabeledSolver2) SetRuntimeDeadline(clock func() int64, deadline int64) {
	p.solver.SetRuntimeDeadline(clock, deadline)
}

// ProofCore returns caller IDs that contributed to the UNSAT result. After a
// bookmark, only clauses added since that bookmark are reported; earlier
// clauses form the fixed background, as in ABC's restarted bitset logger.
func (p *LabeledSolver2) ProofCore() []int32 {
	if p.incremental {
		return p.solver.ProofCore()
	}
	seen := make(map[int32]bool)
	var result []int32
	for _, root := range p.solver.ProofCore() {
		if !p.sourceMode && p.bookmarked && root < p.bookmarkID {
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

// StartIncrementalProof begins ABC's pPrf2 logger after the first bookmark.
func (p *LabeledSolver2) StartIncrementalProof(objectCount, width int32) {
	if !p.incremental || p.solver.solver.pPrf2 != nil || objectCount < 0 {
		panic("invalid incremental proof start")
	}
	p.proofIDs = make([]int32, objectCount)
	for i := range p.proofIDs {
		p.proofIDs[i] = -1
	}
	p.proofCount = 0
	m := Prf_ManAlloc()
	Prf_ManRestart(m, &p.proofIDs, int32(p.solver.solver.stats.learnts), width)
	p.solver.solver.pPrf2 = m
}

func (p *LabeledSolver2) GrowIncrementalProof(additional int32) {
	if p.solver.solver.pPrf2 == nil {
		panic("incremental proof is not active")
	}
	Prf_ManGrow(p.solver.solver.pPrf2, p.proofCount+additional)
}

func (p *LabeledSolver2) RegisterIncrementalProofID(id int32) {
	if p.solver.solver.pPrf2 == nil || id < 0 || int(id) >= len(p.proofIDs) || p.proofIDs[id] != -1 {
		panic("invalid incremental proof object")
	}
	p.proofIDs[id] = p.proofCount
	p.proofCount++
}

func (p *LabeledSolver2) IncrementalProofActive() bool { return p.solver.solver.pPrf2 != nil }
func (p *LabeledSolver2) HasProofManager() bool {
	return p.solver.solver.pPrf1 != nil || p.solver.solver.pPrf2 != nil
}
func (p *LabeledSolver2) StopIncrementalProof() {
	Prf_ManStopP(&p.solver.solver.pPrf2)
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
	if p.incremental {
		p.solver.solver.pPrf1 = nil
	}
	p.proofIDs = nil
	p.proofCount = 0
	clear(p.labels)
	p.bookmarked = false
	p.bookmarkID = 0
}
