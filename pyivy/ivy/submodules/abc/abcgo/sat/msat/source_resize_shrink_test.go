package msat

import "testing"

func TestSourceResizeShrinksAllocationAndCanRegrow(t *testing.T) {
	p := New(4)
	Msat_SolverClean(p, 2)
	p.pFactors[0], p.pFactors[1] = 2, 3
	Msat_SolverResize(p, 2)
	if got := Msat_SolverReadVarAllocNum(p); got != 2 {
		t.Fatalf("after resize to 2, allocated variables = %d, want 2", got)
	}
	if len(p.pvWatched) != 4 || len(p.vVarsUsed.data) != 2 {
		t.Fatalf("watch and used sizes = %d, %d, want 4, 2", len(p.pvWatched), len(p.vVarsUsed.data))
	}
	Msat_SolverResize(p, 3)
	if Msat_SolverReadVarAllocNum(p) != 3 || p.pFactors[0] != 2 || p.pFactors[1] != 3 || p.pFactors[2] != 1 {
		t.Fatalf("regrow did not preserve existing factors and initialize the new factor: %v", p.pFactors)
	}
	if p.pAssigns[2] != -1 || p.pLevel[2] != -1 || p.pReasons[2] != nil || p.pSeen[2] != 0 {
		t.Fatal("regrown variable must have freshly initialized state")
	}
}
