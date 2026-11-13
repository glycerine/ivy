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
