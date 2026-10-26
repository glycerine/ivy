package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthAdaptiveSimpleNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_adaptive_simple.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, High, Phase int
		Source, Result               []uint64
		Expanded                     []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 288 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{HighSymmetry: row.High != 0})
		if e != nil {
			t.Fatal(e)
		}
		phase := 0
		for target, label := range row.Expanded {
			if r.Permutation[label] != target {
				t.Fatal("permutation", index, r.Permutation, row.Expanded)
			}
			if r.InputPhases[label] {
				phase |= 1 << uint(target)
			}
		}
		if r.OutputPhase {
			phase |= 1 << uint(row.Variables)
		}
		if phase != row.Phase || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(index, row.Variables, row.Mode, row.High, phase, row.Phase, r.Table.Words(), row.Result)
		}
		expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
		if e != nil || !expected.Equal(r.Table) {
			t.Fatal("forward metadata", index, e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthAdaptiveSimpleBudgetsAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	if r, e := source.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{MaxRefinements: 1}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
	if r, e := source.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{MaxRefinements: -1}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	if r, e := (*TruthTable)(nil).CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, when := range []int{1, 2, 5} {
		calls := 0
		r, e := source.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	r, e := source.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{HighSymmetry: true})
	if e != nil {
		t.Fatal(e)
	}
	expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
	if e != nil || !expected.Equal(r.Table) {
		t.Fatal("wide", e)
	}
	if r.Rounds < 1 || r.Rounds > 5 || r.Attempts < 35 || r.Changes > r.Attempts {
		t.Fatal("counters", r)
	}
	r.Table.Set(0, false)
	r.Permutation[0] = 99
	r.InputPhases[0] = true
	fresh, e := source.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{HighSymmetry: true})
	if e != nil || !slices.Equal(source.Words(), before) || fresh.Permutation[0] == 99 || fresh.InputPhases[0] {
		t.Fatal("ownership", e)
	}
	for _, value := range []bool{false, true} {
		constant, _ := NewConstantTruthTable(0, value, 0)
		r, e := constant.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{})
		if e != nil || !r.Table.IsConstant(false) || r.OutputPhase != value || len(r.Permutation) != 0 || r.Attempts != 0 {
			t.Fatal("zero input", r, e)
		}
	}
}
func TestTruthAdaptiveGroupSwapMetadata(t *testing.T) {
	// Swap unequal-length chains, with unused inputs appended and source phases.
	source, _ := NewVariableTruthTable(6, 3, 0)
	s := &truthAdaptiveState{table: source.Clone(), order: []int{0, 1, 2, 3, 4, 5}, position: []int{0, 1, 2, 3, 4, 5}, representatives: []int{0, 2}, phases: make([]bool, 6), links: []int{1, -1, 3, 4, -1, -1, 5}, weights: []int{2, 1, 3, 1, 0, 1}}
	before := s.clone()
	s.swapGroups(0)
	if !slices.Equal(s.order, []int{2, 3, 4, 0, 1, 5}) {
		t.Fatal(s.order)
	}
	s.flipGroup(0)
	expected, e := source.TransformNPN(s.position, s.phases, false)
	if e != nil || !expected.Equal(s.table) {
		t.Fatal("group transform", e)
	}
	if s.phases[4] || !s.phases[2] || !s.phases[3] {
		t.Fatal("phase selection")
	}
	s.flipGroup(0)
	s.swapGroups(0)
	if !s.table.Equal(before.table) || !slices.Equal(s.order, before.order) || !slices.Equal(s.phases, before.phases) {
		t.Fatal("group involutions")
	}
}

// C's swap-only branch deliberately remains observable even though its
// comparison direction differs from the minimizing full-phase branch.
func TestTruthAdaptiveSwapOnlyDirection(t *testing.T) {
	source, _ := TruthTableFromWords(2, []uint64{0x2222222222222222}, 0)
	state := &truthAdaptiveState{table: source, order: []int{0, 1}, position: []int{0, 1}, representatives: []int{0, 1}, phases: make([]bool, 2), links: []int{-1, -1, -1}, weights: []int{1, 1}}
	result, config := state.refineGroups(0, true)
	if config != 4 || result.table.words[0] != 0x4444444444444444 || source.words[0] != 0x2222222222222222 {
		t.Fatal(config, result.table.Words(), source.Words())
	}
}

func FuzzTruthAdaptiveSimple(f *testing.F) {
	f.Add(uint8(4), []byte{0x88, 0x66})
	f.Add(uint8(128), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 256 {
			data = data[:256]
		}
		variables := int(flags % 11)
		n := 1 << uint(variables)
		count, _ := TruthTableWordCount(variables)
		model := make([]bool, n)
		words := make([]uint64, count)
		for x := range model {
			model[x] = data[(x/8)%len(data)]&(1<<uint(x%8)) != 0
		}
		for bit := 0; bit < count*64; bit++ {
			if model[bit%n] {
				words[bit/64] |= uint64(1) << uint(bit%64)
			}
		}
		source, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		options := TruthAdaptiveSimpleOptions{HighSymmetry: flags&128 != 0}
		r, e := source.CanonicalizeAdaptiveSimple(options)
		if e != nil {
			t.Fatal(e)
		}
		seen := make([]bool, variables)
		for _, target := range r.Permutation {
			if target < 0 || target >= variables || seen[target] {
				t.Fatal("permutation")
			}
			seen[target] = true
		}
		for target := 0; target < n; target++ {
			old := 0
			for source, position := range r.Permutation {
				value := target&(1<<uint(position)) != 0
				if value != r.InputPhases[source] {
					old |= 1 << uint(source)
				}
			}
			got, e := r.Table.Get(uint64(target))
			if e != nil || got != (model[old] != r.OutputPhase) {
				t.Fatal("configuration", target, e)
			}
		}
		if r.Rounds < 1 || r.Rounds > 5 || r.Attempts < r.Changes || r.Changes < 0 {
			t.Fatal("counters")
		}
		again, e := source.CanonicalizeAdaptiveSimple(options)
		if e != nil || !again.Table.Equal(r.Table) || !slices.Equal(again.Permutation, r.Permutation) || !slices.Equal(again.InputPhases, r.InputPhases) || again.OutputPhase != r.OutputPhase || again.Attempts != r.Attempts || again.Changes != r.Changes {
			t.Fatal("determinism", e)
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source changed")
		}
	})
}
