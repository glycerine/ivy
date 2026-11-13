package minisatp

import (
	"reflect"
	"testing"
)

func TestImmediateInconsistentSolveRetainsPreviousModel(t *testing.T) {
	s := newSolver(1)
	if s.solve() != 1 || len(s.model) != 1 {
		t.Fatal("first SAT solve")
	}
	previous := append([]int8(nil), s.model...)
	if !s.addClause([]int32{0}) || s.addClause([]int32{1}) {
		t.Fatal("contradictory units")
	}
	if s.solve() != 0 {
		t.Fatal("expected immediate UNSAT")
	}
	if !reflect.DeepEqual(s.model, previous) {
		t.Fatalf("upstream solve preserves model when !ok: got %v want %v", s.model, previous)
	}
}
