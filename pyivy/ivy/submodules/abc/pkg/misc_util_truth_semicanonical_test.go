package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSemiCanonicalNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_semicanonical.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, OnlyPermutation, SourceOnes, ResultOnes int
		Phase                                                    uint64
		Source, Result                                           []uint64
		SourceZero, ResultZero, Permutation                      []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 120 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		pop, e := source.CofactorPopulations()
		if e != nil || pop.Ones != row.SourceOnes || !slices.Equal(pop.Zero, row.SourceZero) {
			t.Fatal("population", row.Variables, pop, e)
		}
		r, e := source.SemiCanonicalize(row.OnlyPermutation != 0)
		if e != nil {
			t.Fatal(e)
		}
		inverse := make([]int, row.Variables)
		targetPhases := make([]bool, row.Variables)
		phase := uint64(0)
		if r.OutputPhase {
			phase |= uint64(1) << uint(row.Variables)
		}
		for original, target := range r.Permutation {
			inverse[target] = original
			targetPhases[target] = r.InputPhases[original]
			if targetPhases[target] {
				phase |= uint64(1) << uint(target)
			}
		}
		if phase != row.Phase || r.PackedOnes != row.ResultOnes || !slices.Equal(inverse, row.Permutation) || !slices.Equal(r.ZeroCofactorCounts, row.ResultZero) || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(row.Variables, row.Mode, row.OnlyPermutation, r, phase, row.Phase, inverse, row.Permutation)
		}
		forward, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
		if e != nil || !forward.Equal(r.Table) {
			t.Fatal("forward", e)
		}
		restored, e := r.Table.TransformNPN(inverse, targetPhases, r.OutputPhase)
		if e != nil || !restored.Equal(source) {
			t.Fatal("inverse", e)
		}
	}
}
func TestTruthSemiCanonicalWideAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 0, 0)
	before := source.Words()
	r, e := source.SemiCanonicalize(false)
	if e != nil || r.OutputPhase || !r.InputPhases[0] || r.Permutation[0] != 17 {
		t.Fatal(r, e)
	}
	expected, _ := NewVariableTruthTable(18, 17, 0)
	if !r.Table.Equal(expected.Not()) {
		t.Fatal("phase/permutation direction")
	}
	r.Table.Set(0, false)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("alias")
	}
	only, e := source.SemiCanonicalize(true)
	if e != nil || only.OutputPhase {
		t.Fatal(e)
	}
	for _, flip := range only.InputPhases {
		if flip {
			t.Fatal("permutation-only phase")
		}
	}
	if _, e := (*TruthTable)(nil).CofactorPopulations(); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if r, e := (*TruthTable)(nil).SemiCanonicalize(false); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzTruthSemiCanonical(f *testing.F) {
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
		scale := 1
		if n < 64 {
			scale = 64 / n
		}
		ones := 0
		zero := make([]int, variables)
		for assignment, value := range model {
			if !value {
				continue
			}
			ones++
			for v := 0; v < variables; v++ {
				if assignment&(1<<uint(v)) == 0 {
					zero[v]++
				}
			}
		}
		pop, e := source.CofactorPopulations()
		if e != nil || pop.Ones != ones*scale {
			t.Fatal(e)
		}
		for v := range zero {
			if pop.Zero[v] != zero[v]*scale || pop.One[v] != (ones-zero[v])*scale {
				t.Fatal("packed counts")
			}
		}
		only := flags&128 != 0
		output := !only && ones > n/2
		phases := make([]bool, variables)
		if output {
			ones = n - ones
			for v := range zero {
				zero[v] = n/2 - zero[v]
			}
		}
		if !only {
			for v, count := range zero {
				if count < ones-count {
					phases[v] = true
					zero[v] = ones - count
				}
			}
		}
		labels := make([]int, variables)
		for v := range labels {
			labels[v] = v
		}
		for i := 0; i < variables-1; i++ {
			best := i + 1
			for k := i + 2; k < variables; k++ {
				if zero[k] < zero[best] {
					best = k
				}
			}
			if zero[i] > zero[best] {
				labels[i], labels[best] = labels[best], labels[i]
				zero[i], zero[best] = zero[best], zero[i]
			}
		}
		permutation := make([]int, variables)
		for target, original := range labels {
			permutation[original] = target
		}
		r, e := source.SemiCanonicalize(only)
		if e != nil || r.OutputPhase != output || r.PackedOnes != ones*scale || !slices.Equal(r.InputPhases, phases) || !slices.Equal(r.Permutation, permutation) {
			t.Fatal("signature", r, e)
		}
		for v, count := range zero {
			if r.ZeroCofactorCounts[v] != count*scale {
				t.Fatal("sorted cofactor count")
			}
		}
		for assignment := 0; assignment < n; assignment++ {
			old := 0
			for v, target := range permutation {
				value := assignment&(1<<uint(target)) != 0
				if value != phases[v] {
					old |= 1 << uint(v)
				}
			}
			got, e := r.Table.Get(uint64(assignment))
			if e != nil || got != (model[old] != output) {
				t.Fatal("configuration", variables, assignment)
			}
		}
		resultPop, e := r.Table.CofactorPopulations()
		if e != nil || resultPop.Ones != r.PackedOnes || !slices.Equal(resultPop.Zero, r.ZeroCofactorCounts) {
			t.Fatal("result populations", e)
		}
		if !slices.Equal(words, source.Words()) {
			t.Fatal("source mutation")
		}
	})
}
