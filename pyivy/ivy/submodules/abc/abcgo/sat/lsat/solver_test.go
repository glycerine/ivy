package lsat

import (
	"math/rand"
	"testing"
)

func TestLiteralAndTruthConventions(t *testing.T) {
	for v := 0; v < 10; v++ {
		for _, negative := range []bool{false, true} {
			p := MakeLiteral(v, negative)
			if LiteralVariable(p) != v || LiteralSign(p) != negative || Negate(Negate(p)) != p {
				t.Fatal(v, negative, p)
			}
		}
	}
	s := New()
	v := s.NewVariable()
	if s.Value(v) != Undefined || s.LiteralValue(MakeLiteral(v, true)) != Undefined {
		t.Fatal("unassigned value")
	}
	s.AddClause([]int{MakeLiteral(v, false)})
	if s.Value(v) != True || s.LiteralValue(MakeLiteral(v, true)) != False {
		t.Fatal("unit propagation values")
	}
	if !s.Solve(nil) || s.ModelValue(v) != True || s.ModelLiteralValue(MakeLiteral(v, true)) != False {
		t.Fatal("model")
	}
}
func TestClauseAndAssumptionBuilders(t *testing.T) {
	s := solver_new()
	a := solver_newVar(s)
	b := solver_newVar(s)
	solver_addClause_begin(s)
	solver_addClause_addLit(s, solver_mkLit(a))
	solver_addClause_addLit(s, solver_mkLit(b))
	if solver_addClause_commit(s) != 1 {
		t.Fatal("clause builder")
	}
	solver_solve_begin(s)
	solver_solve_addLit(s, solver_mkLit_args(a, 1))
	solver_solve_addLit(s, solver_mkLit_args(b, 1))
	if solver_solve_commit(s) != 0 || solver_conflict_len(s) == 0 {
		t.Fatal("expected assumption conflict")
	}
	solver_solve_begin(s)
	solver_solve_addLit(s, solver_mkLit(a))
	if solver_solve_commit(s) != 1 {
		t.Fatal("assumptions persisted")
	}
	if solver_num_vars(s) != 2 || solver_num_clauses(s) != 1 || solver_modelValue_Var(s, a) != solver_get_l_True() {
		t.Fatal("counts/model")
	}
	if !s.Okay() {
		t.Fatal("assumption UNSAT poisoned formula")
	}
}
func TestDecisionEligibilityDoesNotBlockPropagation(t *testing.T) {
	s := New()
	a, b := s.NewVariable(), s.NewVariable()
	s.SetDecisionVariable(a, false)
	s.AddClause([]int{MakeLiteral(a, false), MakeLiteral(b, false)})
	s.AddClause([]int{MakeLiteral(b, true)})
	if !s.Solve(nil) || s.ModelValue(a) != True || s.ModelValue(b) != False {
		t.Fatal("disabled decision suppressed implication")
	}
	s = New()
	v := s.NewVariable()
	s.SetDecisionVariable(v, false)
	if !s.Solve(nil) || s.ModelValue(v) != Undefined || s.FreeVariableCount() != 0 {
		t.Fatal("disabled unconstrained variable decided")
	}
	s.SetDecisionVariable(v, true)
	s.SetPolarity(v, true)
	if !s.Solve(nil) || s.ModelValue(v) != False {
		t.Fatal("reenabling/polarity")
	}
}
func TestRandomFormulas(t *testing.T) {
	rng := rand.New(rand.NewSource(946))
	for run := 0; run < 300; run++ {
		s := New()
		n := 1 + rng.Intn(7)
		for i := 0; i < n; i++ {
			s.NewVariable()
		}
		var clauses [][]int
		for i := 0; i < rng.Intn(30); i++ {
			c := make([]int, 1+rng.Intn(4))
			for j := range c {
				c[j] = rng.Intn(2 * n)
			}
			clauses = append(clauses, c)
			s.AddClause(c)
		}
		want := false
		for mask := 0; mask < 1<<n; mask++ {
			ok := true
			for _, c := range clauses {
				sat := false
				for _, lit := range c {
					if ((mask>>(lit>>1))&1)^(lit&1) == 1 {
						sat = true
					}
				}
				if !sat {
					ok = false
					break
				}
			}
			if ok {
				want = true
				break
			}
		}
		if got := s.Solve(nil); got != want {
			t.Fatal(run, clauses, got, want)
		}
	}
}

