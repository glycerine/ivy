package abc

// Stand-alone behavioral tests for src/sat/bsat/satSolver3.c.

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

func satSolver3TestLiteral(t testing.TB, variable int, negative bool) SatLit {
	t.Helper()
	literal, err := MakeSatLit(SatVar(variable), negative)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func TestSatSolver3PushPopPropagationAndConflict(t *testing.T) {
	solver, err := NewSatSolver3(SatSolver3Options{MaxVariables: 8})
	if err != nil || solver.SetVariableCount(3) != nil {
		t.Fatal(err)
	}
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	z := satSolver3TestLiteral(t, 2, false)
	if ok, addErr := solver.AddClause([]SatLit{x.Not(), y}); !ok || addErr != nil {
		t.Fatal(ok, addErr)
	}
	if ok, pushErr := solver.Push(x); !ok || pushErr != nil || solver.CountAssigned() != 2 {
		t.Fatal(ok, pushErr, solver.CountAssigned())
	}
	if value, valueErr := solver.VariableAssignment(x.Var()); valueErr != nil || value != SatSatisfiable {
		t.Fatal(value, valueErr)
	}
	if value, valueErr := solver.VariableAssignment(z.Var()); valueErr != nil || value != SatUnknown {
		t.Fatal(value, valueErr)
	}
	if ok, pushErr := solver.Push(y.Not()); ok || pushErr != nil {
		t.Fatal(ok, pushErr)
	}
	if got := solver.PushedAssumptions(); !slices.Equal(got, []SatLit{x}) || len(solver.PushConflict()) == 0 {
		t.Fatal(got, solver.PushConflict())
	}
	result, solveErr := solver.Solve([]SatLit{z}, SatSolver3Limits{})
	if solveErr != nil || result.Status != SatSatisfiable || !result.Model[0] || !result.Model[1] || !result.Model[2] {
		t.Fatal(result, solveErr)
	}
	if err = solver.Pop(); err != nil || len(solver.PushedAssumptions()) != 0 {
		t.Fatal(err, solver.PushedAssumptions())
	}
	if err = solver.Pop(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatSolver3LimitsActivityBookmarkRestartAndLifecycle(t *testing.T) {
	solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: 8, Seed: 7})
	_ = solver.SetVariableCount(4)
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	if ok, err := solver.AddClause([]SatLit{x}); !ok || err != nil {
		t.Fatal(ok, err)
	}
	stats := solver.Statistics()
	if stats.Propagations == 0 {
		t.Fatal(stats)
	}
	if used, usedErr := solver.CountUsedVariables(); usedErr != nil || used != 1 {
		t.Fatal(used, usedErr)
	}
	if used, usedErr := solver.CountUsedVariables(); usedErr != nil || used != 0 {
		t.Fatal(used, usedErr)
	}
	pivots := []SatVar{3, 1}
	if err := solver.SetPivotVariables(pivots); err != nil {
		t.Fatal(err)
	}
	pivots[0] = 0
	if got := solver.PivotVariables(); !slices.Equal(got, []SatVar{3, 1}) {
		t.Fatal(got)
	}
	result, err := solver.Solve(nil, SatSolver3Limits{GlobalPropagations: stats.Propagations})
	if err != nil || result.Status != SatUnknown || result.Reason != SatPropagationLimit {
		t.Fatal(result, err)
	}
	if err = solver.SetVariableActivity([]SatVar{2, 1}); err != nil || solver.core.activity[2] <= solver.core.activity[1] || solver.core.activity[1] <= solver.core.activity[0] {
		t.Fatal(err, solver.core.activity)
	}
	before := slices.Clone(solver.core.activity)
	if err = solver.SetVariableActivity([]SatVar{1, 1}); !errors.Is(err, ErrDimension) || !slices.Equal(before, solver.core.activity) {
		t.Fatal(err, solver.core.activity)
	}
	if err = solver.SetLiteralPolarity([]SatLit{y.Not()}); err != nil || solver.core.polarity[1] {
		t.Fatal(err, solver.core.polarity)
	}
	if err = solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	if usage, memoryErr := solver.MemoryUsage(true); memoryErr != nil || usage.FacadeStruct == 0 || usage.FacadeBookmark == 0 || usage.Total == 0 {
		t.Fatal(usage, memoryErr)
	}
	if ok, addErr := solver.AddClause([]SatLit{y}); !ok || addErr != nil {
		t.Fatal(ok, addErr)
	}
	if solver.OriginalClauseCount() != 2 || solver.Rollback() != nil || solver.OriginalClauseCount() != 1 {
		t.Fatal(solver.OriginalClauseCount())
	}
	if err = solver.Restart(11); err != nil || solver.VariableCount() != 0 || solver.OriginalClauseCount() != 0 || solver.ConflictCount() != 0 {
		t.Fatal(err, solver.VariableCount(), solver.OriginalClauseCount(), solver.ConflictCount())
	}
	if err = solver.SetVariableCount(1); err != nil {
		t.Fatal(err)
	}
	solver.Stop()
	if _, err = solver.Solve(nil, SatSolver3Limits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func satSolver3ForbidAssignment(t testing.TB, solver *SatSolver3, variables, assignment int) {
	t.Helper()
	clause := make([]SatLit, variables)
	for variable := 0; variable < variables; variable++ {
		value := assignment&(1<<(variables-variable-1)) != 0
		clause[variable] = satSolver3TestLiteral(t, variable, value)
	}
	if _, err := solver.AddClause(clause); err != nil {
		t.Fatal(assignment, err)
	}
}

func satSolver3LexWord(literals []SatLit) int {
	word := 0
	for _, literal := range literals {
		word <<= 1
		if !literal.Negative() {
			word++
		}
	}
	return word
}

func TestSatSolver3LexicographicExhaustive(t *testing.T) {
	const variables = 3
	for allowed := 0; allowed < 1<<(1<<variables); allowed++ {
		for start := 0; start < 1<<variables; start++ {
			solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: variables})
			_ = solver.SetVariableCount(variables)
			for assignment := 0; assignment < 1<<variables; assignment++ {
				if allowed&(1<<assignment) == 0 {
					satSolver3ForbidAssignment(t, solver, variables, assignment)
					if solver.core.inconsistent {
						break
					}
				}
			}
			lower := make([]SatLit, variables)
			for index := range lower {
				bit := variables - index - 1
				lower[index] = satSolver3TestLiteral(t, index, start&(1<<bit) == 0)
			}
			got, err := solver.SolveLexicographic(lower, SatSolver3Limits{})
			want := -1
			for candidate := start; candidate < 1<<variables; candidate++ {
				if allowed&(1<<candidate) != 0 {
					want = candidate
					break
				}
			}
			if err != nil {
				t.Fatal(allowed, start, err)
			}
			if want < 0 {
				if got.Result.Status != SatUnsatisfiable {
					t.Fatal(allowed, start, want, got)
				}
			} else if got.Result.Status != SatSatisfiable || satSolver3LexWord(got.Literals) != want {
				t.Fatal(allowed, start, want, got.Result.Status, got.Literals)
			}
			if !slices.Equal(lower, func() []SatLit {
				copy := make([]SatLit, variables)
				for index := range copy {
					bit := variables - index - 1
					copy[index] = satSolver3TestLiteral(t, index, start&(1<<bit) == 0)
				}
				return copy
			}()) {
				t.Fatal("input mutated")
			}
		}
	}
}

func TestSatSolver3MinimizeAssumptions(t *testing.T) {
	solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: 3})
	_ = solver.SetVariableCount(3)
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	z := satSolver3TestLiteral(t, 2, false)
	_, _ = solver.AddClause([]SatLit{x.Not(), y.Not()})
	_, _ = solver.AddClause([]SatLit{y.Not(), z.Not()})
	core, err := solver.MinimizeAssumptions([]SatLit{x, y, z}, SatSolver3Limits{})
	if err != nil || !slices.Equal(core, []SatLit{y, z}) {
		t.Fatal(core, err)
	}
	if result, solveErr := solver.Solve(core, SatSolver3Limits{}); solveErr != nil || result.Status != SatUnsatisfiable {
		t.Fatal(result, solveErr)
	}
	for index := range core {
		candidate := append(slices.Clone(core[:index]), core[index+1:]...)
		result, solveErr := solver.Solve(candidate, SatSolver3Limits{})
		if solveErr != nil || result.Status != SatSatisfiable {
			t.Fatal(index, result, solveErr)
		}
	}
	if _, err = solver.MinimizeAssumptions([]SatLit{x}, SatSolver3Limits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatSolver3MinimizeAssumptionsWithUnits(t *testing.T) {
	solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: 3})
	_ = solver.SetVariableCount(3)
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	a := satSolver3TestLiteral(t, 2, false)
	// The first two clauses entail x without making it a syntactic unit. The
	// third clause makes y false, so assumptions x,y are UNSAT and x is unused.
	_, _ = solver.AddClause([]SatLit{x, a})
	_, _ = solver.AddClause([]SatLit{x, a.Not()})
	_, _ = solver.AddClause([]SatLit{y.Not()})
	minimized, err := solver.MinimizeAssumptionsWithUnits([]SatLit{x, y}, SatSolver3Limits{})
	if err != nil || !slices.Equal(minimized.Literals, []SatLit{y}) || !slices.Equal(minimized.LearnedUnits, []SatLit{x}) {
		t.Fatal(minimized, err)
	}
	if result, solveErr := solver.Solve([]SatLit{x.Not()}, SatSolver3Limits{}); solveErr != nil || result.Status != SatUnsatisfiable {
		t.Fatal(result, solveErr)
	}
	if result, solveErr := solver.Solve(minimized.Literals, SatSolver3Limits{}); solveErr != nil || result.Status != SatUnsatisfiable {
		t.Fatal(result, solveErr)
	}

	satisfiable, _ := NewSatSolver3(SatSolver3Options{MaxVariables: 1})
	_ = satisfiable.SetVariableCount(1)
	if _, err = satisfiable.MinimizeAssumptionsWithUnits([]SatLit{x}, SatSolver3Limits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if ok, pushErr := satisfiable.Push(x); !ok || pushErr != nil {
		t.Fatal(ok, pushErr)
	}
	if _, err = satisfiable.MinimizeAssumptionsWithUnits([]SatLit{x.Not()}, SatSolver3Limits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatSolver3UtilityTraceAndStoreCompatibility(t *testing.T) {
	solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: 3})
	_ = solver.SetVariableCount(3)
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	_, _ = solver.AddClause([]SatLit{x, y})
	if ok, err := solver.Push(x.Not()); !ok || err != nil {
		t.Fatal(ok, err)
	}
	dimacs, err := solver.DIMACS(nil, SatSolverDIMACSOptions{IncrementVariables: true})
	if err != nil || !strings.Contains(string(dimacs), "-1 0\n") {
		t.Fatal(string(dimacs), err)
	}
	result, err := solver.Solve(nil, SatSolver3Limits{})
	if err != nil || result.Status != SatSatisfiable || solver.VerifyModel() != nil {
		t.Fatal(result, err)
	}
	model, err := solver.Model([]SatVar{0, 1, 2})
	if err != nil || len(model) != 3 || model[0] != 0 || model[1] != 1 {
		t.Fatal(model, err)
	}
	if !strings.Contains(solver.FormatStatistics(), "conflicts") {
		t.Fatal(solver.FormatStatistics())
	}
	if err = solver.DoubleClauses(0); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}

	traceName := t.TempDir() + "/solver3.trace"
	if err = solver.StartTrace(traceName, SatTraceLimits{}); err != nil {
		t.Fatal(err)
	}
	if err = solver.TraceWrite([]SatLit{x, y}, true); err != nil || solver.StopTrace() != nil {
		t.Fatal(err)
	}
	trace, err := os.ReadFile(traceName)
	if err != nil || !strings.HasPrefix(string(trace), "p 3 1 1") {
		t.Fatal(string(trace), err)
	}

	if err = solver.AllocateProofStore(SatProofStoreLimits{}); err != nil {
		t.Fatal(err)
	}
	if id, addErr := solver.StoreProofClause([]SatLit{x, y}); id != 0 || addErr != nil {
		t.Fatal(id, addErr)
	}
	if err = solver.MarkProofRoots(); err != nil || solver.MarkProofClausesA() != nil {
		t.Fatal(err)
	}
	storeName := t.TempDir() + "/solver3.store"
	if err = solver.WriteProofStore(storeName); err != nil {
		t.Fatal(err)
	}
	store, err := solver.ReleaseProofStore()
	if err != nil || store.Stats().Roots != 1 || store.Stats().ClausesA != 1 {
		t.Fatal(store, err)
	}
	if _, err = solver.ReleaseProofStore(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatSolver3CheckedLearnedClauseInsertion(t *testing.T) {
	solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: 2})
	_ = solver.SetVariableCount(2)
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	if ok, err := solver.AddClause([]SatLit{x, y}); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if ok, err := solver.AddLearnedClause([]SatLit{x, y}); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if solver.LearnedClauseCount() == 0 || solver.Check() != nil {
		t.Fatal(solver.LearnedClauseCount(), solver.Check())
	}
	before := solver.LearnedClauseCount()
	if ok, err := solver.AddLearnedClause([]SatLit{x, x.Not()}); !ok || err != nil || solver.LearnedClauseCount() != before {
		t.Fatal(ok, err, solver.LearnedClauseCount(), before)
	}
	if ok, err := solver.AddLearnedClause([]SatLit{x}); ok || !errors.Is(err, ErrInvariant) {
		t.Fatal(ok, err)
	}
}

