package minisatp

import "testing"

func TestSimplifyClearsWatchListsForRootAssignments(t *testing.T) {
	s := newSolver(2)
	if !s.addClause([]int32{0, 2}) || !s.addClause([]int32{1}) {
		t.Fatal("setup")
	}
	s.simplifyDB()
	if !s.ok || s.assign[1] != 1 || len(s.clauses) != 1 || s.reason[1] != s.clauses[0] {
		t.Fatal("expected retained locked clause")
	}
	for _, lit := range []int32{0, 1, 2, 3} {
		if len(s.watches[lit]) != 0 {
			t.Fatalf("source clears assigned-variable watch list %d", lit)
		}
	}
}
