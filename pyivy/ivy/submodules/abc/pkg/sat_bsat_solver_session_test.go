package abc

// Stand-alone coverage for behavior unique to src/sat/bsat/satSolver.c.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func TestSatSolverSessionPoliciesAndActivityFactors(t *testing.T) {
	session, err := NewSatSolverSession(SatSolverSessionOptions{MaxVariables: 4})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	if err = session.SetLearnedReductionPolicy(17, 3, 75); err != nil || session.core.learntLimit != 17 || session.core.learntStart != 17 || session.core.learntDelta != 3 || session.core.learntRatio != 75 {
		t.Fatal(err, session.core.learntLimit, session.core.learntStart, session.core.learntDelta, session.core.learntRatio)
	}
	beforePolicy := [...]int{session.core.learntLimit, session.core.learntStart, session.core.learntDelta, session.core.learntRatio}
	if err = session.SetLearnedReductionPolicy(1, -1, 50); !errors.Is(err, ErrDimension) || beforePolicy != [...]int{session.core.learntLimit, session.core.learntStart, session.core.learntDelta, session.core.learntRatio} {
		t.Fatal(err, session.core.learntLimit, session.core.learntStart, session.core.learntDelta, session.core.learntRatio)
	}
	if err = session.SetRestartInterval(0); err != nil || session.core.restartInterval != 0 {
		t.Fatal(err, session.core.restartInterval)
	}
	if err = session.SetRestartInterval(9); err != nil || session.core.restartInterval != 9 {
		t.Fatal(err, session.core.restartInterval)
	}

	factors := []float64{0, 2}
	variables := []SatVar{1}
	if err = session.SetActivityFactors(factors, variables); err != nil {
		t.Fatal(err)
	}
	factors[1], variables[0] = 99, 0
	if !slices.Equal(session.core.activityFactors, []float64{0, 2}) || !slices.Equal(session.core.activityFactorVariables, []SatVar{1}) {
		t.Fatal("activity configuration aliases caller storage")
	}
	baseline := slices.Clone(session.core.activity)
	for solve := 0; solve < 3; solve++ {
		result, solveErr := session.SolveInternal()
		if solveErr != nil || result.Status != SatSatisfiable {
			t.Fatal(solve, result, solveErr)
		}
	}
	if got, want := session.core.activity[1], baseline[1]+2*session.core.varInc*2; got != want || session.core.activity[0] != baseline[0] {
		t.Fatal("odd-search factor cadence", session.core.activity, baseline, session.core.varInc)
	}
	beforeFactors := slices.Clone(session.core.activityFactors)
	beforeVariables := slices.Clone(session.core.activityFactorVariables)
	invalid := []struct {
		factors   []float64
		variables []SatVar
	}{
		{[]float64{1}, []SatVar{0}},
		{[]float64{0, math.NaN()}, []SatVar{1}},
		{[]float64{0, -1}, []SatVar{1}},
		{[]float64{0, 1}, []SatVar{1, 1}},
		{[]float64{0, 1}, []SatVar{2}},
	}
	for index, input := range invalid {
		if err = session.SetActivityFactors(input.factors, input.variables); !errors.Is(err, ErrDimension) || !slices.Equal(beforeFactors, session.core.activityFactors) || !slices.Equal(beforeVariables, session.core.activityFactorVariables) {
			t.Fatal(index, err, session.core.activityFactors, session.core.activityFactorVariables)
		}
	}
	if err = session.ClearActivityFactors(); err != nil || session.core.activityFactors != nil || session.core.activityFactorVariables != nil {
		t.Fatal(err, session.core.activityFactors, session.core.activityFactorVariables)
	}

	biased, _ := NewSatSolverSession(SatSolverSessionOptions{MaxVariables: 2})
	_ = biased.SetVariableCount(2)
	if err = biased.SetActivityFactors([]float64{0, 0}, []SatVar{1}); err != nil {
		t.Fatal(err)
	}
	globals := []SatVar{1}
	if err = biased.SetGlobalActivityVariables(globals); err != nil {
		t.Fatal(err)
	}
	globals[0] = 0
	globalBaseline := slices.Clone(biased.core.activity)
	for solve := 0; solve < 2; solve++ {
		if result, solveErr := biased.SolveInternal(); solveErr != nil || result.Status != SatSatisfiable {
			t.Fatal(solve, result, solveErr)
		}
	}
	if got, want := biased.core.activity[1], globalBaseline[1]+2*3*biased.core.varInc; got != want || biased.core.activity[0] != globalBaseline[0] {
		t.Fatal("global factor cadence", biased.core.activity, globalBaseline)
	}
	beforeGlobal := slices.Clone(biased.core.globalActivityVariables)
	if err = biased.SetGlobalActivityVariables([]SatVar{1, 1}); !errors.Is(err, ErrDimension) || !slices.Equal(beforeGlobal, biased.core.globalActivityVariables) {
		t.Fatal(err, biased.core.globalActivityVariables)
	}
	if err = biased.SetGlobalActivityVariables(nil); err != nil || biased.core.globalActivityVariables != nil {
		t.Fatal(err, biased.core.globalActivityVariables)
	}
	if ok, compressErr := biased.Compress(); compressErr != nil || !ok {
		t.Fatal(ok, compressErr)
	}
}