func FuzzSatSolver3Lexicographic(f *testing.F) {
	f.Add([]byte{3, 2, 0, 1, 3, 2, 4, 5, 6})
	f.Add([]byte{1, 1, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := int(data[0]%4) + 1
		solver, err := NewSatSolver3(SatSolver3Options{MaxVariables: variables})
		if err != nil || solver.SetVariableCount(variables) != nil {
			t.Fatal(err)
		}
		formula := make([][]SatLit, 0)
		for offset, clauses := 1, 0; offset < len(data) && clauses < 12; clauses++ {
			width := int(data[offset]%byte(variables+1)) + 1
			offset++
			if offset+width > len(data) {
				break
			}
			clause := make([]SatLit, width)
			for index := range clause {
				clause[index] = SatLit(data[offset+index] % byte(2*variables))
			}
			offset += width
			formula = append(formula, slices.Clone(clause))
			_, addErr := solver.AddClause(clause)
			if addErr != nil {
				t.Fatal(addErr)
			}
		}
		start := int(data[len(data)-1]) & ((1 << variables) - 1)
		lower := make([]SatLit, variables)
		for index := range lower {
			lower[index] = satSolver3TestLiteral(t, index, start&(1<<(variables-index-1)) == 0)
		}
		got, solveErr := solver.SolveLexicographic(lower, SatSolver3Limits{})
		if solveErr != nil {
			t.Fatal(solveErr)
		}
		want := -1
		for word := start; word < 1<<variables; word++ {
			assignment := 0
			for index := 0; index < variables; index++ {
				if word&(1<<(variables-index-1)) != 0 {
					assignment |= 1 << index
				}
			}
			valid := true
			for _, clause := range formula {
				if !satTestEval(clause, assignment) {
					valid = false
					break
				}
			}
			if valid {
				want = word
				break
			}
		}
		if want < 0 {
			if got.Result.Status != SatUnsatisfiable {
				t.Fatal(want, got)
			}
			return
		}
		if got.Result.Status != SatSatisfiable || satSolver3LexWord(got.Literals) != want {
			t.Fatal(want, got.Result.Status, got.Literals, formula)
		}
	})
}

