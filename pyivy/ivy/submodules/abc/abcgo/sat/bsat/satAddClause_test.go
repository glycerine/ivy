package bsat

import "testing"

func TestSolverInputNormalization(t *testing.T) {
	s := sat_solver_new()
	if sat_solver_addclause(s, []lit{4, 0, 4}) != 1 || s.size != 3 || s.stats.clauses != 1 {
		t.Fatal("duplicates")
	}
	if sat_solver_addclause(s, []lit{2, 3}) != 1 || s.stats.clauses != 1 {
		t.Fatal("tautology")
	}
	if sat_solver_addclause(s, []lit{1}) != 1 || sat_solver_propagate(s) != 0 || s.assigns[2] != 0 {
		t.Fatal("root implication")
	}
	if sat_solver_addclause(s, []lit{5}) != 0 {
		t.Fatal("conflicting unit")
	}
	if sat_solver_addclause(s, nil) != 0 {
		t.Fatal("empty clause")
	}
}
func TestSolverProofStoreRecording(t *testing.T) {
	s := sat_solver_new()
	sat_solver_store_alloc(s)
	sat_solver_addclause(s, []lit{0, 2})
	sat_solver_store_mark_clauses_a(s)
	sat_solver_addclause(s, []lit{1, 4})
	sat_solver_store_mark_roots(s)
	p := solverClauseStore(s)
	if p.nClauses != 2 || p.nRoots != 2 || p.nClausesA != 1 {
		t.Fatal("proof roots")
	}
	released := sat_solver_store_release(s)
	if released != p || s.pStore != nil {
		t.Fatal("ownership release")
	}
	Sto_ManFree(p)
}

func TestClauseNormalizationReusesPrivateStorage(t *testing.T) {
	s := NewSolver()
	s.SetNumVars(65)
	input := make([]Literal, 0, 66)
	for i := int32(63); i >= 0; i-- {
		input = append(input, Negative(i))
	}
	input = append(input, Positive(1), Negative(0))
	want := append([]Literal(nil), input...)
	allocs := testing.AllocsPerRun(100, func() {
		if !s.AddClause(input...) {
			t.Fatal("rejected tautological input")
		}
	})
	if allocs != 0 {
		t.Errorf("normalizing a repeated 66-literal clause allocated %.0f objects", allocs)
	}
	for i, l := range input {
		if l != want[i] {
			t.Fatal("normalizing changed caller literals")
		}
	}
}
