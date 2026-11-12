package bsat

import (
	"math/rand"
	"testing"
	"time"
)

func formulaHasModel(formula [][]lit, assumptions []lit, n int) bool {
	for model := 0; model < 1<<uint(n); model++ {
		valid := true
		for _, l := range assumptions {
			if !modelSatisfiesClause(model, []lit{l}) {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		for _, c := range formula {
			if !modelSatisfiesClause(model, c) {
				valid = false
				break
			}
		}
		if valid {
			return true
		}
	}
	return false
}
func TestSolverEndToEndAgainstExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(52819))
	for trial := 0; trial < 500; trial++ {
		s := sat_solver_new()
		sat_solver_setnvars(s, 5)
		if trial%3 == 0 {
			s.nLearntStart = 2
			s.nLearntMax = 2
			s.nLearntDelta = 0
		}
		var formula [][]lit
		inputConsistent := true
		for k := 0; k < 15; k++ {
			permutation := rng.Perm(5)
			c := make([]lit, 1+rng.Intn(4))
			for i := range c {
				c[i] = toLitCond(int32(permutation[i]), int32(rng.Intn(2)))
			}
			formula = append(formula, c)
			if inputConsistent && sat_solver_addclause(s, c) == 0 {
				inputConsistent = false
			}
		}
		assumptions := make([]lit, rng.Intn(3))
		for i := range assumptions {
			assumptions[i] = toLitCond(int32(rng.Intn(5)), int32(rng.Intn(2)))
		}
		wantSAT := formulaHasModel(formula, assumptions, 5)
		status := int32(l_False)
		if inputConsistent {
			status = sat_solver_solve(s, assumptions, 0, 0, 0, 0)
		}
		if (status == l_True) != wantSAT || status == l_Undef {
			t.Fatalf("trial %d: status %d want SAT=%v formula %v assumptions %v", trial, status, wantSAT, formula, assumptions)
		}
		if status == l_True {
			model := 0
			for v := int32(0); v < 5; v++ {
				if sat_solver_var_value(s, v) != 0 {
					model |= 1 << uint(v)
				}
			}
			for _, c := range formula {
				if !modelSatisfiesClause(model, c) {
					t.Fatal("invalid model", trial, s.model, c)
				}
			}
			for _, l := range assumptions {
				if !modelSatisfiesClause(model, []lit{l}) {
					t.Fatal("model violates assumption")
				}
			}
		}
		sat_solver_delete(s)
	}
}
func TestSolverIncrementalAssumptions(t *testing.T) {
	s := sat_solver_new()
	sat_solver_addclause(s, []lit{toLit(0), toLit(1)})
	if sat_solver_solve(s, []lit{lit_neg(toLit(0)), lit_neg(toLit(1))}, 0, 0, 0, 0) != l_False {
		t.Fatal("inconsistent assumptions")
	}
	if s.root_level != 0 || s.trail_lim.size != 0 || s.conf_final.size == 0 {
		t.Fatal("assumption cleanup/core")
	}
	if sat_solver_solve(s, []lit{toLit(0)}, 0, 0, 0, 0) != l_True || sat_solver_var_value(s, 0) != 1 {
		t.Fatal("incremental SAT")
	}
	if sat_solver_addclause(s, []lit{lit_neg(toLit(0))}) == 0 || sat_solver_solve(s, nil, 0, 0, 0, 0) != l_True || sat_solver_var_value(s, 1) != 1 {
		t.Fatal("incremental clause")
	}
}
func TestSolverSearchBudgetAndDeadline(t *testing.T) {
	s := sat_solver_new()
	sat_solver_addclause(s, []lit{toLit(0), toLit(1)})
	sat_solver_addclause(s, []lit{toLit(0), lit_neg(toLit(1))})
	if status := sat_solver_search(s, 1); status != l_Undef || s.stats.conflicts != 1 {
		t.Fatal("restart boundary", status, s.stats)
	}
	if sat_solver_solve(s, nil, 0, 0, 0, 0) != l_True {
		t.Fatal("resume after restart")
	}
	s.nRuntimeLimit = time.Now().Add(-time.Second).UnixNano()
	if sat_solver_solve(s, nil, 0, 0, 0, 0) != l_Undef {
		t.Fatal("expired deadline")
	}
}
func TestSolverLubySequence(t *testing.T) {
	want := []float64{1, 1, 2, 1, 1, 2, 4, 1, 1, 2, 1, 1, 2, 4, 8}
	for i, v := range want {
		if got := luby(2, int32(i)); got != v {
			t.Fatal(i, got, v)
		}
	}
}
func TestSolverUnsatProofStore(t *testing.T) {
	s := sat_solver_new()
	sat_solver_store_alloc(s)
	for _, c := range [][]lit{{0, 2}, {1, 2}, {0, 3}, {1, 3}} {
		if sat_solver_addclause(s, c) == 0 {
			t.Fatal("premature conflict")
		}
	}
	sat_solver_store_mark_roots(s)
	if sat_solver_solve(s, nil, 0, 0, 0, 0) != l_False {
		t.Fatal("UNSAT")
	}
	p := solverClauseStore(s)
	if p.pEmpty == nil || p.nRoots != 4 || p.nClauses <= 4 || s.stats.conflicts == 0 {
		t.Fatal("missing proof records", p)
	}
	if s.tagged.size != 0 {
		t.Fatal("stale analysis tags")
	}
}
