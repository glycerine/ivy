package bsat

import "testing"

func TestSolver2SimplifyPropagatesRootUnits(t *testing.T) {
	s := NewSolver2()
	x, y := s.NewVariable(), s.NewVariable()
	if !s.AddClause(Negative(x), Positive(y)) || !s.AddClause(Positive(x)) {
		t.Fatal("unexpected contradiction")
	}
	if !s.Simplify() {
		t.Fatal("consistent root propagation failed")
	}
	if got := s.Solve(Negative(y)); got != Unsatisfiable {
		t.Fatalf("root implication ignored: %v", got)
	}
	if got := s.Solve(Positive(y)); got != Satisfiable {
		t.Fatalf("consistent assumption failed: %v", got)
	}
}

func TestSolver2SimplifyDetectsRootConflict(t *testing.T) {
	s := NewSolver2()
	x, y := s.NewVariable(), s.NewVariable()
	if !s.AddClause(Negative(x), Positive(y)) || !s.AddClause(Negative(x), Negative(y)) || !s.AddClause(Positive(x)) {
		t.Fatal("fixture contradicted before propagation")
	}
	if s.Simplify() {
		t.Fatal("missed root conflict")
	}
}

// The C solver consumes the conflicting trail entries. Its callers abandon
// the solve after simplify fails; this records why continuing is unsupported.
func TestSolver2SimplifyFailedContinuationMatchesC(t *testing.T) {
	s := NewSolver2()
	x, y := s.NewVariable(), s.NewVariable()
	if !s.AddClause(Negative(x), Positive(y)) || !s.AddClause(Negative(x), Negative(y)) || !s.AddClause(Positive(x)) {
		t.Fatal("fixture contradicted before propagation")
	}
	if s.Simplify() || s.Simplify() {
		t.Fatal("expected both simplifications to find a conflict")
	}
	if got := s.Solve(); got != Satisfiable {
		t.Fatalf("continuation differs from C Solver2: %v", got)
	}
}
