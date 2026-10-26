package abc

// Go-only exhaustive and fuzz tests for bmcFault.c cardinality constraints.

import (
	"errors"
	"math/bits"
	"reflect"
	"slices"
	"testing"
)

func satCardinalityAssumptions(t *testing.T, n, mask int) []SatLit {
	t.Helper()
	result := make([]SatLit, n)
	for i := range result {
		literal, err := MakeSatLit(SatVar(i), mask&(1<<i) == 0)
		if err != nil {
			t.Fatal(err)
		}
		result[i] = literal
	}
	return result
}

func checkSatCardinalityCase(t *testing.T, encoding SatCardinalityEncoding, n, k int, exact bool) {
	t.Helper()
	solver := NewSatSolver()
	if err := solver.SetVariableCount(n); err != nil {
		t.Fatal(err)
	}
	inputs := make([]SatVar, n)
	for i := range inputs {
		inputs[i] = SatVar(i)
	}
	result, err := AddSatCardinalityConstraint(solver, inputs, k, exact, encoding, SatCardinalityOptions{})
	if err != nil {
		t.Fatal(encoding, n, k, exact, err)
	}
	inputs[0] = SatVar(n + 100)
	if result.Inputs[0] != 0 || result.Encoding != encoding || result.AuxiliaryVariables != solver.VariableCount()-n || result.Literals < result.Clauses {
		t.Fatalf("bad result: %+v vars=%d", result, solver.VariableCount())
	}
	for mask := 0; mask < 1<<n; mask++ {
		count := bits.OnesCount(uint(mask))
		want := count <= k
		if exact {
			want = count == k
		}
		solved, err := solver.Solve(satCardinalityAssumptions(t, n, mask), SatLimits{})
		if err != nil || (solved.Status == SatSatisfiable) != want {
			t.Fatalf("encoding=%d n=%d k=%d exact=%v mask=%b count=%d got=%d want=%v err=%v", encoding, n, k, exact, mask, count, solved.Status, want, err)
		}
	}
}

func TestSatCardinalitySortingSemantics(t *testing.T) {
	for _, encoding := range []SatCardinalityEncoding{SatCardinalityPairwise, SatCardinalityGeneral} {
		for n := 1; n <= 8; n++ {
			for k := 0; k <= n; k++ {
				checkSatCardinalityCase(t, encoding, n, k, false)
				checkSatCardinalityCase(t, encoding, n, k, true)
			}
		}
	}
}

func TestSatCardinalityAtMostOneTree(t *testing.T) {
	for n := 1; n <= 9; n++ {
		checkSatCardinalityCase(t, SatCardinalityAtMostOneTree, n, 1, false)
		checkSatCardinalityCase(t, SatCardinalityAtMostOneTree, n, 1, true)
	}
}

