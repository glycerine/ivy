package bsat

import "testing"

func TestSolverCoreClausePropagationIntegration(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 4)
	sat_solver_clause_new(s, []lit{lit_neg(toLit(0)), toLit(1)}, 0)
	sat_solver_clause_new(s, []lit{lit_neg(toLit(1)), toLit(2), toLit(3)}, 0)
	if sat_solver_decision(s, toLit(0)) != 1 || sat_solver_propagate(s) != 0 || s.assigns[1] != 0 {
		t.Fatal("binary implication")
	}
	if sat_solver_decision(s, lit_neg(toLit(2))) != 1 || sat_solver_propagate(s) != 0 || s.assigns[3] != 0 {
		t.Fatal("long clause implication")
	}
	sat_solver_canceluntil(s, 0)
	for _, v := range s.assigns {
		if v != satUnassigned {
			t.Fatal("backtrack", s.assigns)
		}
	}
	if s.stats.clauses != 2 || s.stats.clauses_literals != 5 {
		t.Fatal(s.stats)
	}
}
func TestSolverLearnedClauseIdentity(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 3)
	h := sat_solver_clause_new(s, []lit{toLit(0), lit_neg(toLit(1))}, 1)
	if h == 0 || s.hLearnts != h || s.stats.learnts != 1 || s.act_clas.size != 1 || clause_id(Sat_MemClauseHand(&s.Mem, h)) != 0 {
		t.Fatal(h, s)
	}
	sat_solver_enqueue(s, lit_neg(toLit(0)), 0)
	if sat_solver_propagate(s) != 0 || s.assigns[1] != 1 {
		t.Fatal("learned binary watches")
	}
}
