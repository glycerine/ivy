package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthTransformNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/util_truth_transform.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, First, Second                int
		Source, Flip, Swap, Npn, Exists, Forall []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 165 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		flipped, _ := source.FlipVariable(row.First)
		swapped, _ := source.SwapVariables(row.First, row.Second)
		order := make([]int, row.Variables)
		phases := make([]bool, row.Variables)
		for i := range order {
			order[i] = (i + 1) % len(order)
			phases[i] = i&1 != 0
		}
		npn, _ := source.TransformNPN(order, phases, true)
		exists, _ := source.Existential(row.First)
		forall, _ := source.Universal(row.First)
		for _, pair := range []struct {
			got  *TruthTable
			want []uint64
		}{{flipped, row.Flip}, {swapped, row.Swap}, {npn, row.Npn}, {exists, row.Exists}, {forall, row.Forall}} {
			if pair.got == nil || !slices.Equal(pair.got.Words(), pair.want) {
				t.Fatal(row.Variables, row.First, row.Second, pair.got, pair.want)
			}
		}
	}
}
func TestTruthTransformDimensionsAndLargeQuantification(t *testing.T) {
	source, _ := NewVariableTruthTable(13, 12, 0)
	exists, e := source.Existential(12)
	if e != nil || !exists.IsConstant(true) {
		t.Fatal(e)
	}
	forall, e := source.Universal(12)
	if e != nil || !forall.IsConstant(false) {
		t.Fatal(e)
	}
	before := source.Words()
	for _, v := range []int{-1, 13} {
		if r, e := source.FlipVariable(v); r != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
		if r, e := source.SwapVariables(v, 0); r != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
		if r, e := source.Existential(v); r != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if r, e := source.TransformNPN([]int{0}, nil, false); r != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if r, e := source.TransformNPN(nil, []bool{true}, false); r != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	small, _ := NewVariableTruthTable(3, 0, 0)
	for _, order := range [][]int{{0, 0, 2}, {0, 1, 3}, {-1, 1, 2}} {
		if r, e := small.PermuteVariables(order); r != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	order := []int{2, 0, 1}
	phases := []bool{true, false, true}
	orderBefore := slices.Clone(order)
	phaseBefore := slices.Clone(phases)
	small.TransformNPN(order, phases, true)
	if !slices.Equal(order, orderBefore) || !slices.Equal(phases, phaseBefore) || !slices.Equal(before, source.Words()) {
		t.Fatal("source/argument mutation")
	}
	identity, e := source.PermuteVariables(nil)
	if e != nil || !identity.Equal(source) {
		t.Fatal(e)
	}
	identity.Set(0, true)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("identity alias")
	}
	var nilTruth *TruthTable
	if r, e := nilTruth.SwapVariables(0, 0); r != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzTruthTransform(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
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
		order := make([]int, variables)
		phases := make([]bool, variables)
		for i := range order {
			order[i] = i
			phases[i] = data[i%len(data)]&1 != 0
		}
		for i := variables - 1; i > 0; i-- {
			j := int(data[i%len(data)]) % (i + 1)
			order[i], order[j] = order[j], order[i]
		}
		npn, e := source.TransformNPN(order, phases, flags&128 != 0)
		if e != nil {
			t.Fatal(e)
		}
		permuted, e := source.PermuteVariables(order)
		if e != nil {
			t.Fatal(e)
		}
		first, second := 0, 0
		var flipped, swapped, exists, forall *TruthTable
		if variables > 0 {
			first = int(data[0]) % variables
			second = int(data[len(data)-1]) % variables
			flipped, _ = source.FlipVariable(first)
			swapped, _ = source.SwapVariables(first, second)
			exists, _ = source.Existential(first)
			forall, _ = source.Universal(first)
		}
		for assignment := 0; assignment < n; assignment++ {
			old, npnOld := 0, 0
			for variable, target := range order {
				value := assignment&(1<<uint(target)) != 0
				if value {
					old |= 1 << uint(variable)
				}
				if value != phases[variable] {
					npnOld |= 1 << uint(variable)
				}
			}
			got, e := permuted.Get(uint64(assignment))
			if e != nil || got != model[old] {
				t.Fatal("permutation direction", order, assignment, old)
			}
			got, e = npn.Get(uint64(assignment))
			if e != nil || got != (model[npnOld] != (flags&128 != 0)) {
				t.Fatal("NPN", order, phases, assignment)
			}
			if variables > 0 {
				swapAssignment := assignment
				if (assignment>>uint(first))&1 != (assignment>>uint(second))&1 {
					swapAssignment ^= (1 << uint(first)) | (1 << uint(second))
				}
				zero := assignment &^ (1 << uint(first))
				one := zero | (1 << uint(first))
				wants := []bool{model[assignment^(1<<uint(first))], model[swapAssignment], model[zero] || model[one], model[zero] && model[one]}
				for i, table := range []*TruthTable{flipped, swapped, exists, forall} {
					got, e := table.Get(uint64(assignment))
					if e != nil || got != wants[i] {
						t.Fatal("transform", variables, first, second, assignment, i)
					}
				}
			}
		}
		if !slices.Equal(words, source.Words()) {
			t.Fatal("source mutation")
		}
	})
}
