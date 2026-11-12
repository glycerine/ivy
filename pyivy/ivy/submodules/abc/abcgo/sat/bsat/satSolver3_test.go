package bsat

import (
	"math"
	"math/rand"
	"testing"
)

func TestSolver3AgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(919))
	for trial := 0; trial < 150; trial++ {
		s := sat_solver3_new()
		sat_solver3_setnvars(s, 4)
		var formula [][]lit
		consistent := true
		for k := 0; k < 10; k++ {
			perm := rng.Perm(4)
			c := []lit{toLitCond(int32(perm[0]), int32(rng.Intn(2))), toLitCond(int32(perm[1]), int32(rng.Intn(2)))}
			formula = append(formula, c)
			if consistent && sat_solver3_addclause(s, c) == 0 {
				consistent = false
			}
		}
		status := int32(l_False)
		if consistent {
			status = sat_solver3_solve(s, nil, 0, 0, 0, 0)
		}
		if (status == l_True) != formulaHasModel(formula, nil, 4) {
			t.Fatal(trial, status, formula)
		}
		sat_solver3_delete(s)
	}
}
func TestSolver3MinimizerAndLimitRestore(t *testing.T) {
	s := sat_solver3_new()
	sat_solver3_setnvars(s, 3)
	sat_solver3_addclause(s, []lit{0, 2})
	literals := []lit{1, 3, 5}
	s.nConfLimit = 1 << 42
	n := sat_solver3_minimize_assumptions(s, literals, 3, 0)
	if n != 2 || s.nConfLimit != 1<<42 || sat_solver3_solve(s, literals[:n], 0, 0, 0, 0) != l_False {
		t.Fatal(n, literals, s.nConfLimit)
	}
	if s.root_level != 0 {
		t.Fatal("assumptions retained")
	}
}
func TestSolver3DoublePriority(t *testing.T) {
	s := sat_solver3_new()
	sat_solver3_setnvars(s, 3)
	s.VarActType = 1
	sat_solver3_set_var_activity(s, []int32{2, 0}, 2)
	if math.Float64frombits(s.activity[2]) != 2 || math.Float64frombits(s.activity[0]) != 1 || order_select(s, 0) != 2 {
		t.Fatal("solver3 priorities")
	}
}
