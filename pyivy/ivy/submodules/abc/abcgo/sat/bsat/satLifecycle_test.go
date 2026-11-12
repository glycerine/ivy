package bsat

import "testing"

func TestSolverAllocationGrowthRestart(t *testing.T) {
	s := sat_solver_new()
	if s.random_seed != 91648253 || s.Mem.nPageSize != 17 || s.binary.size != 2 {
		t.Fatal("defaults")
	}
	sat_solver_setnvars(s, 4)
	if s.size != 4 || s.order.size != 4 {
		t.Fatal("growth")
	}
	s.satAssignments.enqueue(toLit(0), 0)
	s.activity[2] = 9000
	s.model[1] = 1
	s.factors = make([]float64, 4)
	s.factors[2] = 2
	s.pFreqs[1] = 1
	s.levels[0] = 3
	s.qhead = s.qtail
	s.stats.conflicts = 7
	sat_solver_setnvars(s, 8)
	if s.assigns[0] != 0 || s.activity[2] != 9000 || s.factors[2] != 2 || s.model[1] != 1 || s.levels[0] != 3 {
		t.Fatal("growth lost values")
	}
	for i := 4; i < 8; i++ {
		if s.assigns[i] != satUnassigned || s.orderpos[i] < 0 {
			t.Fatal("new variable initialization")
		}
	}
	oldCap := s.cap
	zsat_solver_restart_seed(s, 99)
	if s.size != 0 || s.qtail != 0 || s.random_seed != 99 || s.stats.conflicts != 0 || s.cap != oldCap {
		t.Fatal("restart")
	}
	sat_solver_setnvars(s, 3)
	if s.assigns[0] != satUnassigned || s.order.size != 3 {
		t.Fatal("restart reuse")
	}
	sat_solver_delete(s)
	if s.Mem.clauses != nil || s.trail != nil || s.wlists != nil {
		t.Fatal("owned buffers retained")
	}
}
func TestSeededSolverAndVariableIDs(t *testing.T) {
	s := zsat_solver_new_seed(12)
	if s.Mem.nPageSize != 15 || s.random_seed != 12 {
		t.Fatal("seeded constructor")
	}
	for i := int32(0); i < 20; i++ {
		if sat_solver_addvar(s) != i {
			t.Fatal(i)
		}
	}
	sat_solver_setnvars(s, 3)
	if s.size != 20 {
		t.Fatal("setnvars shrank variables")
	}
}

func TestEmptySolverDecisionSelection(t *testing.T) {
	s := sat_solver_new()
	if got := order_select(s, 1); got != var_Undef {
		t.Fatal(got)
	}
}
