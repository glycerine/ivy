package msat

import "testing"

func TestSimplifyResetsClauseNumberCounter(t *testing.T) {
	p := New(3)
	if !p.AddClause([]int{0, 2}) || !p.AddClause([]int{0, 4}) || !p.AddClause([]int{0}) {
		t.Fatal("setup")
	}
	if Msat_SolverSimplifyDB(p) == 0 {
		t.Fatal("simplify")
	}
	if got := Msat_SolverReadClauseNum(p); got != 0 {
		t.Fatalf("clause counter after simplification = %d", got)
	}
	if !p.AddClause([]int{3, 4}) {
		t.Fatal("new clause")
	}
	if got := p.vClauses.data[0].Num; got != 0 {
		t.Fatalf("new clause number = %d", got)
	}
}
