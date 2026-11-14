package bsat

import "testing"

func TestACBBridgeSetVariablesAndXor(t *testing.T) {
	s := NewSolver()
	s.SetNumVars(3)
	if s.NumVars() != 3 {
		t.Fatalf("variable count = %d, want 3", s.NumVars())
	}
	if !s.AddXor(0, 1, 2, false) {
		t.Fatal("failed to add XOR constraint")
	}
	for a := int32(0); a < 2; a++ {
		for b := int32(0); b < 2; b++ {
			for out := int32(0); out < 2; out++ {
				assumptions := []Literal{Negative(0), Negative(1), Negative(2)}
				if a != 0 {
					assumptions[1] = Positive(1)
				}
				if b != 0 {
					assumptions[2] = Positive(2)
				}
				if out != 0 {
					assumptions[0] = Positive(0)
				}
				got := s.Solve(assumptions...)
				wantSat := out == (a ^ b)
				if (got == Satisfiable) != wantSat {
					t.Errorf("a=%d b=%d out=%d: status %v, want SAT=%v", a, b, out, got, wantSat)
				}
			}
		}
	}
}

func TestACBBridgeUserValues(t *testing.T) {
	s := NewSolver()
	s.SetNumVars(2)
	s.SetUserVariables([]int32{1, 0})
	if got := s.Solve(Negative(1), Positive(0)); got != Satisfiable {
		t.Fatalf("solve status = %v, want SAT", got)
	}
	if got, want := s.UserValues(), []int32{0, 1}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("user values = %v, want %v", got, want)
	}
	s.ClearUserValues()
	if s.UserValueCount() != 0 {
		t.Fatalf("user value count after clear = %d, want 0", s.UserValueCount())
	}
	s.ClearUserVariables()
}
