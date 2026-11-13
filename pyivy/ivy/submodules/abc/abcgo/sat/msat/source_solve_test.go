package msat

import "testing"

func TestSourceAddClauseDoesNotStoreContradictionOrClearModel(t *testing.T) {
	p := New(1)
	if Msat_SolverAddClause(p, &Msat_IntVec_t{data: []int{0}}) == 0 {
		t.Fatal("first unit")
	}
	if Msat_SolverSolve(p, nil, 0, 0) != int(MSAT_TRUE) {
		t.Fatal("first solve")
	}
	if Msat_SolverAddClause(p, &Msat_IntVec_t{data: []int{1}}) != 0 {
		t.Fatal("contradictory unit accepted")
	}
	if p.pModel[0] != 0 {
		t.Fatal("C does not clear previous model on AddClause")
	}
	if Msat_SolverSolve(p, nil, 0, 0) != int(MSAT_TRUE) {
		t.Fatal("C requires caller to handle rejected clause")
	}
}
