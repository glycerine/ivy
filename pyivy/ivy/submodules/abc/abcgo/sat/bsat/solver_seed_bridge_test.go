package bsat

import "testing"

func TestNewSolverSeed(t *testing.T) {
	for _, seed := range []float64{0, 1, 42, 91648253} {
		s := NewSolverSeed(seed)
		if s.random_seed != seed || s.Mem.nPageSize != 15 {
			t.Fatalf("seed %v: got random seed %v and page size %d", seed, s.random_seed, s.Mem.nPageSize)
		}
		if !s.AddClause(0) || s.Solve() != Satisfiable {
			t.Fatalf("seed %v: solver could not solve a unit clause", seed)
		}
	}
}

func TestRestartSeedKeepsSourceStateAndInvalidatesModel(t *testing.T) {
	s := NewSolverSeed(7)
	s.NewVariable()
	s.SetActivityFactor(0, 2)
	s.SetLimits(11, 13)
	if !s.AddClause(Positive(0)) || s.Solve() != Satisfiable {
		t.Fatal("initial solve")
	}
	s.RestartSeed(42)
	if s.random_seed != 42 || s.NumVars() != 0 || s.act_vars.size != 1 || s.factors[0] != 2 {
		t.Fatalf("restart lost source state: seed=%v vars=%d cone=%v factors=%v", s.random_seed, s.NumVars(), veci_begin(&s.act_vars), s.factors)
	}
	if c, p := s.Limits(); c != 11 || p != 13 {
		t.Fatalf("restart changed controls: %d, %d", c, p)
	}
	if s.Model() != nil {
		t.Fatal("restart retained a stale model")
	}
}
