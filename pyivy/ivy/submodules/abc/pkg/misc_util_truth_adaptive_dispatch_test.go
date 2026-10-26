package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthAdaptiveDispatchNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_adaptive_dispatch.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, Threshold, Cost, Full, Phase int
		Source, Result                                []uint64
		Order                                         []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 1780 {
		t.Fatal(len(rows))
	}
	simple, full, algorithms := [2]int{}, [2]int{}, [2]bool{}
	for index, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: row.Threshold})
		if e != nil {
			t.Fatal(e)
		}
		phase := 0
		for target, label := range row.Order {
			if r.Permutation[label] != target {
				t.Fatal("permutation", index, r.Permutation, row.Order)
			}
			if r.InputPhases[label] {
				phase |= 1 << uint(target)
			}
		}
		if r.OutputPhase {
			phase |= 1 << uint(row.Variables)
		}
		if phase != row.Phase || r.EnumerationCost != row.Cost || r.FullEnumeration != (row.Full != 0) || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(index, row, r, phase)
		}
		algorithms[r.Algorithm] = true
		if r.FullEnumeration {
			full[r.Algorithm]++
		} else {
			simple[r.Algorithm]++
		}
		expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
		if e != nil || !expected.Equal(r.Table) {
			t.Fatal("metadata", index, e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
	if !algorithms[0] || !algorithms[1] || full[0] == 0 || full[1] == 0 || simple[0] == 0 || simple[1] == 0 {
		t.Fatal("missing branches", algorithms, full, simple)
	}
}
func TestTruthAdaptiveDispatchLimits(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	for _, tc := range []struct {
		options TruthAdaptiveOptions
		want    error
	}{
		{TruthAdaptiveOptions{Threshold: -1}, ErrDimension}, {TruthAdaptiveOptions{MaxConfigurations: -1}, ErrDimension},
		{TruthAdaptiveOptions{Threshold: 85, MaxConfigurations: 1}, ErrLimit},
		{TruthAdaptiveOptions{Threshold: 1085, MaxConfigurations: 4}, ErrLimit},
		{TruthAdaptiveOptions{Threshold: 1000, MaxConfigurations: 2}, ErrLimit},
		{TruthAdaptiveOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		r, e := source.CanonicalizeAdaptive(tc.options)
		if r.Table != nil || !errors.Is(e, tc.want) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthTable)(nil).CanonicalizeAdaptive(TruthAdaptiveOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, when := range []int{3, 6, 12} {
		calls := 0
		r, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: 1085, Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	r, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: 1085, MaxConfigurations: 5})
	if e != nil || r.Configurations != 5 || !r.FullEnumeration || r.Algorithm != 1 || r.EnumerationCost != 18 {
		t.Fatal(r, e)
	}
	r.Table.Set(0, false)
	r.Permutation[0] = 99
	r.InputPhases[0] = true
	fresh, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: 1085})
	if e != nil || fresh.Permutation[0] == 99 || fresh.InputPhases[0] || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	// C's zero-input branch clears the truth without preserving its phase; Go
	// explicitly retains a reconstructible configuration for both constants.
	for _, variables := range []int{0, 3} {
		for _, value := range []bool{false, true} {
			source, _ := NewConstantTruthTable(variables, value, 0)
			r, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: 1085})
			if e != nil || !r.Table.IsConstant(false) || r.OutputPhase != value || r.EnumerationCost != 0 {
				t.Fatal("constant extension", variables, r, e)
			}
			expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
			if e != nil || !expected.Equal(r.Table) {
				t.Fatal("constant metadata", e)
			}
		}
	}
}
func FuzzTruthAdaptiveDispatch(f *testing.F) {
	f.Add(uint16(4), uint64(0x96254718abcdef31))
	f.Add(uint16(128), uint64(0))
	f.Fuzz(func(t *testing.T, flags uint16, word uint64) {
		thresholds := [16]int{0, 1, 10, 25, 84, 85, 100, 110, 185, 1000, 1001, 1010, 1025, 1085, 1100, 1185}
		variables := int(flags % 6)
		width := 1 << uint(variables)
		word &= (uint64(1) << uint(width)) - 1
		for shift := width; shift < 64; shift *= 2 {
			word |= word << uint(shift)
		}
		source, e := TruthTableFromWords(variables, []uint64{word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		threshold := thresholds[(flags/6)%16]
		r, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: threshold})
		if e != nil {
			t.Fatal(e)
		}
		if r.Algorithm != 0 && threshold < 1000 || r.Algorithm != 1 && threshold >= 1000 {
			t.Fatal("algorithm")
		}
		if variables > 0 && r.FullEnumeration != (threshold%100 > 84 || r.EnumerationCost < threshold%100) {
			t.Fatal("dispatch boundary", threshold, r.EnumerationCost)
		}
		n := 1 << uint(variables)
		seen := make([]bool, variables)
		for _, target := range r.Permutation {
			if target < 0 || target >= variables || seen[target] {
				t.Fatal("permutation")
			}
			seen[target] = true
		}
		for x := 0; x < n; x++ {
			old := 0
			for v, target := range r.Permutation {
				value := x&(1<<uint(target)) != 0
				if value != r.InputPhases[v] {
					old |= 1 << uint(v)
				}
			}
			got, e := r.Table.Get(uint64(x))
			want := word&(uint64(1)<<uint(old)) != 0
			if e != nil || got != (want != r.OutputPhase) {
				t.Fatal("assignment", x, e)
			}
		}
		again, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: threshold})
		if e != nil || !again.Table.Equal(r.Table) || !slices.Equal(again.Permutation, r.Permutation) || !slices.Equal(again.InputPhases, r.InputPhases) || again.OutputPhase != r.OutputPhase || again.Configurations != r.Configurations {
			t.Fatal("determinism", e)
		}
		if source.words[0] != word {
			t.Fatal("source changed")
		}
	})
}
