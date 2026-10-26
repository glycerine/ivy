package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthRefinementNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_refine.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Pairs []struct {
			Variables, First, Mode, OnlyPermutation, Configuration int
			Source, Result                                         []uint64
		}
		Canonical []struct {
			Variables, Mode, OnlyPermutation int
			Phase                            uint64
			Source, Result                   []uint64
			Permutation                      []int
		}
	}
	if e = json.Unmarshal(data, &fixture); e != nil {
		t.Fatal(e)
	}
	if len(fixture.Pairs) != 576 || len(fixture.Canonical) != 160 {
		t.Fatal("fixture dimensions")
	}
	for _, row := range fixture.Pairs {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.RefineAdjacentCofactors(row.First, row.OnlyPermutation != 0)
		if e != nil || int(r.Configuration) != row.Configuration || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal("pair", row.Variables, row.First, row.Mode, row.OnlyPermutation, r.Configuration, row.Configuration, e)
		}
	}
	for _, row := range fixture.Canonical {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeNPN(TruthCanonicalOptions{OnlyPermutation: row.OnlyPermutation != 0})
		if e != nil {
			t.Fatal(e)
		}
		inverse := make([]int, row.Variables)
		targetPhase := make([]bool, row.Variables)
		phase := uint64(0)
		if r.OutputPhase {
			phase |= uint64(1) << uint(row.Variables)
		}
		for original, target := range r.Permutation {
			inverse[target] = original
			targetPhase[target] = r.InputPhases[original]
			if targetPhase[target] {
				phase |= uint64(1) << uint(target)
			}
		}
		if phase != row.Phase || !slices.Equal(inverse, row.Permutation) || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal("canonical", row.Variables, row.Mode, row.OnlyPermutation, phase, row.Phase, inverse, row.Permutation, r.Table.Words(), row.Result)
		}
		forward, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
		if e != nil || !forward.Equal(r.Table) {
			t.Fatal("forward", e)
		}
		backward, e := r.Table.TransformNPN(inverse, targetPhase, r.OutputPhase)
		if e != nil || !backward.Equal(source) {
			t.Fatal("inverse", e)
		}
	}
}
func TestTruthCanonicalBudgetsAndOwnership(t *testing.T) {
	source, _ := NewConstantTruthTable(6, false, 0)
	before := source.Words()
	if r, e := source.CanonicalizeNPN(TruthCanonicalOptions{MaxRefinements: 1}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if r, e := source.CanonicalizeNPN(TruthCanonicalOptions{MaxRefinements: -1}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	for _, threshold := range []int{1, 4, 7} {
		calls := 0
		r, e := source.CanonicalizeNPN(TruthCanonicalOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, first := range []int{-1, 5} {
		r, e := source.RefineAdjacentCofactors(first, false)
		if r.Table != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	r, e := source.CanonicalizeNPN(TruthCanonicalOptions{})
	if e != nil || r.Rounds != 1 || r.Attempts != 5 || r.Changes != 0 {
		t.Fatal(r, e)
	}
	r.Table.Set(0, true)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("alias")
	}
	if r, e := (*TruthTable)(nil).CanonicalizeNPN(TruthCanonicalOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	wide, _ := NewVariableTruthTable(18, 0, 0)
	r, e = wide.CanonicalizeNPN(TruthCanonicalOptions{})
	if e != nil {
		t.Fatal(e)
	}
	forward, e := wide.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
	if e != nil || !forward.Equal(r.Table) {
		t.Fatal("wide configuration", e)
	}
}
func FuzzTruthRefinement(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(128), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 96 {
			data = data[:96]
		}
		variables := int(flags % 10)
		n := 1 << uint(variables)
		model := make([]bool, n)
		for i := range model {
			model[i] = data[(i/8)%len(data)]&(1<<uint(i%8)) != 0
		}
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		for i := 0; i < count*64; i++ {
			if model[i%n] {
				words[i/64] |= uint64(1) << uint(i%64)
			}
		}
		source, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		only := flags&128 != 0
		if variables >= 2 {
			first := int(data[0]) % (variables - 1)
			r, e := source.RefineAdjacentCofactors(first, only)
			if e != nil {
				t.Fatal(e)
			}
			best := slices.Clone(model)
			config := uint8(0)
			configs := []uint8{1, 3, 2, 6, 7, 5, 4}
			if only {
				configs = []uint8{4}
			}
			for _, candidate := range configs {
				values := make([]bool, n)
				for assignment := 0; assignment < n; assignment++ {
					old := assignment
					if candidate&4 != 0 {
						a, b := (assignment>>uint(first))&1, (assignment>>uint(first+1))&1
						if a != b {
							old ^= (1 << uint(first)) | (1 << uint(first+1))
						}
					}
					if candidate&1 != 0 {
						old ^= 1 << uint(first)
					}
					if candidate&2 != 0 {
						old ^= 1 << uint(first+1)
					}
					values[assignment] = model[old]
				}
				less := false
				for assignment := n - 1; assignment >= 0; assignment-- {
					if values[assignment] != best[assignment] {
						less = !values[assignment] && best[assignment]
						break
					}
				}
				if less {
					best = values
					config = candidate
				}
			}
			if r.Configuration != config {
				t.Fatal("tie/config order", r.Configuration, config)
			}
			for assignment, want := range best {
				got, e := r.Table.Get(uint64(assignment))
				if e != nil || got != want {
					t.Fatal("pair result", assignment)
				}
			}
		}
		r, e := source.CanonicalizeNPN(TruthCanonicalOptions{OnlyPermutation: only})
		if e != nil {
			t.Fatal(e)
		}
		if r.Rounds < 1 || r.Rounds > 5 || r.Changes > r.Attempts {
			t.Fatal("work counters")
		}
		if only {
			if r.OutputPhase {
				t.Fatal("output phase")
			}
			for _, flip := range r.InputPhases {
				if flip {
					t.Fatal("input phase")
				}
			}
		}
		for assignment := 0; assignment < n; assignment++ {
			old := 0
			for original, target := range r.Permutation {
				value := assignment&(1<<uint(target)) != 0
				if value != r.InputPhases[original] {
					old |= 1 << uint(original)
				}
			}
			got, e := r.Table.Get(uint64(assignment))
			if e != nil || got != (model[old] != r.OutputPhase) {
				t.Fatal("canonical configuration", assignment)
			}
		}
		pop, e := r.Table.CofactorPopulations()
		if e != nil || pop.Ones != r.PackedOnes || !slices.Equal(pop.Zero, r.ZeroCofactorCounts) {
			t.Fatal("signature preservation", e)
		}
		semi, e := source.SemiCanonicalize(only)
		if e != nil {
			t.Fatal(e)
		}
		comparison, e := r.Table.Compare(semi.Table, true)
		if e != nil || comparison > 0 {
			t.Fatal("refinement increased truth value")
		}
		if !slices.Equal(words, source.Words()) {
			t.Fatal("source mutation")
		}
	})
}