// Review regressions: builder storage must not alias persisted clauses, and
// temporary failed assumptions must not leak into later ordinary solves.
func TestBuilderStorageAndIncrementalReview(t *testing.T) {
	s := New()
	a, b := s.NewVariable(), s.NewVariable()
	s.BeginClause()
	s.AddClauseLiteral(MakeLiteral(a, false))
	if !s.CommitClause() {
		t.Fatal("first unit")
	}
	s.BeginClause()
	s.AddClauseLiteral(MakeLiteral(b, true))
	if !s.CommitClause() {
		t.Fatal("second unit")
	}
	if !s.Solve(nil) || s.ModelValue(a) != True || s.ModelValue(b) != False {
		t.Fatal("reused builder corrupted stored clause")
	}
	if s.Solve([]int{MakeLiteral(a, true)}) {
		t.Fatal("inconsistent assumption accepted")
	}
	if !s.Solve(nil) {
		t.Fatal("temporary assumption persisted")
	}
	s.AddClause([]int{MakeLiteral(a, true)})
	if s.Okay() || s.Solve(nil) || s.ModelValue(a) != Undefined {
		t.Fatal("inconsistent formula/model state")
	}
}
func TestConflictSnapshotAndCurrentModelDistinction(t *testing.T) {
	s := New()
	a, b := s.NewVariable(), s.NewVariable()
	s.AddClause([]int{MakeLiteral(a, false), MakeLiteral(b, false)})
	if s.Solve([]int{MakeLiteral(a, true), MakeLiteral(b, true)}) {
		t.Fatal("expected conflict")
	}
	saved := s.Conflict()
	if len(saved) == 0 {
		t.Fatal("missing conflict")
	}
	saved[0] ^= 1
	if s.Conflict()[0] == saved[0] {
		t.Fatal("caller mutated stored conflict")
	}
	if !s.Solve(nil) {
		t.Fatal("formula should remain satisfiable")
	}
	if s.AssignmentCount() != 0 || s.Value(a) != Undefined || s.Value(b) != Undefined {
		t.Fatal("search assignments leaked into root state")
	}
	if s.ModelValue(a) != True && s.ModelValue(b) != True {
		t.Fatal("model snapshot lost")
	}
}

func TestIncrementalQueriesAndConflictCores(t *testing.T) {
	rng := rand.New(rand.NewSource(41734))
	for trial := 0; trial < 300; trial++ {
		const n = 5
		s := New()
		for i := 0; i < n; i++ {
			s.NewVariable()
		}
		var clauses [][]int
		satisfiable := func(assumptions []int) bool {
			for mask := 0; mask < 1<<n; mask++ {
				ok := true
				for _, lit := range assumptions {
					if ((mask>>(lit>>1))&1)^(lit&1) == 0 {
						ok = false
						break
					}
				}
				for _, clause := range clauses {
					any := false
					for _, lit := range clause {
						if ((mask>>(lit>>1))&1)^(lit&1) != 0 {
							any = true
							break
						}
					}
					if !any {
						ok = false
						break
					}
				}
				if ok {
					return true
				}
			}
			return false
		}
		for step := 0; step < 12; step++ {
			if step%2 == 0 {
				clause := []int{rng.Intn(2 * n), rng.Intn(2 * n), rng.Intn(2 * n)}
				clauses = append(clauses, append([]int(nil), clause...))
				s.AddClause(clause)
			}
			var assumptions []int
			for i := 0; i < rng.Intn(5); i++ {
				assumptions = append(assumptions, rng.Intn(2*n))
			}
			want, got := satisfiable(assumptions), s.Solve(assumptions)
			if got != want {
				t.Fatalf("trial=%d step=%d got=%v want=%v clauses=%v assumptions=%v", trial, step, got, want, clauses, assumptions)
			}
			if got {
				continue
			}
			var core []int
			for _, lit := range s.Conflict() {
				found := false
				for _, a := range assumptions {
					if a == lit^1 {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("conflict %v absent from %v", s.Conflict(), assumptions)
				}
				core = append(core, lit^1)
			}
			if satisfiable(core) {
				t.Fatalf("trial=%d step=%d satisfiable conflict core %v", trial, step, core)
			}
		}
	}
}

func TestInvalidClauseDoesNotGrowVariables(t *testing.T) {
	s := New()
	s.NewVariable()
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid literal accepted")
			}
		}()
		s.AddClause([]int{20, -1})
	}()
	if s.VariableCount() != 1 || !s.Solve(nil) {
		t.Fatal("invalid clause changed solver state")
	}
}
