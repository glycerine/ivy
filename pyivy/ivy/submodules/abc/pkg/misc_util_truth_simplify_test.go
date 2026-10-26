package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSimplifyNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/util_truth_simplify.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, Moved        int
		Source, Result                uint64
		InputBindings, ResultBindings []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 35 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, []uint64{row.Source}, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.SimplifyInputs(row.InputBindings)
		if e != nil {
			t.Fatal(e)
		}
		if r.Moved != (row.Moved != 0) || !slices.Equal(r.Bindings, row.ResultBindings) || !slices.Equal(r.Table.Words(), []uint64{row.Result}) || r.Table.Variables() != len(row.ResultBindings) {
			t.Fatal(row.Variables, row.Mode, r.Bindings, r.Table.Words(), row.ResultBindings, row.Result)
		}
	}
}
func TestTruthSimplifyWideAliasesAndValidation(t *testing.T) {
	source, _ := NewVariableTruthTable(8, 7, 0)
	bindings := make([]GiaLit, 8)
	for i := range bindings {
		bindings[i] = GiaLit(22).NotIf(i&1 != 0)
	}
	before := slices.Clone(bindings)
	words := source.Words()
	r, e := source.SimplifyInputs(bindings)
	if e != nil {
		t.Fatal(e)
	}
	expected, _ := NewVariableTruthTable(1, 0, 0)
	if !r.Table.Equal(expected.Not()) || !slices.Equal(r.Bindings, []GiaLit{22}) || !slices.Equal(r.SourceVariables, []int{0}) {
		t.Fatal(r)
	}
	if !slices.Equal(bindings, before) || !slices.Equal(words, source.Words()) {
		t.Fatal("mutation")
	}
	r.Bindings[0] = 99
	if !slices.Equal(bindings, before) {
		t.Fatal("binding alias")
	}
	for _, bad := range [][]GiaLit{nil, make([]GiaLit, 9), {22, 22, 22, 22, 22, 22, 22, InvalidGiaLit}} {
		r, e := source.SimplifyInputs(bad)
		if r.Table != nil || e == nil {
			t.Fatal(e)
		}
	}
	var nilTruth *TruthTable
	if r, e := nilTruth.SimplifyInputs(nil); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	constants := make([]GiaLit, 8)
	constants[7] = GiaConst1
	r, e = source.SimplifyInputs(constants)
	if e != nil || !r.Table.IsConstant(true) || len(r.Bindings) != 0 || r.Table.Variables() != 0 {
		t.Fatal(e)
	}
}
func FuzzTruthSimplify(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5, 6})
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
		bindings := make([]GiaLit, variables)
		labels := []GiaID{}
		positions := make(map[GiaID]int)
		for i := range bindings {
			v := data[i%len(data)]
			id := GiaID(v % 8)
			bindings[i] = GiaLit(id << 1).NotIf(v&128 != 0)
			if id != 0 {
				if _, ok := positions[id]; !ok {
					positions[id] = len(labels)
					labels = append(labels, id)
				}
			}
		}
		saved := slices.Clone(bindings)
		r, e := source.SimplifyInputs(bindings)
		if e != nil {
			t.Fatal(e)
		}
		if r.Table.Variables() != len(r.Bindings) || len(r.Bindings) != len(r.SourceVariables) || !slices.Equal(saved, bindings) || !slices.Equal(words, source.Words()) {
			t.Fatal("metadata")
		}
		used := make(map[GiaID]bool)
		last := -1
		for i, l := range r.Bindings {
			slot := r.SourceVariables[i]
			if slot <= last || slot >= len(bindings) || l != bindings[slot] || l.ID() == 0 || used[l.ID()] {
				t.Fatal("projection", r)
			}
			last = slot
			used[l.ID()] = true
		}
		evaluate := func(l GiaLit, assignment int) bool {
			value := false
			if l.ID() != 0 {
				value = assignment&(1<<uint(positions[l.ID()])) != 0
			}
			return value != l.Complemented()
		}
		for assignment := 0; assignment < 1<<uint(len(labels)); assignment++ {
			old, target := 0, 0
			for i, l := range bindings {
				if evaluate(l, assignment) {
					old |= 1 << uint(i)
				}
			}
			for i, l := range r.Bindings {
				if evaluate(l, assignment) {
					target |= 1 << uint(i)
				}
			}
			got, e := r.Table.Get(uint64(target))
			if e != nil || got != model[old] {
				t.Fatal("bound truth", variables, bindings, r.Bindings, assignment, old, target, e)
			}
		}
	})
}
