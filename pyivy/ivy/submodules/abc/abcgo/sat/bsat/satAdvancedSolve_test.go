package bsat

import (
	"math/rand"
	"testing"
)

func TestLexSatAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(819))
	for trial := 0; trial < 80; trial++ {
		s := sat_solver_new()
		sat_solver_setnvars(s, 3)
		var formula [][]lit
		consistent := true
		for k := 0; k < 5; k++ {
			permutation := rng.Perm(3)
			c := []lit{toLitCond(int32(permutation[0]), int32(rng.Intn(2))), toLitCond(int32(permutation[1]), int32(rng.Intn(2)))}
			formula = append(formula, c)
			if sat_solver_addclause(s, c) == 0 {
				consistent = false
			}
		}
		if !consistent {
			continue
		}
		best, bestKey := -1, 9
		for model := 0; model < 8; model++ {
			valid := true
			for _, c := range formula {
				if !modelSatisfiesClause(model, c) {
					valid = false
				}
			}
			key := (model&1)*4 + ((model>>1)&1)*2 + ((model >> 2) & 1)
			if valid && key < bestKey {
				best, bestKey = model, key
			}
		}
		literals := []lit{1, 3, 5}
		status := sat_solver_solve_lexsat(s, literals, 3)
		if best < 0 {
			if status != l_False {
				t.Fatal("expected unsat")
			}
			continue
		}
		if status != l_True {
			t.Fatal(trial, status, formula)
		}
		for v, l := range literals {
			expected := toLitCond(int32(v), int32(1-((best>>uint(v))&1)))
			if l != expected {
				t.Fatalf("trial %d: got %v want model %d formula %v", trial, literals, best, formula)
			}
		}
		if s.root_level != 0 || s.trail_lim.size != 0 {
			t.Fatal("lex assumptions not restored")
		}
	}
}
func TestAssumptionMinimization(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 3)
	sat_solver_addclause(s, []lit{0, 2})
	literals := []lit{1, 3, 5}
	if sat_solver_solve(s, literals, 0, 0, 0, 0) != l_False {
		t.Fatal("precondition")
	}
	n := sat_solver_minimize_assumptions(s, literals, 3, 0)
	if n != 2 || sat_solver_solve(s, literals[:n], 0, 0, 0, 0) != l_False {
		t.Fatal(n, literals)
	}
	for i := int32(0); i < n; i++ {
		other := append([]lit(nil), literals[:i]...)
		other = append(other, literals[i+1:n]...)
		if sat_solver_solve(s, other, 0, 0, 0, 0) != l_True {
			t.Fatal("not minimal", literals[:n])
		}
	}
	if s.root_level != 0 {
		t.Fatal("assumptions left pushed")
	}
}
func TestSolverVariableRollback(t *testing.T) {
	s := sat_solver_new()
	sat_solver_addclause(s, []lit{0, 2})
	sat_solver_simplify(s)
	sat_solver_bookmark(s)
	sat_solver_addclause(s, []lit{1, 4})
	sat_solver_addclause(s, []lit{5})
	if sat_solver_propagate(s) != 0 {
		t.Fatal("temporary formula")
	}
	sat_solver_rollback(s)
	if s.size != 2 || s.qtail != 0 {
		t.Fatal("rollback state")
	}
	if sat_solver_solve(s, []lit{0}, 0, 0, 0, 0) != l_True {
		t.Fatal("temporary implication survived")
	}
	if s.order.size > 2 {
		t.Fatal("heap contains removed variables")
	}
}
func TestSpecializedAssumptionMinimizationBase(t *testing.T) {
	s := sat_solver_new()
	for _, c := range [][]lit{{0, 2}, {1, 2}, {0, 3}, {1, 3}} {
		sat_solver_addclause(s, c)
	}
	sat_solver_setnvars(s, 3)
	literals := []lit{4}
	if sat_solver_minimize_assumptions2(s, literals, 1, 0) != 0 || s.root_level != 0 {
		t.Fatal("redundant specialized literal")
	}
}
