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

func TestSourceArrayAccessorsExposeAllocatedStorage(t *testing.T) {
	p := Msat_SolverAlloc(4, 1, 1, 1, 1, 0)
	Msat_SolverAddVar(p, 0)
	if len(Msat_SolverReadAssignsArray(p)) != 4 || len(Msat_SolverReadModelArray(p)) != 4 || len(Msat_SolverReadFactors(p)) != 4 || len(Msat_SolverReadAdjacents(p)) != 4 {
		t.Fatal("C exposes allocated arrays, not only active variables")
	}
}
func TestSourceStatisticAccessorsCastToCInt(t *testing.T) {
	p := New(1)
	p.Stats.nConflicts, p.Stats.nInspects = 1<<31, 1<<31
	if got := Msat_SolverReadBackTracks(p); got != -1<<31 {
		t.Fatalf("backtracks cast = %d", got)
	}
	if got := Msat_SolverReadInspects(p); got != -1<<31 {
		t.Fatalf("inspects cast = %d", got)
	}
}
