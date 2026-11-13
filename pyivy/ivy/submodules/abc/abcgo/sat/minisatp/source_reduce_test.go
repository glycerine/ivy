package minisatp

import "testing"

func TestReduceDBPreservesSourceTieOrdering(t *testing.T) {
	s := newSolver(3)
	a := &clause{id: 0, learnt: true, lits: []int32{0, 2, 4}, activity: 2}
	b := &clause{id: 1, learnt: true, lits: []int32{0, 2, 4}, activity: 2}
	c := &clause{id: 2, learnt: true, lits: []int32{0, 2, 4}, activity: 1}
	for _, cl := range []*clause{a, b, c} {
		s.learnts = append(s.learnts, cl)
		for _, lit := range cl.lits[:2] {
			s.watches[lit^1] = append(s.watches[lit^1], cl)
		}
	}
	s.reduceDB()
	if len(s.learnts) != 2 || s.learnts[0] != b || s.learnts[1] != a {
		t.Fatalf("upstream selection sort retains tied clauses in B,A order, got %v", s.learnts)
	}
}
