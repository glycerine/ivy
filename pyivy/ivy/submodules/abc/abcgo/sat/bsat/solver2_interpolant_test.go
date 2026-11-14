package bsat

import (
	"reflect"
	"testing"
)

func TestStoredSolver2InterpolantPreservesProofAcrossReads(t *testing.T) {
	s := NewSolver2()
	defer s.Close()
	s.SetNumVars(3)
	s.SetVariablePartitionA(0, true)
	s.AddPartitionClause(true, Positive(0), Positive(1))
	s.AddPartitionClause(true, Negative(0), Positive(1))
	s.AddPartitionClause(false, Negative(1), Positive(2))
	s.AddPartitionClause(false, Negative(1), Negative(2))
	if s.Solve() != Unsatisfiable {
		t.Fatal("partitions should be inconsistent")
	}
	before := s.ProofCore()
	for repeat := 0; repeat < 2; repeat++ {
		result := s.ProofInterpolant([]int32{1})
		if result == nil || result.InputCount() != 1 || result.OutputCount() != 1 {
			t.Fatal("invalid interpolant interface")
		}
		for _, value := range []bool{false, true} {
			got, err := result.Evaluate([]bool{value})
			if err != nil || got[0] != value {
				t.Fatalf("input=%v result=%v err=%v", value, got, err)
			}
		}
		if !reflect.DeepEqual(s.ProofCore(), before) {
			t.Fatal("interpolant extraction changed stored proof")
		}
	}
}

func TestStoredInterpolantRetainsRequestedSharedVariableOrder(t *testing.T) {
	s := NewSolver2()
	defer s.Close()
	s.SetNumVars(4)
	s.SetVariablePartitionA(0, true)
	s.AddPartitionClause(true, Positive(0), Positive(1))
	s.AddPartitionClause(true, Negative(0), Positive(1))
	s.AddPartitionClause(false, Negative(1), Positive(2))
	s.AddPartitionClause(false, Negative(1), Negative(2))
	if s.ProofInterpolant([]int32{3, 1}) != nil {
		t.Fatal("interpolant appeared before a refutation")
	}
	if s.Solve() != Unsatisfiable {
		t.Fatal("expected UNSAT")
	}
	graph := s.ProofInterpolant([]int32{3, 1})
	if graph.InputCount() != 2 {
		t.Fatal("unused common variable was dropped")
	}
	for mask := 0; mask < 4; mask++ {
		got, err := graph.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || got[0] != (mask&2 != 0) {
			t.Fatalf("mask=%d output=%v err=%v", mask, got, err)
		}
	}
}
