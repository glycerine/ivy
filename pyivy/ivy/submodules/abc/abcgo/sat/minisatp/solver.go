package minisatp

import (
	"sort"
	"time"
)

// These types and methods follow Solver.C in MiniSat-p 1.14. The solver is
// intentionally private until its search and proof export are complete.
type clause struct {
	lits     []int32
	id       ClauseID
	learnt   bool
	activity float64
}

type solver struct {
	ok            bool
	proof         proof
	clauses       []*clause
	watches       [][]*clause
	assign        []int8 // -1 false, 0 undefined, 1 true
	model         []int8
	level         []int
	reason        []*clause
	unitID        []ClauseID
	trail         []int32
	limits        []int
	qhead         int
	trailPos      []int
	seen          []bool
	activity      []float64
	varInc        float64
	claInc        float64
	order         varOrder
	learnts       []*clause
	conflicts     int32
	inputCount    int
	rootInputs    []int
	simpAssigns   int
	simpProps     int64
	deadline      time.Time
	conflictLimit int32
	interrupted   bool
}

func (s *solver) expired() bool {
	if s.interrupted || !s.deadline.IsZero() && !time.Now().Before(s.deadline) {
		s.interrupted = true
		return true
	}
	return false
}

func newSolver(n int) *solver {
	s := &solver{ok: true, watches: make([][]*clause, 2*n), assign: make([]int8, n), level: make([]int, n), reason: make([]*clause, n), unitID: make([]ClauseID, n), trailPos: make([]int, n), seen: make([]bool, n), activity: make([]float64, n), varInc: 1, claInc: 1}
	for i := range s.unitID {
		s.unitID[i] = -1
	}
	s.order = newVarOrder(s)
	return s
}

func (s *solver) value(lit int32) int8 {
	v := s.assign[lit>>1]
	if lit&1 != 0 {
		return -v
	}
	return v
}

func (s *solver) enqueue(lit int32, from *clause) bool {
	if v := s.value(lit); v != 0 {
		return v > 0
	}
	v := lit >> 1
	s.assign[v] = 1
	if lit&1 != 0 {
		s.assign[v] = -1
	}
	s.level[v] = len(s.limits)
	s.reason[v] = from
	s.trailPos[v] = len(s.trail)
	s.trail = append(s.trail, lit)
	return true
}

// addClause follows newClause's original-clause normalization and root-level
// resolution. It reports whether the solver remains consistent.
func (s *solver) addClause(input []int32) bool {
	if !s.ok || s.expired() {
		return false
	}
	inputIndex := s.inputCount
	s.inputCount++
	q := append([]int32(nil), input...)
	sort.Slice(q, func(i, j int) bool { return q[i] < q[j] })
	n := 0
	for _, lit := range q {
		if n == 0 || q[n-1] != lit {
			q[n] = lit
			n++
		}
	}
	q = q[:n]
	for i, lit := range q {
		if i+1 < len(q) && lit^1 == q[i+1] || s.value(lit) > 0 {
			return true
		}
	}
	s.rootInputs = append(s.rootInputs, inputIndex)
	s.proof.beginChain(s.proof.addRoot(q))
	kept := q[:0]
	for _, lit := range q {
		if s.value(lit) < 0 {
			s.proof.resolve(s.unitID[lit>>1], lit>>1)
		} else {
			kept = append(kept, lit)
		}
	}
	id := s.proof.endChain()
	switch len(kept) {
	case 0:
		s.ok = false
	case 1:
		s.unitID[kept[0]>>1] = id
		s.ok = s.enqueue(kept[0], nil)
	default:
		c := &clause{lits: append([]int32(nil), kept...), id: id}
		s.clauses = append(s.clauses, c)
		s.watches[c.lits[0]^1] = append(s.watches[c.lits[0]^1], c)
		s.watches[c.lits[1]^1] = append(s.watches[c.lits[1]^1], c)
	}
	return s.ok
}

// propagate follows MiniSat's two-watched-literal propagation and records
// root-level unit derivations and the final empty clause.
func (s *solver) propagate() *clause {
	for s.qhead < len(s.trail) {
		if s.expired() {
			return nil
		}
		p := s.trail[s.qhead]
		s.qhead++
		s.simpProps--
		watches := s.watches[p]
		kept := watches[:0]
		for i, c := range watches {
			if s.expired() {
				kept = append(kept, watches[i:]...)
				s.watches[p] = kept
				return nil
			}
			falseLit := p ^ 1
			if c.lits[0] == falseLit {
				c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
			}
			first := c.lits[0]
			if s.value(first) > 0 {
				kept = append(kept, c)
				continue
			}
			moved := false
			for k := 2; k < len(c.lits); k++ {
				if s.value(c.lits[k]) >= 0 {
					c.lits[1], c.lits[k] = c.lits[k], falseLit
					s.watches[c.lits[1]^1] = append(s.watches[c.lits[1]^1], c)
					moved = true
					break
				}
			}
			if moved {
				continue
			}
			if len(s.limits) == 0 {
				s.proof.beginChain(c.id)
				for _, lit := range c.lits[1:] {
					s.proof.resolve(s.unitID[lit>>1], lit>>1)
				}
				id := s.proof.endChain()
				if s.value(first) >= 0 {
					s.unitID[first>>1] = id
				} else {
					s.proof.beginChain(s.unitID[first>>1])
					s.proof.resolve(id, first>>1)
					s.proof.endChain()
				}
			}
			kept = append(kept, c)
			if !s.enqueue(first, c) {
				if len(s.limits) == 0 {
					s.ok = false
				}
				s.qhead = len(s.trail)
				kept = append(kept, watches[i+1:]...)
				s.watches[p] = kept
				return c
			}
		}
		s.watches[p] = kept
	}
	return nil
}
