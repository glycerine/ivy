package msat

import "testing"

func TestAddVarDoesNotAlterConeUntilPrepared(t *testing.T) {
	p := Msat_SolverAlloc(2, 1, 1, 1, 1, 0)
	Msat_SolverAddVar(p, 3)
	if len(p.vConeVars.data) != 0 || len(p.pOrder.heap) != 0 {
		t.Fatalf("C AddVar changes only variable count/level, cone=%v heap=%v", p.vConeVars.data, p.pOrder.heap)
	}
}
