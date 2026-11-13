package msat

import "testing"

func TestPreparePreservesRootUnitAcrossQueries(t *testing.T) {
	p := Msat_SolverAlloc(4, 1, 1, 1, 1, 0)
	Msat_SolverAddVar(p, 0)
	Msat_SolverAddVar(p, 0)
	if Msat_SolverAddClause(p, &Msat_IntVec_t{data: []int{0}}) == 0 {
		t.Fatal("unit rejected")
	}
	vars := &Msat_IntVec_t{data: []int{0, 1}}
	for i := 0; i < 2; i++ {
		Msat_SolverPrepare(p, vars)
		if got := Msat_SolverSolve(p, &Msat_IntVec_t{data: []int{1}}, -1, 0); got != int(MSAT_FALSE) {
			t.Fatalf("query %d forgot unit: %d", i, got)
		}
	}
}

func TestPrepareDiscardsAssumptionsButKeepsRootFacts(t *testing.T) {
	p := Msat_SolverAlloc(4, 1, 1, 1, 1, 0)
	for i := 0; i < 3; i++ {
		Msat_SolverAddVar(p, 0)
	}
	if Msat_SolverAddClause(p, &Msat_IntVec_t{data: []int{0}}) == 0 {
		t.Fatal("unit rejected")
	}
	vars := &Msat_IntVec_t{data: []int{0, 1, 2}}
	Msat_SolverPrepare(p, vars)
	if got := Msat_SolverSolve(p, &Msat_IntVec_t{data: []int{2}}, -1, 0); got != int(MSAT_TRUE) {
		t.Fatal("first assumption", got)
	}
	Msat_SolverPrepare(p, vars)
	if got := Msat_SolverSolve(p, &Msat_IntVec_t{data: []int{3}}, -1, 0); got != int(MSAT_TRUE) {
		t.Fatal("old assumption leaked", got)
	}
	if got := Msat_SolverSolve(p, &Msat_IntVec_t{data: []int{1}}, -1, 0); got != int(MSAT_FALSE) {
		t.Fatal("root unit lost", got)
	}
}
