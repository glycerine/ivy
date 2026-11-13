package msat

import "testing"

func TestClauseSimplifyRemovesFalseUnwatchedLiterals(t *testing.T) {
	c := &Msat_Clause_t{lits: []int{0, 2, 4}, learned: true, activity: 3}
	if Msat_ClauseSimplify(c, []int{-1, -1, 5}) != 0 {
		t.Fatal("not satisfied")
	}
	if len(c.lits) != 2 || c.lits[0] != 0 || c.lits[1] != 2 || c.activity != 3 {
		t.Fatalf("C compacts unwatched false literal: %+v", c)
	}
}

func TestClauseSimplifyAssertsFalseWatch(t *testing.T) {
	c := &Msat_Clause_t{lits: []int{0, 2, 4}}
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts when a watched literal is false")
		}
	}()
	Msat_ClauseSimplify(c, []int{1, -1, -1})
}
