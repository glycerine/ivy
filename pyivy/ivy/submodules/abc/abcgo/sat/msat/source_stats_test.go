package msat

import (
	"math"
	"testing"
)

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

func TestRemoveMarkedResetsClauseNumberCounter(t *testing.T) {
	p := New(4)
	if !p.AddClause([]int{0, 2}) {
		t.Fatal("base clause")
	}
	Msat_SolverMarkClausesStart(p)
	if !p.AddClause([]int{4, 6}) {
		t.Fatal("later clause")
	}
	Msat_SolverRemoveMarked(p)
	if got := Msat_SolverReadClauseNum(p); got != 1 {
		t.Fatalf("clause counter after removal = %d", got)
	}
}

func TestSourceProgressEstimateWithZeroVariablesIsNaN(t *testing.T) {
	p := New(0)
	if got := Msat_SolverProgressEstimate(p); !math.IsNaN(got) {
		t.Fatalf("C computes 0/0 for zero variables, got %v", got)
	}
}
