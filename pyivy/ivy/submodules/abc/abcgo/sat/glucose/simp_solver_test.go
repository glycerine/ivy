package glucose

import (
	"math/rand"
	"testing"
)

func TestSimpRejectsEliminatedAssumptionWithoutPreprocessing(t *testing.T) {
	for _, turnOff := range []bool{false, true} {
		s := NewSimpSolver()
		s.NewVar(true, true)
		s.NewVar(true, true)
		s.SetFrozen(1, true)
		s.AddClause([]Lit{0, 2})
		if !s.Eliminate(turnOff) || !s.IsEliminated(0) {
			t.Fatal("expected eliminated variable")
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("turnOff=%v: eliminated assumption accepted", turnOff)
				}
			}()
			s.SolveLimited([]Lit{1}, false, false)
		}()
		if s.SolveLimited([]Lit{2}, false, false) != LTrue {
			t.Fatal("solver unusable after rejected assumption")
		}
	}
}

func TestSimpRejectsInvalidSubstitutionBeforeMutation(t *testing.T) {
	s := NewSimpSolver()
	s.NewVar(true, true)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid replacement accepted")
			}
		}()
		s.Substitute(0, MkLit(5, false))
	}()
	if s.IsEliminated(0) {
		t.Fatal("invalid substitution modified solver")
	}
}

func TestSimpClearsModelOnPreprocessingFailure(t *testing.T) {
	s := NewSimpSolver()
	s.NewVar(true, true)
	s.SetFrozen(0, true)
	if s.SolveLimited(nil) != LTrue || len(s.Model) == 0 {
		t.Fatal("expected initial model")
	}
	s.AddEmptyClause()
	if s.SolveLimited(nil) != LFalse {
		t.Fatal("expected unsatisfiable result")
	}
	if len(s.Model) != 0 || len(s.Conflict) != 0 {
		t.Fatalf("stale result model=%v conflict=%v", s.Model, s.Conflict)
	}
}

func TestSimpIncrementalClausesAndModels(t *testing.T) {
	rng := rand.New(rand.NewSource(723489))
	for trial := 0; trial < 300; trial++ {
		const n = 6
		s := NewSimpSolver()
		s.UseAsymm, s.UseRCheck = trial%3 == 0, trial%5 == 0
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
			if i < 3 {
				s.SetFrozen(Var(i), true)
			}
		}
		var formula [][]Lit
		for i := 0; i < 3+rng.Intn(16); i++ {
			clause := make([]Lit, 1+rng.Intn(4))
			for j := range clause {
				clause[j] = MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0)
			}
			formula = append(formula, append([]Lit(nil), clause...))
			s.AddClause(clause)
		}
		for step := 0; step < 8; step++ {
			var assumptions []Lit
			for i := 0; i < rng.Intn(4); i++ {
				assumptions = append(assumptions, MkLit(Var(rng.Intn(3)), rng.Intn(2) != 0))
			}
			want := bruteCNF(n, formula, assumptions)
			if got := s.Solve(assumptions); got != want {
				t.Fatalf("trial=%d step=%d got=%v want=%v formula=%v assumptions=%v", trial, step, got, want, formula, assumptions)
			}
			if want {
				for _, clause := range formula {
					satisfied := false
					for _, lit := range clause {
						if s.ModelValue(lit) == LTrue {
							satisfied = true
							break
						}
					}
					if !satisfied {
						t.Fatalf("trial=%d step=%d invalid model %v clause %v", trial, step, s.Model, clause)
					}
				}
				for _, lit := range assumptions {
					if s.ModelValue(lit) != LTrue {
						t.Fatalf("trial=%d step=%d assumption %v model=%v", trial, step, lit, s.Model)
					}
				}
			}
			if step%2 == 1 {
				clause := []Lit{MkLit(Var(rng.Intn(3)), rng.Intn(2) != 0), MkLit(Var(rng.Intn(3)), rng.Intn(2) != 0)}
				formula = append(formula, append([]Lit(nil), clause...))
				s.AddClause(clause)
			}
		}
	}
}

