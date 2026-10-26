package abc

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"slices"
	"testing"
)

func TestTruthCostAwareNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_cost_aware.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, Enabled, Broadened, Phase int
		EnumerationEstimate, PhaseEstimate         float64
		Source, Result                             []uint64
		Order                                      []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 202 {
		t.Fatal(len(rows))
	}
	broadened := false
	for index, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeCostAware(TruthCostAwareOptions{BroadenSearch: row.Enabled != 0})
		if e != nil {
			t.Fatal(index, e)
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
		if phase != row.Phase || r.Broadened != (row.Broadened != 0) || math.Abs(r.EnumerationEstimate-row.EnumerationEstimate) > 1e-12 || math.Abs(r.PhaseEstimate-row.PhaseEstimate) > 1e-12 || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(index, row, r, phase)
		}
		broadened = broadened || r.Broadened
		expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
		if e != nil || !expected.Equal(r.Table) {
			t.Fatal("metadata", e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
	if !broadened {
		t.Fatal("native cases missed broadening branch")
	}
}
func TestTruthCostAwareReset(t *testing.T) {
	source, _ := TruthTableFromWords(4, []uint64{0xa08ca08ca08ca08c}, 0)
	reduction, e := source.ReduceInputSymmetry(true)
	if e != nil {
		t.Fatal(e)
	}
	s := newTruthAdaptiveState(reduction)
	before := s.clone()
	s.resetGroups()
	if !slices.IsSorted(s.representatives) || s.phased || !slices.Equal(s.links, before.links) || !slices.Equal(s.weights, before.weights) || !slices.Equal(s.phases, before.phases) {
		t.Fatal("reset invariants")
	}
	if len(s.representatives) > 0 && (len(s.groups) != 1 || s.groups[0].Start != 0 || s.groups[0].Size != len(s.representatives)) {
		t.Fatal("reset groups")
	}
	expected, e := source.TransformNPN(s.position, s.phases, s.output)
	if e != nil || !expected.Equal(s.table) {
		t.Fatal("reset metadata", e)
	}
	if !slices.Equal(before.groups, reduction.Groups) {
		t.Fatal("groups alias")
	}
}
func TestTruthCostAwareLimitsAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	for _, tc := range []struct {
		options TruthCostAwareOptions
		want    error
	}{
		{TruthCostAwareOptions{MaxConfigurations: -1}, ErrDimension}, {TruthCostAwareOptions{MaxConfigurations: 1}, ErrLimit},
		{TruthCostAwareOptions{MaxConfigurations: 37}, ErrLimit}, {TruthCostAwareOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		r, e := source.CanonicalizeCostAware(tc.options)
		if r.Table != nil || !errors.Is(e, tc.want) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthTable)(nil).CanonicalizeCostAware(TruthCostAwareOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, when := range []int{3, 20, 44} {
		calls := 0
		r, e := source.CanonicalizeCostAware(TruthCostAwareOptions{Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	r, e := source.CanonicalizeCostAware(TruthCostAwareOptions{BroadenSearch: true, MaxConfigurations: 38})
	if e != nil || r.Configurations != 38 || r.Broadened || r.Permutations != 1 {
		t.Fatal(r, e)
	}
	expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
	if e != nil || !expected.Equal(r.Table) {
		t.Fatal("wide", e)
	}
	r.Table.Set(0, false)
	r.Permutation[0] = 99
	r.InputPhases[0] = true
	fresh, e := source.CanonicalizeCostAware(TruthCostAwareOptions{})
	if e != nil || fresh.Permutation[0] == 99 || fresh.InputPhases[0] || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	for _, variables := range []int{0, 3} {
		for _, value := range []bool{false, true} {
			source, _ := NewConstantTruthTable(variables, value, 0)
			r, e := source.CanonicalizeCostAware(TruthCostAwareOptions{BroadenSearch: true})
			if e != nil || !r.Table.IsConstant(false) || r.OutputPhase != value {
				t.Fatal("constant", variables, r, e)
			}
		}
	}
}
func FuzzTruthCostAware(f *testing.F) {
	f.Add(uint8(4), uint64(0x96254718abcdef31))
	f.Add(uint8(128), uint64(0))
	f.Fuzz(func(t *testing.T, flags uint8, word uint64) {
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
		options := TruthCostAwareOptions{BroadenSearch: flags&128 != 0}
		r, e := source.CanonicalizeCostAware(options)
		if e != nil {
			t.Fatal(e)
		}
		if variables > 0 && r.Broadened != (options.BroadenSearch && r.EnumerationEstimate > r.PhaseEstimate) {
			t.Fatal("cost decision")
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
		again, e := source.CanonicalizeCostAware(options)
		if e != nil || !again.Table.Equal(r.Table) || !slices.Equal(again.Permutation, r.Permutation) || !slices.Equal(again.InputPhases, r.InputPhases) || again.OutputPhase != r.OutputPhase || again.Configurations != r.Configurations {
			t.Fatal("determinism", e)
		}
		if source.words[0] != word {
			t.Fatal("source changed")
		}
	})
}