func TestSatCardinalityDeterminismAndExistingFormula(t *testing.T) {
	build := func() (*SatSolver, SatCardinalityResult) {
		s := NewSatSolver()
		_ = s.SetVariableCount(4)
		ok, err := s.AddClause([]SatLit{satCardinalityLit(0, false), satCardinalityLit(1, false)})
		if err != nil || !ok {
			t.Fatal(err, ok)
		}
		r, err := AddSatCardinalityConstraint(s, []SatVar{0, 1, 2, 3}, 2, true, SatCardinalityPairwise, SatCardinalityOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return s, r
	}
	a, ar := build()
	b, br := build()
	if !reflect.DeepEqual(ar, br) || a.VariableCount() != b.VariableCount() || a.OriginalClauseCount() != b.OriginalClauseCount() {
		t.Fatal("nondeterministic construction", ar, br)
	}
	for mask := 0; mask < 16; mask++ {
		result, err := a.Solve(satCardinalityAssumptions(t, 4, mask), SatLimits{})
		want := bits.OnesCount(uint(mask)) == 2 && mask&3 != 0
		if err != nil || (result.Status == SatSatisfiable) != want {
			t.Fatal(mask, result.Status, want, err)
		}
	}
}

func TestSatCardinalityOversizedCallerBudget(t *testing.T) {
	s, err := NewSatSolverWithLimit(16)
	if err != nil || s.SetVariableCount(2) != nil {
		t.Fatal(err)
	}
	result, err := AddSatCardinalityConstraint(s, []SatVar{0, 1}, 1, false, SatCardinalityPairwise, SatCardinalityOptions{MaxVariables: 1 << 20})
	if err != nil || result.AuxiliaryVariables != 2 || s.VariableCount() != 4 {
		t.Fatal(result, s.VariableCount(), err)
	}
}

func TestSatCardinalityAtomicFailures(t *testing.T) {
	newSolver := func() *SatSolver {
		s := NewSatSolver()
		_ = s.SetVariableCount(4)
		_, _ = s.AddClause([]SatLit{satCardinalityLit(0, false), satCardinalityLit(1, false)})
		return s
	}
	inputs := []SatVar{0, 1, 2, 3}
	for _, tc := range []struct {
		variables []SatVar
		k         int
		encoding  SatCardinalityEncoding
		options   SatCardinalityOptions
		want      error
	}{
		{nil, 0, SatCardinalityPairwise, SatCardinalityOptions{}, ErrDimension},
		{inputs, -1, SatCardinalityPairwise, SatCardinalityOptions{}, ErrDimension},
		{inputs, 5, SatCardinalityPairwise, SatCardinalityOptions{}, ErrDimension},
		{[]SatVar{0, 0}, 1, SatCardinalityPairwise, SatCardinalityOptions{}, ErrDimension},
		{[]SatVar{0, 4}, 1, SatCardinalityPairwise, SatCardinalityOptions{}, ErrDimension},
		{inputs, 2, 99, SatCardinalityOptions{}, ErrDimension},
		{inputs, 2, SatCardinalityAtMostOneTree, SatCardinalityOptions{}, ErrDimension},
		{inputs, 2, SatCardinalityPairwise, SatCardinalityOptions{MaxVariables: 4}, ErrLimit},
		{inputs, 2, SatCardinalityPairwise, SatCardinalityOptions{MaxClauses: 1}, ErrLimit},
		{inputs, 2, SatCardinalityPairwise, SatCardinalityOptions{MaxLiterals: 1}, ErrLimit},
		{inputs, 2, SatCardinalityPairwise, SatCardinalityOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		s := newSolver()
		before := s.cloneState()
		if result, err := AddSatCardinalityConstraint(s, tc.variables, tc.k, false, tc.encoding, tc.options); !errors.Is(err, tc.want) || result.Inputs != nil || !reflect.DeepEqual(before, *s) {
			t.Fatalf("case %+v result=%+v err=%v atomic=%v", tc, result, err, reflect.DeepEqual(before, *s))
		}
	}
	calls := 0
	s := newSolver()
	before := s.cloneState()
	_, err := AddSatCardinalityConstraint(s, inputs, 2, false, SatCardinalityGeneral, SatCardinalityOptions{Stop: func() bool { calls++; return calls == 7 }})
	if !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(before, *s) {
		t.Fatal("mid-build cancellation was not atomic", calls, err)
	}
}

func TestSatCardinalityContradictionAndBookmark(t *testing.T) {
	s := NewSatSolver()
	_ = s.SetVariableCount(2)
	_, _ = s.AddClause([]SatLit{satCardinalityLit(0, false)})
	_, _ = s.AddClause([]SatLit{satCardinalityLit(1, false)})
	if err := s.Bookmark(); err != nil {
		t.Fatal(err)
	}
	result, err := AddSatCardinalityConstraint(s, []SatVar{0, 1}, 1, false, SatCardinalityGeneral, SatCardinalityOptions{})
	if err != nil || !result.Inconsistent || s.bookmark == nil {
		t.Fatal(result, err)
	}
	if err = s.Rollback(); err != nil {
		t.Fatal(err)
	}
	solved, err := s.Solve(nil, SatLimits{})
	if err != nil || solved.Status != SatSatisfiable {
		t.Fatal(solved.Status, err)
	}
}

func TestSatCardinalityResultOwnership(t *testing.T) {
	s := NewSatSolver()
	_ = s.SetVariableCount(3)
	inputs := []SatVar{0, 1, 2}
	r, err := AddSatCardinalityConstraint(s, inputs, 2, false, SatCardinalityGeneral, SatCardinalityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	copyInputs, copyOutputs := slices.Clone(r.Inputs), slices.Clone(r.Outputs)
	inputs[0] = 2
	r.Inputs[1], r.Outputs[0] = 2, 0
	if !slices.Equal(copyInputs, []SatVar{0, 1, 2}) || len(copyOutputs) != 3 {
		t.Fatal(copyInputs, copyOutputs)
	}
	checkSatCardinalityCase(t, SatCardinalityGeneral, 3, 2, false)
}

func FuzzSatCardinalitySemantics(f *testing.F) {
	for _, seed := range [][4]byte{{2, 0, 1, 0}, {5, 31, 2, 1}, {7, 85, 7, 0}, {8, 255, 3, 1}} {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}
	f.Fuzz(func(t *testing.T, rawN, rawMask, rawK, rawFlags byte) {
		n := int(rawN%7) + 1
		mask := int(rawMask) & ((1 << n) - 1)
		k := int(rawK) % (n + 1)
		exact := rawFlags&1 != 0
		encoding := SatCardinalityEncoding((rawFlags >> 1) & 1)
		s := NewSatSolver()
		_ = s.SetVariableCount(n)
		inputs := make([]SatVar, n)
		for i := range inputs {
			inputs[i] = SatVar(i)
		}
		if _, err := AddSatCardinalityConstraint(s, inputs, k, exact, encoding, SatCardinalityOptions{}); err != nil {
			t.Fatal(err)
		}
		result, err := s.Solve(satCardinalityAssumptions(t, n, mask), SatLimits{})
		count := bits.OnesCount(uint(mask))
		want := count <= k
		if exact {
			want = count == k
		}
		if err != nil || (result.Status == SatSatisfiable) != want {
			t.Fatal(n, mask, k, exact, encoding, result.Status, want, err)
		}
	})
}