func TestSimpSolverRandomCNF(t *testing.T) {
	rng := rand.New(rand.NewSource(8231))
	for trial := 0; trial < 1500; trial++ {
		n := 1 + rng.Intn(9)
		s := NewSimpSolver()
		s.UseAsymm = trial%3 == 0
		s.UseRCheck = trial%5 == 0
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
		}
		var cnf [][]Lit
		for count := rng.Intn(45); count > 0; count-- {
			var row []Lit
			for k := 1 + rng.Intn(5); k > 0; k-- {
				row = append(row, MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0))
			}
			cnf = append(cnf, row)
			s.AddClause(row)
		}
		var as []Lit
		for k := rng.Intn(4); k > 0; k-- {
			as = append(as, MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0))
		}
		want := bruteCNF(n, cnf, as)
		if got := s.Solve(as); got != want {
			t.Fatalf("trial%d got%v want%v cnf%v assumptions%v", trial, got, want, cnf, as)
		}
		if want {
			for _, c := range cnf {
				sat := false
				for _, p := range c {
					if s.ModelValue(p) == LTrue {
						sat = true
						break
					}
				}
				if !sat {
					t.Fatalf("trial%d bad model clause%v model%v", trial, c, s.Model)
				}
			}
		}
	}
}
func TestSimpSolverSubstitutionAndReuse(t *testing.T) {
	s := NewSimpSolver()
	for i := 0; i < 3; i++ {
		s.NewVar(true, true)
	}
	s.SetFrozen(1, true)
	s.SetFrozen(2, true)
	s.AddClause([]Lit{0, 2})
	s.AddClause([]Lit{1, 4})
	if !s.Substitute(0, 3) {
		t.Fatal("substitution")
	}
	if !s.Solve([]Lit{2, 4}) || s.Model[0] != LFalse {
		t.Fatal("substitution extension", s.Model)
	}
	s.Reset()
	s.NewVar(true, true)
	s.SetFrozen(0, true)
	s.AddClause([]Lit{1})
	if !s.Solve(nil) || s.Model[0] != LFalse {
		t.Fatal("reset")
	}
}

func TestSimpRCheckMustRetainResolvents(t *testing.T) {
	s := NewSimpSolver()
	s.UseRCheck = true
	for i := 0; i < 4; i++ {
		s.NewVar(true, true)
		if i > 0 {
			s.SetFrozen(Var(i), true)
		}
	}
	s.AddClause([]Lit{0, 2, 4})
	s.AddClause([]Lit{1, 2, 6})
	if s.Solve([]Lit{3, 5, 7}) {
		t.Fatal("resolution must retain (a OR b OR c) when both x parents are removed")
	}
}

func TestSimpSolverNativeStatusOracle(t *testing.T) {
	state := uint32(8231)
	next := func() uint32 { state = state*1664525 + 1013904223; return state }
	hash := uint64(1469598103934665603)
	for trial := 0; trial < 1500; trial++ {
		s := NewSimpSolver()
		n := 1 + int(next()%9)
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
		}
		count := int(next() % 45)
		for c := 0; c < count; c++ {
			length := 1 + int(next()%5)
			var ps []Lit
			for k := 0; k < length; k++ {
				v := Var(next() % uint32(n))
				sign := next()>>31 != 0
				ps = append(ps, MkLit(v, sign))
			}
			s.AddClause(ps)
		}
		count = int(next() % 4)
		var as []Lit
		for k := 0; k < count; k++ {
			v := Var(next() % uint32(n))
			sign := next()>>31 != 0
			as = append(as, MkLit(v, sign))
		}
		if s.Solve(as) {
			hash ^= 1
		}
		hash *= 1099511628211
	}
	if hash != 818442487605194937 {
		t.Fatalf("native preprocessing status oracle %d", hash)
	}
}
