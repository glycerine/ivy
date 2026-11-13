package bsat2

import "testing"

func TestReduceDBKeepsLockedAndBinaryClausesThroughCollection(t *testing.T) {
	s := NewSolver()
	for i := 0; i < 8; i++ {
		s.NewVariable()
	}
	add := func(lits []Lit, activity float32) CRef {
		h := s.ca.Alloc(lits, true)
		s.ca.At(h).Activity = activity
		s.learnts = append(s.learnts, h)
		s.attach(h)
		return h
	}
	locked := add([]Lit{0, 2, 4}, 0)
	add([]Lit{2, 4, 6}, 1)
	add([]Lit{4, 6, 8}, 2)
	add([]Lit{6, 8, 10}, 100)
	add([]Lit{12, 14}, 0)
	s.trailLimits = append(s.trailLimits, 0)
	s.uncheckedEnqueue(0, locked)
	s.GarbageFraction = 0 // Force relocation of retained clauses and reasons.
	s.reduceDB()
	if len(s.learnts) != 4 {
		t.Fatalf("retained %d clauses, want 4", len(s.learnts))
	}
	reason := s.vardata[0].reason
	if reason == CRefUndef || !s.locked(reason) || len(s.ca.At(reason).Lits) != 3 {
		t.Fatal("locked reason was not preserved through relocation")
	}
	binary := 0
	for _, h := range s.learnts {
		c := s.ca.At(h)
		if c.Mark != 0 {
			t.Fatal("deleted clause remained in database")
		}
		if len(c.Lits) == 2 {
			binary++
		}
		if c.Activity == 1 {
			t.Fatal("low activity unlocked clause survived")
		}
	}
	if binary != 1 {
		t.Fatalf("retained %d binary clauses", binary)
	}
	s.cancelUntil(0)
	if s.Solve() != True {
		t.Fatal("solver unusable after reduction and collection")
	}
}
