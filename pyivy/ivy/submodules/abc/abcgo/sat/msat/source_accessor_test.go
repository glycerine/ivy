package msat

import "testing"

func TestReadVarValueReturnsEncodedAssignment(t *testing.T) {
	p := New(3)
	if got := Msat_SolverReadVarValue(p, 2); int(got) != MSAT_VAR_UNASSIGNED {
		t.Fatalf("unassigned = %d", got)
	}
	Msat_SolverEnqueue(p, 4, nil)
	if got := Msat_SolverReadVarValue(p, 2); int(got) != 4 {
		t.Fatalf("C returns encoded literal 4, got %d", got)
	}
}
