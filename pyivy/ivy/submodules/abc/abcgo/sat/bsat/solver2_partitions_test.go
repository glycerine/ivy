package bsat

import "testing"

func TestPartitionedSolver2SourceClausesRefuteAssumptions(t *testing.T) {
	s := NewSolver2()
	defer s.Close()
	s.SetNumVars(3)
	s.SetVariablePartitionA(0, true)
	// A implies shared x1; B equates x1 and x2. Assuming !x2 refutes A & B.
	s.AddPartitionClause(true, Positive(0), Positive(1))
	s.AddPartitionClause(true, Negative(0), Positive(1))
	s.AddPartitionClause(false, Negative(1), Positive(2))
	s.AddPartitionClause(false, Positive(1), Negative(2))
	if got := s.Solve(Negative(2)); got != Unsatisfiable {
		t.Fatalf("status=%v", got)
	}
	if len(s.ProofCore()) == 0 {
		t.Fatal("solver2 did not retain a resolution core")
	}
	if got := s.Solve(Positive(2)); got != Satisfiable {
		t.Fatalf("incremental status=%v", got)
	}
}

func TestSolver2SourceVariableGrowthInvalidatesOldModel(t *testing.T) {
	s := NewSolver2()
	defer s.Close()
	s.SetNumVars(1)
	s.AddClause(Positive(0))
	if s.Solve() != Satisfiable {
		t.Fatal("unit clause is SAT")
	}
	s.SetNumVars(2)
	if model := s.Model(); model != nil {
		t.Fatalf("model for fewer variables remained visible: %v", model)
	}
}

func TestSourcePartitionClauseRejectsZeroID(t *testing.T) {
	s := NewSolver2()
	defer s.Close()
	s.SetNumVars(1)
	defer func() {
		if recover() == nil {
			t.Fatal("source forbids clause ID zero")
		}
	}()
	s.AddPartitionClauseWithID(0, true, Positive(0))
}