func FuzzSatSolver3MinimizeWithUnits(f *testing.F) {
	f.Add([]byte{3, 2, 2, 0, 4, 2, 1, 3, 3, 0, 2, 4})
	f.Add([]byte{2, 1, 1, 0, 2, 0, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%4) + 1
		solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: variables})
		_ = solver.SetVariableCount(variables)
		formula := make([][]SatLit, 0)
		offset := 1
		clauseCount := int(data[offset] % 9)
		offset++
		for clauseIndex := 0; clauseIndex < clauseCount && offset < len(data); clauseIndex++ {
			width := int(data[offset]%byte(variables+1)) + 1
			offset++
			if offset+width > len(data) {
				break
			}
			clause := make([]SatLit, width)
			for index := range clause {
				clause[index] = SatLit(data[offset+index] % byte(2*variables))
			}
			offset += width
			formula = append(formula, slices.Clone(clause))
			_, _ = solver.AddClause(clause)
		}
		assumptions := make([]SatLit, 0, 6)
		for ; offset < len(data) && len(assumptions) < 6; offset++ {
			assumptions = append(assumptions, SatLit(data[offset]%byte(2*variables)))
		}
		baseSAT, _ := satBruteForce(variables, formula, nil)
		fullSAT, _ := satBruteForce(variables, formula, assumptions)
		if !baseSAT || fullSAT || len(assumptions) == 0 {
			return
		}
		got, err := solver.MinimizeAssumptionsWithUnits(assumptions, SatSolver3Limits{})
		if err != nil {
			t.Fatal(err, formula, assumptions)
		}
		if sat, _ := satBruteForce(variables, formula, got.Literals); sat || len(got.Literals) == 0 {
			t.Fatal("invalid minimized core", got, formula, assumptions)
		}
		position := 0
		for _, literal := range got.Literals {
			for position < len(assumptions) && assumptions[position] != literal {
				position++
			}
			if position == len(assumptions) {
				t.Fatal("not an ordered subset", got.Literals, assumptions)
			}
			position++
		}
		for _, unit := range got.LearnedUnits {
			if sat, _ := satBruteForce(variables, formula, []SatLit{unit.Not()}); sat {
				t.Fatal("unit is not entailed", unit, formula)
			}
		}
	})
}

