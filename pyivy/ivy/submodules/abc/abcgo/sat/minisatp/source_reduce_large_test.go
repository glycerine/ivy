package minisatp

import "testing"

// Upstream Sort.h quicksort partitions 16 equal entries by swapping ends;
// its two 8-entry selection sorts then leave those ties in reversed order.
func TestReduceDBLargeEqualActivitiesUsesSourcePartitionOrder(t *testing.T) {
	s := newSolver(3)
	var original []*clause
	for i := 0; i < 16; i++ {
		c := &clause{id: ClauseID(i), learnt: true, lits: []int32{0, 2, 4}, activity: 2}
		original = append(original, c)
		s.learnts = append(s.learnts, c)
		s.watches[1] = append(s.watches[1], c)
		s.watches[3] = append(s.watches[3], c)
	}
	s.reduceDB()
	if len(s.learnts) != 8 {
		t.Fatalf("remaining clauses = %d", len(s.learnts))
	}
	for i, c := range s.learnts {
		if c != original[7-i] {
			t.Fatalf("remaining order at %d is clause %d, want %d", i, c.id, original[7-i].id)
		}
	}
}
