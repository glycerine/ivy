package if_

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestDsdMaximumSupportedManagerFitsCaches(t *testing.T) {
	p := If_DsdManAlloc(16, 0)
	defer If_DsdManFree(p, 0)
	if If_DsdManVarNum(p) != 16 || If_DsdManObjNum(p) != 2 {
		t.Fatal("maximum-support manager geometry changed")
	}
}

func TestDsdFreeReleasesOwnedSATInstance(t *testing.T) {
	p := If_DsdManAlloc(4, 3)
	solver, ok := p.pSat.(*bsat.Solver)
	if !ok {
		t.Fatalf("missing original owned SAT solver: %T", p.pSat)
	}
	If_DsdManFree(p, 0)
	if solver.NumVars() != 0 || p.pSat != nil {
		t.Fatalf("freed DSD retained SAT state: %d vars, handle%T", solver.NumVars(), p.pSat)
	}
}