func FuzzSatSolver3PushPop(f *testing.F) {
	f.Add([]byte{3, 2, 2, 1, 3, 2, 0, 4, 0, 2, 1, 5, 3, 4})
	f.Add([]byte{1, 1, 1, 0, 0, 1, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		variables := int(data[0]%4) + 1
		solver, _ := NewSatSolver3(SatSolver3Options{MaxVariables: variables})
		_ = solver.SetVariableCount(variables)
		formula := make([][]SatLit, 0)
		offset := 1
		clauseCount := int(data[offset] % 7)
		offset++
		for clauseIndex := 0; clauseIndex < clauseCount && offset < len(data); clauseIndex++ {
			width := int(data[offset]%byte(variables+1)) + 1
			offset++
			if offset+width > len(data) {
				break
			}
			clause := make([]SatLit, width)
			for index := range clause {
				clause[index] = SatLit(data[offset+index] % byte(2*variables))
			}
			offset += width
			formula = append(formula, slices.Clone(clause))
			_, _ = solver.AddClause(clause)
		}
		stack := make([]SatLit, 0)
		for ; offset < len(data); offset++ {
			operation := data[offset]
			if operation%4 == 0 && len(stack) != 0 {
				if err := solver.Pop(); err != nil {
					t.Fatal(err)
				}
				stack = stack[:len(stack)-1]
				continue
			}
			literal := SatLit(operation % byte(2*variables))
			ok, err := solver.Push(literal)
			if err != nil {
				t.Fatal(err)
			}
			candidate := append(slices.Clone(stack), literal)
			if ok {
				stack = candidate
			} else if sat, _ := satBruteForce(variables, formula, candidate); sat {
				t.Fatal("propagation rejected satisfiable prefix", formula, candidate)
			}
			if !slices.Equal(solver.PushedAssumptions(), stack) {
				t.Fatal(solver.PushedAssumptions(), stack)
			}
			result, solveErr := solver.SolveInternal()
			if solveErr != nil {
				t.Fatal(solveErr)
			}
			wantSAT, _ := satBruteForce(variables, formula, stack)
			if (result.Status == SatSatisfiable) != wantSAT {
				t.Fatal(result.Status, wantSAT, formula, stack)
			}
			if err = solver.Check(); err != nil {
				t.Fatal(err)
			}
		}
	})
}
