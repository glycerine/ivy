package abc

// Go-only SAT models for bmcFault.c reusable campaign operations.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestAnalyzeFaultParameters(t *testing.T) {
	s := NewSatSolver()
	_ = s.SetVariableCount(3)
	_, _ = s.AddClause([]SatLit{satCardinalityLit(0, true)})
	_, _ = s.AddClause([]SatLit{satCardinalityLit(1, false), satCardinalityLit(2, false)})
	result, err := AnalyzeFaultParameters(s, []SatVar{0, 1, 2}, []bool{true, true, true}, FaultParameterAnalysisOptions{})
	if err != nil || result.Unsatisfiable != 1 || result.Runs < 2 || result.Runs > 3 || !slices.Equal(result.Active, []bool{false, true, true}) || result.ModelCovered < 1 {
		t.Fatal(result, err)
	}
	limited, err := AnalyzeFaultParameters(s, []SatVar{0, 1, 2}, []bool{true, true, true}, FaultParameterAnalysisOptions{SAT: SatLimits{Conflicts: 1}})
	if err != nil || limited.Runs == 0 {
		t.Fatal(limited, err)
	}
	if result, err := AnalyzeFaultParameters(s, []SatVar{0}, nil, FaultParameterAnalysisOptions{}); result.Active != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func TestEnumerateFaultAssignments(t *testing.T) {
	s := NewSatSolver()
	_ = s.SetVariableCount(3)
	_, err := AddSatCardinalityConstraint(s, []SatVar{0, 1, 2}, 1, true, SatCardinalityGeneral, SatCardinalityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	beforeVariables, beforeClauses := s.VariableCount(), s.OriginalClauseCount()
	result, err := EnumerateFaultAssignments(s, []SatVar{0, 1, 2}, FaultAssignmentEnumerationOptions{})
	if err != nil || result.Status != SatUnsatisfiable || result.Iterations != 3 || len(result.Assignments) != 3 {
		t.Fatal(result, err)
	}
	seen := make([]bool, 3)
	for _, assignment := range result.Assignments {
		if len(assignment) != 1 || assignment[0] < 0 || assignment[0] >= 3 || seen[assignment[0]] {
			t.Fatal(result.Assignments)
		}
		seen[assignment[0]] = true
	}
	if s.VariableCount() != beforeVariables || s.OriginalClauseCount() != beforeClauses {
		t.Fatal("enumeration mutated caller solver")
	}
	limited, err := EnumerateFaultAssignments(s, []SatVar{0, 1, 2}, FaultAssignmentEnumerationOptions{MaxIterations: 2})
	if err != nil || limited.Iterations != 2 || limited.Status != SatSatisfiable {
		t.Fatal(limited, err)
	}
}

func faultPatternMiter(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	x, _ := g.AppendCI()
	p, _ := g.AppendCI()
	root, err := g.AppendXorReal(x, p)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = g.AppendCO(root)
	return g
}

func checkAddedFaultPattern(t *testing.T, pattern, detect bool) {
	t.Helper()
	s := NewSatSolver()
	_ = s.SetVariableCount(1)
	result, err := AddFaultPatternInstance(s, faultPatternMiter(t), []bool{pattern}, 1, detect, []SatVar{0}, AddFaultPatternOptions{})
	if err != nil || result.Variables == 0 || result.Clauses == 0 || len(result.InputLiterals) != 2 || len(result.OutputLiterals) != 1 {
		t.Fatal(result, err)
	}
	wantParameter := pattern != detect
	for _, parameter := range []bool{false, true} {
		query, err := s.Solve([]SatLit{faultCNFAssumption(satCardinalityLit(0, false), parameter)}, SatLimits{})
		if err != nil || (query.Status == SatSatisfiable) != (parameter == wantParameter) {
			t.Fatal(pattern, detect, parameter, wantParameter, query.Status, err)
		}
	}
}

func TestAddFaultPatternInstance(t *testing.T) {
	for _, pattern := range []bool{false, true} {
		for _, detect := range []bool{false, true} {
			checkAddedFaultPattern(t, pattern, detect)
		}
	}
}

func TestAddFaultPatternAtomicFailures(t *testing.T) {
	miter := faultPatternMiter(t)
	for _, tc := range []struct {
		pattern    []bool
		functional int
		shared     []SatVar
		options    AddFaultPatternOptions
		want       error
	}{
		{nil, 1, []SatVar{0}, AddFaultPatternOptions{}, ErrDimension},
		{[]bool{false}, 1, nil, AddFaultPatternOptions{}, ErrDimension},
		{[]bool{false}, 1, []SatVar{1}, AddFaultPatternOptions{}, ErrDimension},
		{[]bool{false}, 1, []SatVar{0}, AddFaultPatternOptions{CNF: GiaFaultCNFOptions{MaxVariables: 1}}, ErrLimit},
		{[]bool{false}, 1, []SatVar{0}, AddFaultPatternOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		s := NewSatSolver()
		_ = s.SetVariableCount(1)
		before := s.cloneState()
		result, err := AddFaultPatternInstance(s, miter, tc.pattern, tc.functional, true, tc.shared, tc.options)
		if !errors.Is(err, tc.want) || result.InputLiterals != nil || !reflect.DeepEqual(before, *s) {
			t.Fatal(tc, result, err, reflect.DeepEqual(before, *s))
		}
	}
}

func FuzzAddFaultPatternInstance(f *testing.F) {
	for _, seed := range []byte{0, 1, 2, 3} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, flags byte) {
		pattern, detect := flags&1 != 0, flags&2 != 0
		s := NewSatSolver()
		_ = s.SetVariableCount(1)
		_, err := AddFaultPatternInstance(s, faultPatternMiter(t), []bool{pattern}, 1, detect, []SatVar{0}, AddFaultPatternOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := pattern != detect
		for _, value := range []bool{false, true} {
			query, err := s.Solve([]SatLit{faultCNFAssumption(0, value)}, SatLimits{})
			if err != nil || (query.Status == SatSatisfiable) != (value == want) {
				t.Fatal(pattern, detect, value, want, query.Status, err)
			}
		}
	})
}
