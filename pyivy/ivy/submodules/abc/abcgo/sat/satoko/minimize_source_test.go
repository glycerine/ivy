package satoko

import "testing"

func TestMinimizeAssumptionsSourceDropsIrrelevantAssumption(t *testing.T) {
	s := New()
	s.SetVariables(3)
	// x and y cannot both be true; z is irrelevant to the contradiction.
	if !s.AddClause([]int{1, 3}) { // !x OR !y
		t.Fatal("failed to add constraint")
	}
	assumptions := []int{0, 2, 4} // x, y, z
	if status := s.SolveAssumptions(assumptions); status != SATOKO_UNSAT {
		t.Fatalf("initial assumptions status = %d, want UNSAT", status)
	}
	kept := s.MinimizeAssumptionsSource(assumptions, 0)
	if kept != 2 {
		t.Fatalf("minimized assumption count = %d, want 2 (%v)", kept, assumptions)
	}
	if s.SolveAssumptions(assumptions[:kept]) != SATOKO_UNSAT {
		t.Fatalf("retained assumptions %v are not UNSAT", assumptions[:kept])
	}
	for i := 0; i < kept; i++ {
		if assumptions[i] == 4 {
			t.Fatal("irrelevant z assumption was retained")
		}
	}
}

func TestMinimizeAssumptionsSourcePreservesPushedContext(t *testing.T) {
	s := New()
	s.SetVariables(3)
	if !s.AddClause([]int{1, 3}) { // !x OR !y
		t.Fatal("failed to add constraint")
	}
	s.PushAssumption(0)        // x
	assumptions := []int{2, 4} // y and irrelevant z (created below)
	if kept := s.MinimizeAssumptionsSource(assumptions, 0); kept != 1 || assumptions[0] != 2 {
		t.Fatalf("minimized under pushed x context = %v (count %d), want [y]", assumptions, kept)
	}
	s.PopAssumption()
}
