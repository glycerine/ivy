package abc

// Go-only tests for src/sat/proof/pr.c and pr.h.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func proofLit(t testing.TB, v int, negative bool) SatLit {
	t.Helper()
	l, e := MakeSatLit(SatVar(v), negative)
	if e != nil {
		t.Fatal(e)
	}
	return l
}

func proofContradiction(t testing.TB, swap bool) *SatProofRecorder {
	t.Helper()
	r, e := NewSatProofRecorder(2, SatProofRecordingLimits{})
	if e != nil {
		t.Fatal(e)
	}
	x, y := proofLit(t, 0, false), proofLit(t, 1, false)
	roots := [][]SatLit{{x, y}, {x.Not(), y}, {x, y.Not()}, {x.Not(), y.Not()}}
	if swap {
		slices.Reverse(roots)
	}
	for _, c := range roots {
		if _, e = r.AddClause(c, true, false); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = r.AddClause([]SatLit{x}, false, false); e != nil {
		t.Fatal(e)
	}
	if _, e = r.AddClause(nil, false, false); e != nil {
		t.Fatal(e)
	}
	return r
}

// proofHypercube builds a family that exercises successively shorter learned
// clauses.  Its roots contain every polarity vector; resolving away the last
// variable at each level eventually proves a unit and then the empty clause.
func proofHypercube(t testing.TB, variables int, reverseRoots bool) *SatProofRecorder {
	t.Helper()
	if variables < 2 || variables > 6 {
		t.Fatal("unsupported hypercube dimension", variables)
	}
	r, e := NewSatProofRecorder(variables, SatProofRecordingLimits{})
	if e != nil {
		t.Fatal(e)
	}
	roots := make([][]SatLit, 1<<variables)
	for mask := range roots {
		roots[mask] = make([]SatLit, variables)
		for v := 0; v < variables; v++ {
			roots[mask][v] = proofLit(t, v, mask&(1<<v) != 0)
		}
	}
	if reverseRoots {
		slices.Reverse(roots)
	}
	for _, clause := range roots {
		if _, e = r.AddClause(clause, true, false); e != nil {
			t.Fatal(e)
		}
	}
	for width := variables - 1; width >= 1; width-- {
		count := 1 << width
		if width == 1 {
			count = 1 // the proved unit immediately forces the final conflict
		}
		for mask := 0; mask < count; mask++ {
			clause := make([]SatLit, width)
			for v := 0; v < width; v++ {
				clause[v] = proofLit(t, v, mask&(1<<v) != 0)
			}
			if _, e = r.AddClause(clause, false, false); e != nil {
				t.Fatal(e)
			}
		}
	}
	if _, e = r.AddClause(nil, false, false); e != nil {
		t.Fatal(e)
	}
	return r
}

// proofModelResolve deliberately does not call satProofResolve.  It is a
// small, test-only checker built around a per-variable polarity table so proof
// replay does not share the implementation under test.
func proofModelResolve(a, b []SatLit, pivot int) ([]SatLit, error) {
	maxVar := pivot
	for _, clause := range [][]SatLit{a, b} {
		for _, literal := range clause {
			if int(literal.Var()) > maxVar {
				maxVar = int(literal.Var())
			}
		}
	}
	polarities := make([]int8, maxVar+1)
	for i := range polarities {
		polarities[i] = -1
	}
	for _, clause := range [][]SatLit{a, b} {
		for _, literal := range clause {
			v := int(literal.Var())
			if v == pivot {
				continue
			}
			polarity := int8(0)
			if literal.Negative() {
				polarity = 1
			}
			if polarities[v] >= 0 && polarities[v] != polarity {
				return nil, ErrInvariant
			}
			polarities[v] = polarity
		}
	}
	result := make([]SatLit, 0, len(a)+len(b)-2)
	for v, polarity := range polarities {
		if v != pivot && polarity >= 0 {
			literal, e := MakeSatLit(SatVar(v), polarity == 1)
			if e != nil {
				return nil, e
			}
			result = append(result, literal)
		}
	}
	return result, nil
}

func replayProofResult(result *SatProofRecordingResult) error {
	byID := make(map[int][]SatLit)
	for _, c := range result.Clauses {
		if c.Root {
			byID[c.ProofID] = slices.Clone(c.Literals)
		}
	}
	for _, s := range result.Steps {
		a, ok := byID[s.Left]
		if !ok {
			return ErrInvariant
		}
		b, ok := byID[s.Right]
		if !ok {
			return ErrInvariant
		}
		var pivot int = -1
		for _, x := range a {
			for _, y := range b {
				if x == y.Not() {
					if pivot >= 0 {
						return ErrInvariant
					}
					pivot = int(x.Var())
				}
			}
		}
		if pivot < 0 {
			return ErrInvariant
		}
		r, e := proofModelResolve(a, b, pivot)
		if e != nil {
			return e
		}
		byID[s.ID] = r
	}
	if c, ok := byID[result.FinalProofID]; !ok || len(c) != 0 {
		return ErrInvariant
	}
	return nil
}

func TestSatProofRecordingContradictionReplayAndText(t *testing.T) {
	r := proofContradiction(t, false)
	result, e := r.Record()
	if e != nil {
		t.Fatal(e)
	}
	if !result.DerivedEmpty || result.FinalProofID == 0 || len(result.Steps) == 0 {
		t.Fatalf("%+v", result)
	}
	if len(result.GlobalVars) != 0 || result.Interpolant != 0 {
		t.Fatalf("unexpected interpolation without A clauses: %v %08x", result.GlobalVars, result.Interpolant)
	}
	if result.Interpolant != 0 || len(result.GlobalVars) != 0 {
		t.Fatalf("interpolation should be disabled without A roots: %+v", result)
	}
	if e = replayProofResult(result); e != nil {
		t.Fatal(e)
	}
	text, e := result.MarshalText(4096)
	if e != nil || !strings.Contains(string(text), " * ") {
		t.Fatal(string(text), e)
	}
	before := slices.Clone(result.Clauses[0].Literals)
	text[0] ^= 1
	if !slices.Equal(before, result.Clauses[0].Literals) {
		t.Fatal("ownership")
	}
	corrupt := *result
	corrupt.Steps = slices.Clone(result.Steps)
	corrupt.Steps[0].Left = 1 << 30
	if e = replayProofResult(&corrupt); !errors.Is(e, ErrInvariant) {
		t.Fatal("corrupt antecedent accepted", e)
	}
}

func TestSatProofRecordingInputAndResultOwnership(t *testing.T) {
	r, e := NewSatProofRecorder(2, SatProofRecordingLimits{})
	if e != nil {
		t.Fatal(e)
	}
	x, y := proofLit(t, 0, false), proofLit(t, 1, false)
	input := []SatLit{x, y}
	if _, e = r.AddClause(input, true, false); e != nil {
		t.Fatal(e)
	}
	input[0] = x.Not()
	got, e := r.Clauses()
	if e != nil || !slices.Equal(got[0].Literals, []SatLit{x, y}) {
		t.Fatal("recorder retained caller storage", got, e)
	}
	got[0].Literals[0] = x.Not()
	again, _ := r.Clauses()
	if !slices.Equal(again[0].Literals, []SatLit{x, y}) {
		t.Fatal("Clauses exposed recorder storage")
	}
}

func TestSatProofRecordingWatchedLiteralMovement(t *testing.T) {
	x, y, z := proofLit(t, 0, false), proofLit(t, 1, false), proofLit(t, 2, false)
	clauses := []SatProofRecordingClause{{Literals: []SatLit{x, y, z}}}
	b := newSatProofBCP(3, clauses)
	if e := b.watchClause(0); e != nil {
		t.Fatal(e)
	}
	if !b.enqueue(x.Not(), -1) {
		t.Fatal("enqueue")
	}
	if conflict, e := b.propagate(0, nil); e != nil || conflict >= 0 {
		t.Fatal(conflict, e)
	}
	if b.watches[0] != [2]int{2, 1} {
		t.Fatal("watch did not move", b.watches[0])
	}
	if !b.enqueue(y.Not(), -1) {
		t.Fatal("enqueue")
	}
	if conflict, e := b.propagate(1, nil); e != nil || conflict >= 0 || !satProofLitTrue(z, b.assigns) || b.reasons[z.Var()] != 0 {
		t.Fatal("unit propagation", conflict, e, b.assigns, b.reasons)
	}
}

func TestSatProofRecordingParserStatsAndLifecycle(t *testing.T) {
	data := []byte("c demo\np 2 6 4\n1 2 0\n-1 2 0\n1 -2 0\n-1 -2 0\n1 0\n0\n")
	r, e := ParseSatProofRecording(data, SatProofRecordingLimits{})
	if e != nil {
		t.Fatal(e)
	}
	s, e := r.Stats()
	if e != nil || s.Variables != 2 || s.Roots != 4 || s.Learned != 2 || s.Clauses != 6 {
		t.Fatal(s, e)
	}
	result, e := r.Record()
	if e != nil || replayProofResult(result) != nil {
		t.Fatal(e)
	}
	r.Close()
	r.Close()
	if _, e = r.Stats(); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "proof.cnf")
	if e = os.WriteFile(path, data, 0o600); e != nil {
		t.Fatal(e)
	}
	fromFile, e := ReadSatProofRecordingFile(path, SatProofRecordingLimits{})
	if e != nil {
		t.Fatal(e)
	}
	if stats, _ := fromFile.Stats(); stats.Clauses != 6 {
		t.Fatal(stats)
	}
	if _, e = ReadSatProofRecordingFile(path, SatProofRecordingLimits{MaxBytes: 1}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
}

func TestSatProofRecordingInterpolant(t *testing.T) {
	r, e := NewSatProofRecorder(2, SatProofRecordingLimits{})
	if e != nil {
		t.Fatal(e)
	}
	x, y := proofLit(t, 0, false), proofLit(t, 1, false)
	// A: x OR y, x OR !y. B: !x OR y, !x OR !y. Their interpolant is x.
	for _, c := range []struct {
		l []SatLit
		a bool
	}{{[]SatLit{x, y}, true}, {[]SatLit{x, y.Not()}, true}, {[]SatLit{x.Not(), y}, false}, {[]SatLit{x.Not(), y.Not()}, false}} {
		if _, e = r.AddClause(c.l, true, c.a); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = r.AddClause([]SatLit{x}, false, false); e != nil {
		t.Fatal(e)
	}
	if _, e = r.AddClause(nil, false, false); e != nil {
		t.Fatal(e)
	}
	res, e := r.Record()
	if e != nil {
		t.Fatal(e)
	}
	if !slices.Equal(res.GlobalVars, []SatVar{0, 1}) || res.Interpolant != 0xAAAAAAAA {
		t.Fatalf("globals=%v truth=%08x", res.GlobalVars, res.Interpolant)
	}
}

func TestSatProofRecordingErrorsLimitsCancellationAtomicity(t *testing.T) {
	x := proofLit(t, 0, false)
	if _, e := NewSatProofRecorder(-1, SatProofRecordingLimits{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	r, _ := NewSatProofRecorder(1, SatProofRecordingLimits{MaxClauses: 2, MaxLiterals: 2})
	if _, e := r.AddClause([]SatLit{x, x.Not()}, true, false); !errors.Is(e, ErrInvalidLiteral) {
		t.Fatal(e)
	}
	if _, e := satProofResolve([]SatLit{x}, []SatLit{x}, 0); !errors.Is(e, ErrInvariant) {
		t.Fatal("same-polarity resolution accepted", e)
	}
	if s, _ := r.Stats(); s.Clauses != 0 {
		t.Fatal(s)
	}
	if _, e := r.AddClause([]SatLit{x}, false, true); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	_, _ = r.AddClause([]SatLit{x}, true, false)
	if _, e := r.AddClause([]SatLit{x.Not(), proofLit(t, 1, false)}, false, false); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	for _, bad := range []string{"", "1 0\n", "p 1 2 1\n1 0\n", "p 1 1 2\n0\n", "p x 1 0\n0\n", "p 1 1 0\n1\n"} {
		if _, e := ParseSatProofRecording([]byte(bad), SatProofRecordingLimits{}); e == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	if _, e := ParseSatProofRecording([]byte("p 1 1 0\n0\n"), SatProofRecordingLimits{MaxBytes: 1}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if _, e := NewSatProofRecorder(0, SatProofRecordingLimits{Stop: func() bool { return true }}); !errors.Is(e, ErrCancelled) {
		t.Fatal(e)
	}
	limited := proofContradiction(t, false)
	limited.limits.MaxSteps = 1
	if result, e := limited.Record(); result != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(result, e)
	}
	cancelled := proofContradiction(t, false)
	before, _ := cancelled.Clauses()
	calls := 0
	cancelled.limits.Stop = func() bool { calls++; return calls == 8 }
	if result, e := cancelled.Record(); result != nil || !errors.Is(e, ErrCancelled) {
		t.Fatal(result, e)
	}
	after, _ := cancelled.Clauses()
	if len(before) != len(after) {
		t.Fatal("cancelled record mutated manager")
	}
	for i := range before {
		if before[i].ProofID != after[i].ProofID || !slices.Equal(before[i].Literals, after[i].Literals) {
			t.Fatal("cancelled record mutated clauses")
		}
	}
}

func TestSatProofRecordingModelCampaign(t *testing.T) {
	for execution := 0; execution < 5000; execution++ {
		r := proofHypercube(t, 2+execution%5, execution&1 != 0)
		result, e := r.Record()
		if e != nil {
			t.Fatalf("%d %v", execution, e)
		}
		if e = replayProofResult(result); e != nil {
			t.Fatalf("%d %v", execution, e)
		}
	}
}

func FuzzSatProofRecording(f *testing.F) {
	f.Add(uint8(2), false)
	f.Add(uint8(6), true)
	f.Fuzz(func(t *testing.T, dimension uint8, reverseRoots bool) {
		r := proofHypercube(t, 2+int(dimension%5), reverseRoots)
		result, e := r.Record()
		if e != nil {
			t.Fatal(e)
		}
		if e = replayProofResult(result); e != nil {
			t.Fatal(e)
		}
		text, e := result.MarshalText(1 << 20)
		if e != nil || len(text) == 0 {
			t.Fatal(e)
		}
	})
}

func FuzzParseSatProofRecording(f *testing.F) {
	f.Add([]byte("p 2 6 4\n1 2 0\n-1 2 0\n1 -2 0\n-1 -2 0\n1 0\n0\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			return
		}
		r, e := ParseSatProofRecording(data, SatProofRecordingLimits{MaxBytes: 4096, MaxVariables: 64, MaxClauses: 128, MaxLiterals: 1024, MaxSteps: 4096})
		if e != nil {
			return
		}
		clauses, _ := r.Clauses()
		for _, c := range clauses {
			for i := 1; i < len(c.Literals); i++ {
				if c.Literals[i-1] >= c.Literals[i] {
					t.Fatal("not canonical")
				}
			}
		}
		if result, e := r.Record(); e == nil {
			if replayProofResult(result) != nil {
				t.Fatal("bad replay")
			}
		}
	})
}
