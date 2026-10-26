package abc

// Go-only validation for the src/sat/satoko port.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func TestSatokoScaledDouble(t *testing.T) {
	for _, value := range []float64{1, 1.25, 2, 3.5, 10, 1e100} {
		scaled, err := MakeSatokoScaledDouble(value)
		if err != nil {
			t.Fatal(value, err)
		}
		if relative := math.Abs(scaled.Float64()-value) / value; relative > 2e-14 {
			t.Fatalf("round trip %g: %016x -> %g", value, uint64(scaled), scaled.Float64())
		}
	}
	if _, err := MakeSatokoScaledDouble(0.5); !errors.Is(err, ErrDimension) {
		t.Fatalf("small value: %v", err)
	}
	two, _ := MakeSatokoScaledDouble(2)
	three, _ := MakeSatokoScaledDouble(3)
	if got := two.Add(three).Float64(); got != 5 {
		t.Fatalf("add: %g", got)
	}
	if got := two.Multiply(three).Float64(); got != 6 {
		t.Fatalf("multiply: %g", got)
	}
	if got := SatokoScaledOne.DividePowerOfTwo(1); got != SatokoScaledOne {
		t.Fatalf("source floor: %016x", uint64(got))
	}
	nearMaximum := SatokoScaledDouble(uint64(math.MaxUint16)<<48 | uint64(SatokoScaledOne))
	if nearMaximum.Multiply(two) != SatokoScaledMax || nearMaximum.Add(nearMaximum) != SatokoScaledMax {
		t.Fatal("overflow did not saturate")
	}
}

