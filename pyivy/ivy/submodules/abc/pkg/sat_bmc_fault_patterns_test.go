package abc

// Stand-alone tests for bmcFault.c pattern persistence and reporting.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestFaultPatternRoundTripAndFile(t *testing.T) {
	patterns := []bool{false, true, true, false, true, false, false, true, true, true, false, false}
	data, err := MarshalFaultPatterns(patterns, 4, FaultPatternOptions{})
	if err != nil || string(data) != "0110\n1001\n1100\n" {
		t.Fatal(string(data), err)
	}
	parsed, err := ParseFaultPatterns([]byte(" 0110\n\t1001\r\n1100 "), FaultPatternOptions{})
	if err != nil || !slices.Equal(parsed, patterns) {
		t.Fatal(parsed, err)
	}
	path := filepath.Join(t.TempDir(), "vectors.txt")
	if err = WriteFaultPatterns(path, patterns, 4, FaultPatternOptions{}); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadFaultPatterns(path, FaultPatternOptions{})
	if err != nil || !slices.Equal(loaded, patterns) {
		t.Fatal(loaded, err)
	}
	if raw, err := os.ReadFile(path); err != nil || string(raw) != string(data) {
		t.Fatal(string(raw), err)
	}
}

func TestFaultDelayPatterns(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	r, _ := g.AppendCI()
	next, _ := g.AppendXorReal(a, r)
	out, _ := g.AppendAnd(b, r)
	_, _ = g.AppendCO(out)
	_, _ = g.AppendCO(next)
	_ = g.SetRegisterCount(1)
	// Rows are [RO0, PI0(a,b), PI1(a,b)].
	patterns := []bool{true, false, true, true, false, false, true, false, false, true}
	data, err := MarshalFaultDelayPatterns(g, patterns, FaultPatternOptions{})
	if err != nil || string(data) != "011\n101\n100\n011\n" {
		t.Fatal(string(data), err)
	}
}

func TestDuplicateGiaWithPrimaryInputs(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	r, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, r)
	_, _ = g.AppendCO(n)
	_, _ = g.AppendCO(a.Not())
	_ = g.SetRegisterCount(1)
	g.Name, g.Spec = "dup", "dup.aig"
	result, extra, err := DuplicateGiaWithPrimaryInputs(g, 2, GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.PrimaryInputCount() != 3 || result.RegisterCount() != 1 || len(extra) != 2 || result.Name != g.Name || result.Spec != g.Spec {
		t.Fatal(result.PrimaryInputCount(), result.RegisterCount(), extra)
	}
	for mask := 0; mask < 16; mask++ {
		primary := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		state := []bool{mask&8 != 0}
		got, next, err := result.Step(primary, state)
		want, wantNext, sourceErr := g.Step(primary[:1], state)
		if err != nil || sourceErr != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(next, wantNext) {
			t.Fatal(mask, got, next, want, wantNext, err, sourceErr)
		}
	}
}

func TestFaultRunStatistics(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	_, _ = g.AppendCO(n)
	g.Name = "stats"
	s := NewSatSolver()
	_ = s.SetVariableCount(3)
	_, _ = s.AddClause([]SatLit{satCardinalityLit(0, false), satCardinalityLit(1, true)})
	stats, err := FaultRunStatistics(g, s, 7)
	if err != nil || stats.Name != "stats" || stats.PrimaryInputs != 2 || stats.PrimaryOutputs != 1 || stats.Gates != 1 || stats.Variables != 3 || stats.Clauses != 1 || stats.Iterations != 7 {
		t.Fatal(stats, err)
	}
}

func TestFaultPatternFailures(t *testing.T) {
	if _, err := MarshalFaultPatterns([]bool{true}, 2, FaultPatternOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := ParseFaultPatterns([]byte("01x"), FaultPatternOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for _, options := range []FaultPatternOptions{{MaxBits: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if _, err := MarshalFaultPatterns([]bool{true, false}, 2, options); !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, err)
		}
	}
	g := NewGiaManager()
	if _, _, err := DuplicateGiaWithPrimaryInputs(g, 2, GiaCofactorOptions{MaxObjects: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzFaultPatternRoundTrip(f *testing.F) {
	for _, seed := range []string{"0", "1", "0110", "10101010"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		bits := make([]bool, 0, min(len(text), 64))
		for i := 0; i < len(text) && i < 64; i++ {
			bits = append(bits, text[i]&1 != 0)
		}
		if len(bits) == 0 {
			bits = append(bits, false)
		}
		data, err := MarshalFaultPatterns(bits, 1, FaultPatternOptions{})
		if err != nil {
			t.Fatal(err)
		}
		got, err := ParseFaultPatterns(data, FaultPatternOptions{})
		if err != nil || !slices.Equal(got, bits) {
			t.Fatal(got, bits, err)
		}
	})
}
