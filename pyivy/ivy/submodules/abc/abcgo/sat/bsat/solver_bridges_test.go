package bsat

import (
	"math"
	"testing"
)

func TestPublicActivityConeFactors(t *testing.T) {
	s := NewSolver()
	v := s.NewVariable()
	s.SetActivityFactor(v, 1)
	s.SetActivityFactor(v, 2)
	if s.act_vars.size != 1 || s.factors[v] != 2 {
		t.Fatal("duplicate activity variable")
	}
	old := s.activity[v]
	if s.Solve() != Satisfiable || s.activity[v] <= old {
		t.Fatal("factor did not reach actual activity bump")
	}
	s.NewVariable()
	if len(s.factors) != 2 || s.factors[v] != 2 {
		t.Fatal("factor growth")
	}
	s.ClearActivityFactors()
	if s.act_vars.size != 0 || s.factors[v] != 0 {
		t.Fatal("cone not cleared")
	}
	s.SetActivityFactor(v, 1)
	s.Reset()
	s.NewVariable()
	s.SetActivityFactor(0, 3)
	if s.act_vars.size != 1 {
		t.Fatal("reset left stale active variables")
	}
}
func TestPublicCumulativeBudgetsAndInspections(t *testing.T) {
	s := NewSolver()
	s.AddClause(0, 2)
	if s.Solve() != Satisfiable {
		t.Fatal("SAT")
	}
	stats := s.Statistics()
	if stats.Propagations == 0 || stats.Inspections == 0 {
		t.Fatal(stats)
	}
	if s.SolveWithGlobalLimits(nil, 0, 1) != Unknown {
		t.Fatal("absolute propagation limit ignored")
	}
	if s.SolveWithGlobalLimits(nil, 0, 0) != Satisfiable {
		t.Fatal("limits leaked into later call")
	}
}
func TestPublicClauseStoreOwnership(t *testing.T) {
	s := NewSolver()
	s.EnableClauseStore()
	for _, c := range [][]lit{{0, 2}, {1, 2}, {0, 3}, {1, 3}} {
		s.AddClause(c...)
	}
	s.MarkClauseStorePartitionA()
	s.MarkClauseStoreRoots()
	if s.Solve() != Unsatisfiable {
		t.Fatal("UNSAT")
	}
	store := s.ReleaseClauseStore()
	if store.NumVariables() != 2 || store.NumRoots() != 4 || store.NumClauses() < 5 {
		t.Fatal("store counts")
	}
	copies := store.Clauses()
	copies[0].Literals[0] = 999
	if store.Clauses()[0].Literals[0] == 999 || s.ReleaseClauseStore() != nil {
		t.Fatal("ownership")
	}
	p := Intp_ManAlloc()
	core := Intp_ManUnsatCore(p, store, 0, 0)
	if Intp_ManError(p) != nil || len(core) != 4 {
		t.Fatal(core, Intp_ManError(p))
	}
}

func TestPublicVariableActivityRepresentations(t *testing.T) {
	s := NewSolver()
	s.NewVariable()
	if s.VariableActivity(0) != 1024 {
		t.Fatal("integer activity")
	}
	s.VarActType = 1
	s.activity[0] = math.Float64bits(1.5)
	if s.VariableActivity(0) != 1.5 {
		t.Fatal("double activity")
	}
	s.VarActType = 2
	s.activity[0] = Xdbl_FromDouble(2.5)
	if s.VariableActivity(0) != 2.5 {
		t.Fatal("extended activity")
	}
}
