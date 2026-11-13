package msat

import "testing"

func TestClausePropagateKeepsOutputWhenSatisfied(t *testing.T) {
	c := &Msat_Clause_t{lits: []int{0, 2}}
	out := 42
	if Msat_ClausePropagate(c, 1, []int{1, 2}, &out) != 1 || out != 42 {
		t.Fatalf("C leaves output untouched when satisfied: %d", out)
	}
}

func TestClausePropagateRequiresMatchingWatch(t *testing.T) {
	c := &Msat_Clause_t{lits: []int{0, 2}}
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts when the literal does not match a watch")
		}
	}()
	out := -1
	Msat_ClausePropagate(c, 4, []int{-1, -1, -1}, &out)
}
