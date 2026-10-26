package abc

// Tests for the cgtSat.c implication checker.

import (
	"errors"
	"testing"
)

func cgtTestLiteral(variable int, negative bool) SatLit {
	literal, err := MakeSatLit(SatVar(variable), negative)
	if err != nil {
		panic(err)
	}
	return literal
}

func cgtTestSolver(t testing.TB, variables int, clauses [][]SatLit) *SatSolver3 {
	t.Helper()
	solver, err := NewSatSolver3(SatSolver3Options{MaxVariables: max(variables, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if err = solver.SetVariableCount(variables); err != nil {
		t.Fatal(err)
	}
	for _, clause := range clauses {
		if ok, addErr := solver.AddClause(clause); addErr != nil {
			t.Fatal(addErr)
		} else if !ok {
			break
		}
	}
	return solver
}

func TestClockGateImplicationSATUNSATLearningAndComplement(t *testing.T) {
	g, m, x := cgtTestLiteral(0, false), cgtTestLiteral(1, false), cgtTestLiteral(2, false)

	// g => x => !m. The binary consequence is not initially present.
	solver := cgtTestSolver(t, 3, [][]SatLit{{g.Not(), x}, {x.Not(), m.Not()}})
	session, err := NewClockGateImplicationSession(solver, ClockGateImplicationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	before := solver.OriginalClauseCount()
	answer, err := session.Check(g, m)
	if err != nil || !answer.Implication || answer.SAT.Status != SatUnsatisfiable || !answer.ClauseInstalled {
		t.Fatalf("unexpected implication answer: %+v, %v", answer, err)
	}
	if solver.OriginalClauseCount() != before+1 {
		t.Fatalf("entailed binary clause was not retained: %d -> %d", before, solver.OriginalClauseCount())
	}
	if answer.SAT.Conflict == nil {
		t.Fatal("UNSAT assumptions should return an owned conflict")
	}
	answer.SAT.Conflict[0] ^= 1
	repeat, err := session.Check(g, m)
	if err != nil || !repeat.Implication || repeat.SAT.Status != SatUnsatisfiable {
		t.Fatalf("learned implication was not reusable: %+v, %v", repeat, err)
	}

	// With no relation, g AND m is satisfiable and no consequence is learned.
	freeSolver := cgtTestSolver(t, 2, nil)
	free, err := NewClockGateImplicationSession(freeSolver, ClockGateImplicationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	freeAnswer, err := free.Check(g, m)
	if err != nil || freeAnswer.Implication || freeAnswer.SAT.Status != SatSatisfiable || freeAnswer.ClauseInstalled {
		t.Fatalf("unexpected SAT answer: %+v, %v", freeAnswer, err)
	}
	if len(freeAnswer.SAT.Model) != 2 || !freeAnswer.SAT.Model[0] || !freeAnswer.SAT.Model[1] {
		t.Fatalf("assumptions missing from model: %v", freeAnswer.SAT.Model)
	}

	// A complemented gate checks !g => !m and learns (g OR !m).
	complementedSolver := cgtTestSolver(t, 2, [][]SatLit{{g, m.Not()}})
	complemented, err := NewClockGateImplicationSession(complementedSolver, ClockGateImplicationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	complementedAnswer, err := complemented.Check(g.Not(), m)
	if err != nil || !complementedAnswer.Implication || complementedAnswer.SAT.Status != SatUnsatisfiable {
		t.Fatalf("complemented gate mismatch: %+v, %v", complementedAnswer, err)
	}

	// If root propagation already proves !g, the learned consequence is true
	// at level zero. AddClause succeeds without retaining another clause;
	// ClauseInstalled and its statistic must describe storage, not success.
	rootSolver := cgtTestSolver(t, 2, [][]SatLit{{g.Not()}})
	rootSession, err := NewClockGateImplicationSession(rootSolver, ClockGateImplicationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rootBefore := rootSolver.OriginalClauseCount()
	rootAnswer, err := rootSession.Check(g, m)
	if err != nil || !rootAnswer.Implication || rootAnswer.SAT.Status != SatUnsatisfiable || rootAnswer.ClauseInstalled {
		t.Fatalf("root implication mismatch: %+v, %v", rootAnswer, err)
	}
	if rootSolver.OriginalClauseCount() != rootBefore {
		t.Fatalf("root-satisfied clause was unexpectedly retained: %d -> %d", rootBefore, rootSolver.OriginalClauseCount())
	}
	rootStats, err := rootSession.Stats()
	if err != nil || rootStats.ClausesInstalled != 0 || rootStats.Unsatisfiable != 1 {
		t.Fatalf("root implication stats: %+v, %v", rootStats, err)
	}
}

func TestClockGateImplicationUnknownStatsLimitsAndLifecycle(t *testing.T) {
	g, m := cgtTestLiteral(0, false), cgtTestLiteral(1, false)
	stops := 0
	solver := cgtTestSolver(t, 2, nil)
	session, err := NewClockGateImplicationSession(solver, ClockGateImplicationOptions{
		MaxCalls: 1,
		Stop: func() bool {
			stops++
			return true
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	answer, err := session.Check(g, m)
	if err != nil || answer.SAT.Status != SatUnknown || answer.SAT.Reason != SatCancelled || answer.Implication {
		t.Fatalf("cancel result: %+v, %v", answer, err)
	}
	if stops == 0 {
		t.Fatal("stop callback was not polled")
	}
	stats, err := session.Stats()
	if err != nil || stats.Calls != 1 || stats.Undecided != 1 || stats.Satisfiable != 0 || stats.Unsatisfiable != 0 {
		t.Fatalf("stats: %+v, %v", stats, err)
	}
	if _, err = session.Check(g, m); !errors.Is(err, ErrLimit) {
		t.Fatalf("call limit error = %v", err)
	}
	statsAfter, _ := session.Stats()
	if statsAfter != stats {
		t.Fatalf("rejected call changed stats: before=%+v after=%+v", stats, statsAfter)
	}
	if err = session.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Stats(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed stats error = %v", err)
	}
	if _, err = session.Check(g, m); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed check error = %v", err)
	}
	if err = session.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("double close error = %v", err)
	}
	// The session never owns the solver.
	if result, solveErr := solver.Solve(nil, SatSolver3Limits{}); solveErr != nil || result.Status != SatSatisfiable {
		t.Fatalf("caller-owned solver unusable after close: %+v, %v", result, solveErr)
	}
}

func TestClockGateImplicationConflictLimit(t *testing.T) {
	// Three pigeons in two holes needs search to prove UNSAT. Gate and miter
	// are separate unconstrained variables so the single-conflict budget is
	// exercised by the base relation rather than by assumption propagation.
	clauses := make([][]SatLit, 0, 9)
	for pigeon := 0; pigeon < 3; pigeon++ {
		clauses = append(clauses, []SatLit{
			cgtTestLiteral(2+2*pigeon, false),
			cgtTestLiteral(3+2*pigeon, false),
		})
	}
	for hole := 0; hole < 2; hole++ {
		for first := 0; first < 3; first++ {
			for second := first + 1; second < 3; second++ {
				clauses = append(clauses, []SatLit{
					cgtTestLiteral(2+2*first+hole, true),
					cgtTestLiteral(2+2*second+hole, true),
				})
			}
		}
	}
	solver := cgtTestSolver(t, 8, clauses)
	session, err := NewClockGateImplicationSession(solver, ClockGateImplicationOptions{ConflictLimit: 1})
	if err != nil {
		t.Fatal(err)
	}
	answer, err := session.Check(cgtTestLiteral(0, false), cgtTestLiteral(1, false))
	if err != nil || answer.SAT.Status != SatUnknown || answer.SAT.Reason != SatConflictLimit || answer.Implication {
		t.Fatalf("conflict-limited answer: %+v, %v", answer, err)
	}
	stats, err := session.Stats()
	if err != nil || stats.Calls != 1 || stats.Undecided != 1 || stats.Solver.Conflicts != 1 {
		t.Fatalf("conflict-limited stats: %+v, %v", stats, err)
	}
}

func TestClockGateImplicationValidationAndAtomicErrors(t *testing.T) {
	g, m := cgtTestLiteral(0, false), cgtTestLiteral(1, false)
	if session, err := NewClockGateImplicationSession(nil, ClockGateImplicationOptions{}); session != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil solver: %v, %v", session, err)
	}
	stopped := cgtTestSolver(t, 2, nil)
	stopped.Stop()
	if session, err := NewClockGateImplicationSession(stopped, ClockGateImplicationOptions{}); session != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("stopped solver: %v, %v", session, err)
	}
	solver := cgtTestSolver(t, 2, nil)
	session, err := NewClockGateImplicationSession(solver, ClockGateImplicationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		gate, miter SatLit
		want        error
	}{
		{g, g, ErrDimension},
		{g, m.Not(), ErrDimension},
		{cgtTestLiteral(2, false), m, ErrInvalidLiteral},
		{SatLit(^uint32(0)), m, ErrInvalidLiteral},
	}
	for _, test := range cases {
		if _, checkErr := session.Check(test.gate, test.miter); !errors.Is(checkErr, test.want) {
			t.Fatalf("Check(%d,%d) error = %v, want %v", test.gate, test.miter, checkErr, test.want)
		}
	}
	stats, err := session.Stats()
	if err != nil || stats != (ClockGateImplicationStats{}) {
		t.Fatalf("invalid calls changed stats: %+v, %v", stats, err)
	}
	if _, err = (*ClockGateImplicationSession)(nil).Check(g, m); !errors.Is(err, ErrInvariant) {
		t.Fatalf("nil receiver error = %v", err)
	}
	if err = (*ClockGateImplicationSession)(nil).Close(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("nil close error = %v", err)
	}
}

func cgtTruthSolver(t testing.TB, allowed byte) *SatSolver3 {
	t.Helper()
	clauses := make([][]SatLit, 0, 8)
	for assignment := 0; assignment < 8; assignment++ {
		if allowed&(1<<assignment) != 0 {
			continue
		}
		clause := make([]SatLit, 3)
		for variable := 0; variable < 3; variable++ {
			// The blocking clause is false only on this assignment.
			clause[variable] = cgtTestLiteral(variable, assignment&(1<<variable) != 0)
		}
		clauses = append(clauses, clause)
	}
	return cgtTestSolver(t, 3, clauses)
}

func cgtTruthImplication(allowed byte, gateVariable int, gateNegative bool, miterVariable int) bool {
	for assignment := 0; assignment < 8; assignment++ {
		if allowed&(1<<assignment) == 0 {
			continue
		}
		gate := assignment&(1<<gateVariable) != 0
		if gateNegative {
			gate = !gate
		}
		miter := assignment&(1<<miterVariable) != 0
		if gate && miter {
			return false
		}
	}
	return true
}

func cgtCheckTruthCase(t testing.TB, allowed byte, gateVariable int, gateNegative bool, miterVariable int) {
	t.Helper()
	if gateVariable == miterVariable {
		miterVariable = (miterVariable + 1) % 3
	}
	solver := cgtTruthSolver(t, allowed)
	session, err := NewClockGateImplicationSession(solver, ClockGateImplicationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	gate := cgtTestLiteral(gateVariable, gateNegative)
	miter := cgtTestLiteral(miterVariable, false)
	answer, err := session.Check(gate, miter)
	if err != nil {
		t.Fatal(err)
	}
	want := cgtTruthImplication(allowed, gateVariable, gateNegative, miterVariable)
	if answer.Implication != want || (answer.SAT.Status == SatUnsatisfiable) != want || answer.SAT.Status == SatUnknown {
		t.Fatalf("allowed=%08b gate=%d/%t miter=%d: answer=%+v want=%t", allowed, gateVariable, gateNegative, miterVariable, answer, want)
	}
}

func TestClockGateImplicationModelCampaign(t *testing.T) {
	state := uint32(0x51c4a72d)
	for execution := 0; execution < 5000; execution++ {
		state = state*1664525 + 1013904223
		allowed := byte(state >> 24)
		gateVariable := int((state >> 8) % 3)
		gateNegative := state&(1<<4) != 0
		miterVariable := int((state >> 16) % 3)
		cgtCheckTruthCase(t, allowed, gateVariable, gateNegative, miterVariable)
	}
}

func TestClockGateImplicationExhaustiveTruthTables(t *testing.T) {
	for allowed := 0; allowed < 256; allowed++ {
		for gateVariable := 0; gateVariable < 3; gateVariable++ {
			for miterVariable := 0; miterVariable < 3; miterVariable++ {
				if gateVariable == miterVariable {
					continue
				}
				for _, gateNegative := range []bool{false, true} {
					cgtCheckTruthCase(t, byte(allowed), gateVariable, gateNegative, miterVariable)
				}
			}
		}
	}
}

func FuzzClockGateImplication(f *testing.F) {
	for _, seed := range []struct {
		allowed, gate, miter byte
	}{{0, 0, 1}, {0xff, 1, 2}, {0x69, 2, 0}, {0x96, 1, 1}} {
		f.Add(seed.allowed, seed.gate, seed.miter)
	}
	f.Fuzz(func(t *testing.T, allowed, gate, miter byte) {
		cgtCheckTruthCase(t, allowed, int(gate%3), gate&0x80 != 0, int(miter%3))
	})
}