func satokoTestLiteral(t *testing.T, variable int, negative bool) SatLit {
	t.Helper()
	literal, err := MakeSatLit(SatVar(variable), negative)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func satokoModelSatisfies(variables int, clauses [][]SatLit, assumptions []SatLit, assignment uint64) bool {
	for _, literal := range assumptions {
		value := assignment>>literal.Var()&1 != 0
		if value == literal.Negative() {
			return false
		}
	}
	for _, clause := range clauses {
		satisfied := false
		for _, literal := range clause {
			value := assignment>>literal.Var()&1 != 0
			if value != literal.Negative() {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

func satokoBruteForce(variables int, clauses [][]SatLit, assumptions []SatLit) (bool, uint64) {
	if variables < 0 || variables > 20 {
		return false, 0
	}
	for assignment := uint64(0); assignment < uint64(1)<<variables; assignment++ {
		if satokoModelSatisfies(variables, clauses, assumptions, assignment) {
			return true, assignment
		}
	}
	return false, 0
}

func TestSatokoDefaultsConfigurationAndLifecycle(t *testing.T) {
	var nilSolver *SatokoSolver
	if nilSolver.ClauseCount() != 0 {
		t.Fatal("nil solver reported clauses")
	}
	defaults := DefaultSatokoOptions()
	if defaults.RestartForce != 0.8 || defaults.RestartBlock != 1.4 || defaults.LBDQueueSize != 50 || defaults.TrailQueueSize != 5000 || defaults.FirstReduction != 2000 || defaults.LearnedRatio != 0.5 || defaults.VariableDecay != 0.95 || defaults.GarbageRatio != 0.3 {
		t.Fatalf("unexpected defaults: %+v", defaults)
	}
	if _, err := NewSatokoSolver(SatokoLimits{MaxVariables: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative limit: %v", err)
	}
	solver, err := NewSatokoSolver(SatokoLimits{MaxVariables: 3, MaxClauses: 4, MaxLiterals: 8})
	if err != nil {
		t.Fatal(err)
	}
	bad := defaults
	bad.VariableDecay = 1
	if err = solver.Configure(bad); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad options: %v", err)
	}
	got, _ := solver.Options()
	if got != defaults {
		t.Fatal("invalid configure changed options")
	}
	for i, negative := range []bool{false, true, false} {
		variable, addErr := solver.AddVariable(negative)
		if addErr != nil || int(variable) != i || solver.core.polarity[i] == negative {
			t.Fatalf("variable %d: %d %v", i, variable, addErr)
		}
		if polarity, polarityErr := solver.VariablePolarity(variable); polarityErr != nil || polarity != negative {
			t.Fatalf("polarity %d: %v %v", i, polarity, polarityErr)
		}
	}
	if _, err = solver.AddVariable(false); !errors.Is(err, ErrLimit) {
		t.Fatalf("variable limit: %v", err)
	}
	solver.Close()
	solver.Close()
	if solver.ClauseCount() != 0 {
		t.Fatal("closed solver reported clauses")
	}
	if _, err = solver.Solve(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed solve: %v", err)
	}
}

func TestSatokoClausesAssumptionsConflictAndModel(t *testing.T) {
	solver, _ := NewSatokoSolver(SatokoLimits{MaxVariables: 4})
	_ = solver.SetVariableCount(3)
	x, y, z := satokoTestLiteral(t, 0, false), satokoTestLiteral(t, 1, false), satokoTestLiteral(t, 2, false)
	clauses := [][]SatLit{{x, y}, {x.Not(), z}, {y.Not(), z}}
	for _, clause := range clauses {
		if ok, err := solver.AddClause(clause); err != nil || !ok {
			t.Fatalf("add %v: %v %v", clause, ok, err)
		}
	}
	result, err := solver.Solve()
	if err != nil || result.Status != SatSatisfiable || len(result.Model) != 3 {
		t.Fatalf("solve: %+v %v", result, err)
	}
	assignment := uint64(0)
	for i, value := range result.Model {
		if value {
			assignment |= uint64(1) << i
		}
	}
	if !satokoModelSatisfies(3, clauses, nil, assignment) {
		t.Fatalf("bad model: %v", result.Model)
	}
	if err = solver.PushAssumption(z.Not()); err != nil {
		t.Fatal(err)
	}
	result, err = solver.Solve()
	if err != nil || result.Status != SatUnsatisfiable || len(result.Conflict) == 0 {
		t.Fatalf("assumption conflict: %+v %v", result, err)
	}
	conflict, err := solver.FinalConflict()
	if err != nil || !slices.Equal(conflict, result.Conflict) {
		t.Fatalf("final conflict: %v %v", conflict, err)
	}
	if popped, popErr := solver.PopAssumption(); popErr != nil || popped != z.Not() {
		t.Fatalf("pop: %v %v", popped, popErr)
	}
	if _, err = solver.PopAssumption(); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty pop: %v", err)
	}
	if result, err = solver.SolveAssumptions([]SatLit{x.Not(), y.Not()}, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("extra assumptions: %+v %v", result, err)
	}
	if stats, statsErr := solver.Statistics(); statsErr != nil || stats.AllConflicts < result.Stats.Conflicts {
		t.Fatalf("statistics: %+v %v", stats, statsErr)
	}
	if value, valueErr := solver.CounterexampleValue(z.Var()); valueErr != nil || !value {
		t.Fatalf("saved phase: %v %v", value, valueErr)
	}
}

func TestSatokoNormalizationLimitsAndAtomicFailure(t *testing.T) {
	solver, _ := NewSatokoSolver(SatokoLimits{MaxVariables: 2, MaxClauses: 1, MaxLiterals: 2})
	x, y := satokoTestLiteral(t, 0, false), satokoTestLiteral(t, 1, false)
	if ok, err := solver.AddClause([]SatLit{x, x.Not()}); err != nil || !ok || solver.ClauseCount() != 0 {
		t.Fatalf("tautology: %v %v %d", ok, err, solver.ClauseCount())
	}
	if ok, err := solver.AddClause([]SatLit{x, x, y}); err != nil || !ok || solver.ClauseCount() != 1 || solver.literals != 2 {
		t.Fatalf("normalized: %v %v %d/%d", ok, err, solver.ClauseCount(), solver.literals)
	}
	before, _ := solver.DIMACS(SatokoDIMACSOptions{})
	if _, err := solver.AddClause([]SatLit{x.Not(), y}); !errors.Is(err, ErrLimit) {
		t.Fatalf("clause limit: %v", err)
	}
	after, _ := solver.DIMACS(SatokoDIMACSOptions{})
	if !slices.Equal(before, after) {
		t.Fatal("failed clause addition mutated solver")
	}
	if _, err := solver.AddClause([]SatLit{SatLit(^uint32(0))}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid literal: %v", err)
	}
	cancelled, _ := NewSatokoSolver(SatokoLimits{Stop: func() bool { return true }})
	if _, err := cancelled.AddClause([]SatLit{x}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled add: %v", err)
	}
}

func TestSatokoBookmarkRollbackResetAndConeMarks(t *testing.T) {
	solver, _ := NewSatokoSolver(SatokoLimits{})
	_ = solver.SetVariableCount(2)
	x, y := satokoTestLiteral(t, 0, false), satokoTestLiteral(t, 1, false)
	_, _ = solver.AddClause([]SatLit{x, y})
	_ = solver.PushAssumption(x.Not())
	if err := solver.MarkCone([]SatVar{0}); err != nil {
		t.Fatal(err)
	}
	if err := solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	_, _ = solver.AddClause([]SatLit{y.Not()})
	_, _ = solver.AddVariable(false)
	_, _ = solver.PopAssumption()
	_ = solver.UnmarkCone([]SatVar{0})
	if err := solver.Rollback(); err != nil {
		t.Fatal(err)
	}
	if solver.VariableCount() != 2 || solver.ClauseCount() != 1 || !slices.Equal(solver.assumptions, []SatLit{x.Not()}) || !solver.marks[0] {
		t.Fatalf("bad rollback: vars=%d clauses=%d assumptions=%v marks=%v", solver.VariableCount(), solver.ClauseCount(), solver.assumptions, solver.marks)
	}
	if err := solver.Unbookmark(); err != nil {
		t.Fatal(err)
	}
	if err := solver.Rollback(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("rollback without bookmark: %v", err)
	}
	if err := solver.Reset(); err != nil || solver.VariableCount() != 0 || solver.ClauseCount() != 0 || len(solver.assumptions) != 0 {
		t.Fatalf("reset: vars=%d clauses=%d err=%v", solver.VariableCount(), solver.ClauseCount(), err)
	}
}

func TestSatokoMinimizeAssumptions(t *testing.T) {
	solver, _ := NewSatokoSolver(SatokoLimits{})
	_ = solver.SetVariableCount(3)
	x, y, z := satokoTestLiteral(t, 0, false), satokoTestLiteral(t, 1, false), satokoTestLiteral(t, 2, false)
	_, _ = solver.AddClause([]SatLit{x, y})
	minimum, err := solver.MinimizeAssumptions([]SatLit{x.Not(), z, y.Not()}, SatLimits{})
	if err != nil || !slices.Equal(minimum, []SatLit{x.Not(), y.Not()}) {
		t.Fatalf("minimum: %v %v", minimum, err)
	}
	if _, err = solver.MinimizeAssumptions([]SatLit{x}, SatLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("satisfiable assumptions: %v", err)
	}
	if _, err = solver.MinimizeAssumptions([]SatLit{x}, SatLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled minimization: %v", err)
	}
}

func TestSatokoStopRunIDAndRuntimeConfiguration(t *testing.T) {
	solver, _ := NewSatokoSolver(SatokoLimits{})
	_ = solver.SetVariableCount(1)
	if err := solver.SetRunID(17); err != nil {
		t.Fatal(err)
	}
	if err := solver.SetStopForRun(func(id int) bool { return id == 17 }); err != nil {
		t.Fatal(err)
	}
	result, err := solver.Solve()
	if err != nil || result.Status != SatUnknown || result.Reason != SatCancelled {
		t.Fatalf("run stop: %+v %v", result, err)
	}
	_ = solver.SetStopForRun(nil)
	if old, limitErr := solver.SetRuntimeLimit(-1); !errors.Is(limitErr, ErrDimension) || old != 0 {
		t.Fatalf("negative runtime: %v %v", old, limitErr)
	}
}

func TestSatokoDIMACSParserRoundTripAndBounds(t *testing.T) {
	text := []byte("c tiny\np cnf 3 3\n1 2 0\n-1 3 0\n-2 3 0\n")
	solver, err := ParseSatokoDIMACS(text, SatokoLimits{})
	if err != nil || solver.VariableCount() != 3 || solver.ClauseCount() != 3 {
		t.Fatalf("parse: %v %+v", err, solver)
	}
	encoded, err := solver.DIMACS(SatokoDIMACSOptions{})
	if err != nil {
		t.Fatal(err)
	}
	copySolver, err := ParseSatokoDIMACS(encoded, SatokoLimits{})
	if err != nil {
		t.Fatalf("round trip: %v\n%s", err, encoded)
	}
	for _, assumptions := range [][]SatLit{nil, {satokoTestLiteral(t, 0, false)}, {satokoTestLiteral(t, 2, true)}} {
		one, _ := solver.SolveAssumptions(assumptions, SatLimits{})
		two, _ := copySolver.SolveAssumptions(assumptions, SatLimits{})
		if one.Status != two.Status {
			t.Fatalf("status mismatch: %v %v", one.Status, two.Status)
		}
	}
	zero, err := solver.DIMACS(SatokoDIMACSOptions{ZeroBased: true})
	if err != nil || len(zero) == 0 {
		t.Fatalf("zero based: %q %v", zero, err)
	}
	if _, err = solver.DIMACS(SatokoDIMACSOptions{MaxBytes: len(encoded) - 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("output limit: %v", err)
	}
	if _, err = ParseSatokoDIMACS([]byte("p cnf 2 1\n3 0\n"), SatokoLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatalf("bad literal: %v", err)
	}
	if _, err = ParseSatokoDIMACS(text, SatokoLimits{MaxClauses: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("parse limit: %v", err)
	}
	calls := 0
	if _, err = solver.DIMACS(SatokoDIMACSOptions{Stop: func() bool { calls++; return calls > 1 }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("output cancellation: %v", err)
	}
}

func TestSatokoDeterministicIndependentModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		state = state*6364136223846793005 + 1442695040888963407
		return state
	}
	for execution := 0; execution < 5000; execution++ {
		variables := int(next()%4) + 1
		clauseCount := int(next() % 7)
		clauses := make([][]SatLit, clauseCount)
		solver, err := NewSatokoSolver(SatokoLimits{MaxVariables: 4, MaxClauses: 8, MaxLiterals: 32})
		if err != nil {
			t.Fatal(err)
		}
		_ = solver.SetVariableCount(variables)
		for i := range clauses {
			width := int(next()%4) + 1
			clauses[i] = make([]SatLit, width)
			for j := range clauses[i] {
				clauses[i][j] = satokoTestLiteral(t, int(next()%uint64(variables)), next()&1 != 0)
			}
			_, _ = solver.AddClause(clauses[i])
		}
		assumptionCount := int(next() % 3)
		assumptions := make([]SatLit, assumptionCount)
		for i := range assumptions {
			assumptions[i] = satokoTestLiteral(t, int(next()%uint64(variables)), next()&1 != 0)
		}
		want, _ := satokoBruteForce(variables, clauses, assumptions)
		got, solveErr := solver.SolveAssumptions(assumptions, SatLimits{})
		if solveErr != nil || got.Status == SatSatisfiable != want {
			t.Fatalf("execution %d vars=%d clauses=%v assumptions=%v want=%v got=%+v err=%v", execution, variables, clauses, assumptions, want, got, solveErr)
		}
		if got.Status == SatSatisfiable {
			assignment := uint64(0)
			for i, value := range got.Model {
				if value {
					assignment |= uint64(1) << i
				}
			}
			if !satokoModelSatisfies(variables, clauses, assumptions, assignment) {
				t.Fatalf("execution %d invalid model %v", execution, got.Model)
			}
		}
	}
}

func FuzzSatokoSolverModel(f *testing.F) {
	f.Add([]byte{2, 2, 1, 0, 1, 2, 3})
	f.Add([]byte{1, 1, 0, 1})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		variables := 1
		if len(data) > 0 {
			variables = int(data[0]%5) + 1
		}
		solver, err := NewSatokoSolver(SatokoLimits{MaxVariables: 5, MaxClauses: 16, MaxLiterals: 64})
		if err != nil {
			t.Fatal(err)
		}
		_ = solver.SetVariableCount(variables)
		clauses := make([][]SatLit, 0)
		for offset := 1; offset < len(data) && len(clauses) < 12; {
			width := int(data[offset]%4) + 1
			offset++
			clause := make([]SatLit, 0, width)
			for len(clause) < width && offset < len(data) {
				clause = append(clause, satokoTestLiteral(t, int(data[offset]>>1)%variables, data[offset]&1 != 0))
				offset++
			}
			clauses = append(clauses, clause)
			if _, err = solver.AddClause(clause); err != nil {
				t.Fatal(err)
			}
		}
		want, _ := satokoBruteForce(variables, clauses, nil)
		result, err := solver.Solve()
		if err != nil || result.Status == SatSatisfiable != want {
			t.Fatalf("want=%v result=%+v err=%v", want, result, err)
		}
	})
}

func FuzzSatokoDIMACS(f *testing.F) {
	f.Add([]byte("p cnf 2 2\n1 0\n-1 2 0\n"))
	f.Add([]byte("p cnf 0 0\n"))
	f.Add([]byte("p cnf 1 1\n0\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		solver, err := ParseSatokoDIMACS(data, SatokoLimits{MaxVariables: 8, MaxClauses: 32, MaxLiterals: 128, MaxDIMACSBytes: 4096})
		if err != nil {
			return
		}
		encoded, err := solver.DIMACS(SatokoDIMACSOptions{MaxBytes: 4096})
		if err != nil {
			t.Fatal(err)
		}
		copySolver, err := ParseSatokoDIMACS(encoded, SatokoLimits{MaxVariables: 8, MaxClauses: 32, MaxLiterals: 128, MaxDIMACSBytes: 4096})
		if err != nil {
			t.Fatalf("accepted input did not round trip: %v\n%s", err, encoded)
		}
		one, err := solver.Solve()
		if err != nil {
			t.Fatal(err)
		}
		two, err := copySolver.Solve()
		if err != nil || one.Status != two.Status {
			t.Fatalf("round-trip status %v %v err=%v", one.Status, two.Status, err)
		}
	})
}
