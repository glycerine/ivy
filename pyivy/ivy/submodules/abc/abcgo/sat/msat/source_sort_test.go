package msat

import "testing"

func TestLearnedClauseSortMatchesSourceSelectionTieOrder(t *testing.T) {
	p := New(1)
	a, b, c := &Msat_Clause_t{activity: 2}, &Msat_Clause_t{activity: 2}, &Msat_Clause_t{activity: 1}
	p.vLearned.data = append(p.vLearned.data, a, b, c)
	Msat_SolverSortDB(p)
	if p.vLearned.data[0] != c || p.vLearned.data[1] != b || p.vLearned.data[2] != a {
		t.Fatal("C selection sort swaps the minimum with the first entry")
	}
}
func TestLearnedClauseSortMatchesSourceQuickSortTieOrder(t *testing.T) {
	p := New(1)
	var original []*Msat_Clause_t
	for i := 0; i < 16; i++ {
		c := &Msat_Clause_t{activity: 1}
		original = append(original, c)
		p.vLearned.data = append(p.vLearned.data, c)
	}
	Msat_SolverSortDB(p)
	for i := 0; i < 16; i++ {
		if p.vLearned.data[i] != original[15-i] {
			t.Fatalf("C partition reverses equal entries: index %d", i)
		}
	}
}
