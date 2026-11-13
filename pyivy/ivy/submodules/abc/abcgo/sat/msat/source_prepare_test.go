package msat

import "testing"

func TestPrepareDoesNotOverwriteStoredConeVector(t *testing.T) {
	p := New(3)
	Msat_SolverPrepare(p, &Msat_IntVec_t{data: []int{1}})
	if got := p.vConeVars.data; len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("C retains cone vector, got %v", got)
	}
	if got := p.pOrder.heap; len(got) != 1 || got[0] != 1 {
		t.Fatalf("requested order = %v", got)
	}
}

func TestSolverAllocPreservesActivityIncrements(t *testing.T) {
	p := Msat_SolverAlloc(2, 0, 1, -2, 1, 0)
	if p.dClaInc != 0 || p.dVarInc != -2 {
		t.Fatalf("C keeps increments: %v %v", p.dClaInc, p.dVarInc)
	}
}

func TestSimplifyDoesNotUseConvenienceFailureFlag(t *testing.T) {
	p := New(1)
	p.okay = false
	if Msat_SolverSimplifyDB(p) != 1 {
		t.Fatal("C simplify checks propagation, not Go convenience state")
	}
}

func TestPrepareDoesNotChangeRootLevelField(t *testing.T) {
	p := New(2)
	p.nLevelRoot = 3
	Msat_SolverPrepare(p, &Msat_IntVec_t{data: []int{0, 1}})
	if p.nLevelRoot != 3 {
		t.Fatalf("C prepare leaves root level unchanged, got %d", p.nLevelRoot)
	}
}
