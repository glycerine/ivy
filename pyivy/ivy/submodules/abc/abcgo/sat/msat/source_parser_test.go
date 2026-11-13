package msat

import (
	"strings"
	"testing"
)

func TestSourceDIMACSParserSimplifiesBeforeReturning(t *testing.T) {
	var p *Msat_Solver_t
	if Msat_SolverParseDimacs(strings.NewReader("p cnf 2 2\n1 2 0\n1 0\n"), &p, 0) != 1 {
		t.Fatal("parse")
	}
	if Msat_SolverReadClauseNum(p) != 0 {
		t.Fatalf("satisfied clause not removed: %d", Msat_SolverReadClauseNum(p))
	}
}

func TestSourceDIMACSParserLeavesOutputOnRejectedClause(t *testing.T) {
	old := New(1)
	p := old
	if Msat_SolverParseDimacs(strings.NewReader("p cnf 1 2\n1 0\n-1 0\n"), &p, 0) != 0 {
		t.Fatal("expected rejected contradictory clause")
	}
	if p != old {
		t.Fatal("C does not update output pointer for clause rejection")
	}
}

func TestSolverAllocInitialSeenID(t *testing.T) {
	p := Msat_SolverAlloc(2, 1, 1, 1, 1, 0)
	if got := Msat_SolverIncrementSeenId(p); got != 2 {
		t.Fatalf("C initializes seen id to 1; next = %d", got)
	}
}

func TestSourceDIMACSParserAcceptsCommentAfterClause(t *testing.T) {
	var p *Msat_Solver_t
	text := "p cnf 2 2\n1 2 0 c trailing comment\n-1 2 0\n"
	if got := Msat_SolverParseDimacs(strings.NewReader(text), &p, 0); got != 1 {
		t.Fatal("C accepts a comment after the clause terminator")
	}
}

func TestSourceDIMACSParserDoesNotRequireDeclaredClauseCount(t *testing.T) {
	var p *Msat_Solver_t
	if Msat_SolverParseDimacs(strings.NewReader("p cnf 2 3\n1 2 0\n"), &p, 0) != 1 {
		t.Fatal("C reads but does not enforce declared clause count")
	}
}

func TestSourceDIMACSParserDoesNotValidateProblemKind(t *testing.T) {
	text := "p other 2 1\n1 2 0\n"
	var p *Msat_Solver_t
	if Msat_SolverParseDimacs(strings.NewReader(text), &p, 0) != 1 {
		t.Fatal("C skips the problem-kind token")
	}
	if _, err := ParseDIMACS(strings.NewReader(text)); err == nil {
		t.Fatal("Go convenience parser should remain strict")
	}
}
