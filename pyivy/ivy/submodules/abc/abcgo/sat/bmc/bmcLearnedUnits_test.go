package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

func TestTransferLearnedUnitsBeforeNextTarget(t *testing.T) {
	s := bsat.NewSolver()
	x, y := s.NewVariable(), s.NewVariable()
	// This formula entails x. The default solver learns x as a unit while
	// solving, which is queued for the BMC post-UNSAT maintenance step.
	if !s.AddClause(bsat.Positive(x), bsat.Positive(y)) ||
		!s.AddClause(bsat.Positive(x), bsat.Negative(y)) {
		t.Fatal("failed to add fixture clauses")
	}
	if got := s.Solve(); got != bsat.Satisfiable {
		t.Fatalf("base solve status = %v, want SAT", got)
	}
	if !transferBmcLearnedUnits(s) {
		t.Fatal("learned-unit transfer/compress reported conflict")
	}
	if units := s.TakeUnitLiterals(); len(units) != 0 {
		t.Fatalf("learned-unit queue was not consumed: %v", units)
	}
	if got := s.Solve(bsat.Negative(x)); got != bsat.Unsatisfiable {
		t.Fatalf("next target with negated learned unit = %v, want UNSAT", got)
	}
}

func TestTransferLearnedUnitsWithEmptyQueue(t *testing.T) {
	s := bsat.NewSolver()
	x := s.NewVariable()
	if !s.AddClause(bsat.Positive(x)) || !transferBmcLearnedUnits(s) {
		t.Fatal("empty-queue transfer should preserve a consistent solver")
	}
	if got := s.Solve(bsat.Negative(x)); got != bsat.Unsatisfiable {
		t.Fatalf("target contradicting root unit = %v, want UNSAT", got)
	}
}