func TestSatSolverSessionVariablesProjectionAndRunStop(t *testing.T) {
	session, err := NewSatSolverSession(SatSolverSessionOptions{MaxVariables: 4})
	if err != nil {
		t.Fatal(err)
	}
	first, err := session.AddVariable()
	if err != nil || first != 0 {
		t.Fatal(first, err)
	}
	second, err := session.AddVariable()
	if err != nil || second != 1 {
		t.Fatal(second, err)
	}
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	_, _ = session.AddClause([]SatLit{x})
	_, _ = session.AddClause([]SatLit{y.Not()})
	if err = session.SetModelProjection([]SatVar{1, 0}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	if err = session.SetRunStop(42, func(id int) bool {
		calls++
		if id != 42 {
			t.Fatalf("run ID %d", id)
		}
		return true
	}); err != nil {
		t.Fatal(err)
	}
	result, err := session.SolveInternal()
	if err != nil || result.Status != SatUnknown || result.Reason != SatCancelled || calls == 0 || len(session.ModelProjectionValues()) != 0 {
		t.Fatal(result, err, calls, session.ModelProjectionValues())
	}
	if err = session.SetRunStop(0, nil); err != nil {
		t.Fatal(err)
	}
	result, err = session.SolveInternal()
	if err != nil || result.Status != SatSatisfiable || !slices.Equal(session.ModelProjectionValues(), []int{0, 1}) {
		t.Fatal(result, err, session.ModelProjectionValues())
	}
	result, err = session.SolveInternal()
	if err != nil || result.Status != SatSatisfiable || !slices.Equal(session.ModelProjectionValues(), []int{0, 1, 0, 1}) {
		t.Fatal(result, err, session.ModelProjectionValues())
	}
	if err = session.ClearModelProjectionValues(); err != nil || len(session.ModelProjectionValues()) != 0 {
		t.Fatal(err, session.ModelProjectionValues())
	}
	projection := []SatVar{0, 9}
	if err = session.SetModelProjection(projection); !errors.Is(err, ErrInvalidLiteral) || len(session.userVariables) != 2 {
		t.Fatal(err, session.userVariables)
	}
}

func TestSatSolverSessionClauseDiagnosticsAndProofStoreChange(t *testing.T) {
	session, _ := NewSatSolverSession(SatSolverSessionOptions{MaxVariables: 3})
	_ = session.SetVariableCount(3)
	x := satSolver3TestLiteral(t, 0, false)
	y := satSolver3TestLiteral(t, 1, false)
	if enabled, err := session.ToggleClauseDiagnostics(); err != nil || !enabled {
		t.Fatal(enabled, err)
	}
	clause := []SatLit{x, y}
	if ok, err := session.AddClause(clause); !ok || err != nil {
		t.Fatal(ok, err)
	}
	clause[0] = x.Not()
	if got := session.ClauseDiagnostics(); len(got) != 1 || !slices.Equal(got[0], []SatLit{x, y}) {
		t.Fatal(got)
	}
	gates, _ := SatSolver3BufferClauses(1, 2, false)
	if ok, err := session.AddClauses(gates); !ok || err != nil || len(session.ClauseDiagnostics()) != 3 {
		t.Fatal(ok, err, session.ClauseDiagnostics())
	}
	if err := session.ClearClauseDiagnostics(); err != nil || len(session.ClauseDiagnostics()) != 0 {
		t.Fatal(err)
	}
	if enabled, err := session.ToggleClauseDiagnostics(); err != nil || enabled {
		t.Fatal(enabled, err)
	}

	if err := session.AllocateProofStore(SatProofStoreLimits{}); err != nil {
		t.Fatal(err)
	}
	if _, err := session.StoreProofClause([]SatLit{x}); err != nil {
		t.Fatal(err)
	}
	if err := session.MarkProofRoots(); err != nil || session.MarkProofClausesA() != nil {
		t.Fatal(err)
	}
	variable, err := session.ChangeLastProofStoreClause()
	if err != nil || variable != 0 || session.store.Stats().ClausesA != 0 {
		t.Fatal(variable, err, session.store.Stats())
	}
	if _, err = session.ChangeLastProofStoreClause(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatSolverSessionRestartRetainsCallerConfiguration(t *testing.T) {
	session, _ := NewSatSolverSession(SatSolverSessionOptions{MaxVariables: 2})
	_ = session.SetVariableCount(2)
	_ = session.SetModelProjection([]SatVar{1})
	_, _ = session.ToggleClauseDiagnostics()
	if err := session.Restart(17); err != nil || session.VariableCount() != 0 || !session.printClauses || !slices.Equal(session.userVariables, []SatVar{1}) {
		t.Fatal(err, session.VariableCount(), session.printClauses, session.userVariables)
	}
	if _, err := session.SolveInternal(); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if err := session.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	if result, err := session.SolveInternal(); err != nil || result.Status != SatSatisfiable || len(session.ModelProjectionValues()) != 1 {
		t.Fatal(result, err, session.ModelProjectionValues())
	}
}

func TestSatSolverSessionAutomaticProofStore(t *testing.T) {
	session, _ := NewSatSolverSession(SatSolverSessionOptions{MaxVariables: 2})
	_ = session.SetVariableCount(2)
	if err := session.AllocateProofStore(SatProofStoreLimits{}); err != nil {
		t.Fatal(err)
	}
	x := satSolver3TestLiteral(t, 0, false)
	if ok, err := session.AddClause([]SatLit{x}); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if stats := session.store.Stats(); stats.Clauses != 1 || stats.EmptyClause != -1 {
		t.Fatal(stats)
	}
	result, err := session.Solve([]SatLit{x.Not()}, SatSolverSessionLimits{})
	if err != nil || result.Status != SatUnsatisfiable {
		t.Fatal(result, err)
	}
	if stats := session.store.Stats(); stats.Clauses != 2 || stats.EmptyClause != 1 {
		t.Fatal(stats)
	}
	result, err = session.Solve([]SatLit{x.Not()}, SatSolverSessionLimits{})
	if err != nil || result.Status != SatUnsatisfiable || session.store.Stats().Clauses != 2 {
		t.Fatal(result, err, session.store.Stats())
	}

	before := session.OriginalClauseCount()
	if ok, err := session.AddClause([]SatLit{x, x.Not()}); ok || !errors.Is(err, ErrInvariant) || session.OriginalClauseCount() != before {
		t.Fatal(ok, err, session.OriginalClauseCount(), before)
	}
}

func FuzzSatSolverSessionActivityFactors(f *testing.F) {
	for _, seed := range [][4]byte{{1, 2, 3, 0}, {7, 0xff, 5, 0x55}, {8, 0x0f, 0, 0xf0}} {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}
	f.Fuzz(func(t *testing.T, rawSearches, activeMask, rawFactor, globalMask byte) {
		const count = 8
		searches := int(rawSearches%8) + 1
		factor := uint64(rawFactor % 17)
		session, err := NewSatSolverSession(SatSolverSessionOptions{MaxVariables: count})
		if err != nil {
			t.Fatal(err)
		}
		if err = session.SetVariableCount(count); err != nil {
			t.Fatal(err)
		}
		factors := make([]float64, count)
		active := make([]SatVar, 0, count)
		globals := make([]SatVar, 0, count)
		for variable := 0; variable < count; variable++ {
			factors[variable] = float64(factor)
			if activeMask&(1<<variable) != 0 {
				active = append(active, SatVar(variable))
			}
			if globalMask&(1<<variable) != 0 {
				globals = append(globals, SatVar(variable))
			}
		}
		if err = session.SetActivityFactors(factors, active); err != nil {
			t.Fatal(err)
		}
		if err = session.SetGlobalActivityVariables(globals); err != nil {
			t.Fatal(err)
		}
		baseline := slices.Clone(session.core.activity)
		increment := session.core.varInc
		for search := 0; search < searches; search++ {
			result, solveErr := session.SolveInternal()
			if solveErr != nil || result.Status != SatSatisfiable {
				t.Fatal(search, result, solveErr)
			}
		}
		oddSearches := uint64((searches + 1) / 2)
		for variable := 0; variable < count; variable++ {
			want := baseline[variable]
			if activeMask&(1<<variable) != 0 {
				want += oddSearches * increment * factor
				if globalMask&(1<<variable) != 0 {
					want += uint64(searches) * increment * 3
				}
			}
			if session.core.activity[variable] != want {
				t.Fatalf("variable %d activity=%d want=%d", variable, session.core.activity[variable], want)
			}
		}
	})
}
