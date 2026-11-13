package bsat

import (
	"reflect"
	"testing"
)

func TestLabeledSolver2CoreAndRollback(t *testing.T) {
	s := NewLabeledSolver2()
	if !s.AddClause(-1, Positive(0)) || s.Solve() != Satisfiable {
		t.Fatal("background unit")
	}
	s.Bookmark()
	if !s.AddClause(42, Negative(0), Positive(1)) {
		t.Fatal("labeled implication")
	}
	if s.AddClause(7, Negative(1)) {
		t.Fatal("contradictory unit accepted")
	}
	if s.Solve() != Unsatisfiable {
		t.Fatal("expected conflict")
	}
	if got := s.ProofCore(); !reflect.DeepEqual(got, []int32{42, 7}) {
		t.Fatalf("core %v", got)
	}
	s.Rollback()
	if s.NumVars() != 1 || s.ProofCore() != nil || s.Solve() != Satisfiable {
		t.Fatal("rollback failed")
	}
	if s.AddClause(23, Negative(0)) {
		t.Fatal("second contradiction accepted")
	}
	if s.Solve() != Unsatisfiable {
		t.Fatal("expected second conflict")
	}
	if got := s.ProofCore(); !reflect.DeepEqual(got, []int32{23}) {
		t.Fatalf("core after rollback %v", got)
	}
}

func TestLabeledSolver2GroupsSharedIDs(t *testing.T) {
	s := NewLabeledSolver2()
	for _, clause := range [][]Literal{{0, 2}, {0, 3}, {1, 2}, {1, 3}} {
		if !s.AddClause(91, clause...) {
			t.Fatal("unexpected root contradiction")
		}
	}
	if s.Solve() != Unsatisfiable {
		t.Fatal("expected unsatisfiable formula")
	}
	if got := s.ProofCore(); !reflect.DeepEqual(got, []int32{91}) {
		t.Fatalf("grouped core %v", got)
	}
}

func TestLabeledSolver2RejectsInvalidIDWithoutMutation(t *testing.T) {
	s := NewLabeledSolver2()
	func() {
		defer func() {
			if recover() == nil {
				t.Error("zero ID accepted")
			}
		}()
		s.AddClause(0, Positive(3))
	}()
	if s.NumVars() != 0 || s.Solve() != Satisfiable {
		t.Fatal("invalid ID changed solver")
	}
}
