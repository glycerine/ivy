package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSupportNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/util_truth_support.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Variant, Moved              int
		Support                                []int
		Source, Compacted, Restored, Stretched []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 30 {
		t.Fatal(len(rows))
	}
	for _, r := range rows {
		source, e := TruthTableFromWords(r.Variables, r.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		reduced, e := source.CompactSupport()
		if e != nil {
			t.Fatal(e)
		}
		if reduced.Moved != (r.Moved != 0) || !slices.Equal(reduced.SourceVariables, r.Support) || !slices.Equal(reduced.Table.Words(), r.Compacted) || reduced.Table.Variables() != len(r.Support) {
			t.Fatal(r.Variables, r.Variant, reduced)
		}
		restored, e := reduced.Table.ExpandVariables(r.Variables, reduced.SourceVariables)
		if e != nil || !slices.Equal(restored.Words(), r.Restored) || !restored.Equal(source) {
			t.Fatal("restore", r.Variables, e)
		}
		stretched, e := source.Stretch(r.Variables + 2)
		if e != nil || !slices.Equal(stretched.Words(), r.Stretched) {
			t.Fatal("stretch", r.Variables, e)
		}
	}
}
func TestTruthSupportProjectionValidation(t *testing.T) {
	source, _ := NewVariableTruthTable(8, 6, 0)
	before := source.Words()
	projected, e := source.ProjectVariables([]int{6, 2, 0})
	if e != nil || projected.Variables() != 3 {
		t.Fatal(e)
	}
	expected, _ := NewVariableTruthTable(3, 0, 0)
	if !projected.Equal(expected) {
		t.Fatal("projection direction")
	}
	if result, e := source.ProjectVariables([]int{0, 1}); result != nil || !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
	for _, keep := range [][]int{{6, 6}, {-1, 6}, {6, 8}} {
		if result, e := source.ProjectVariables(keep); result != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, mapping := range [][]int{{0, 0, 2, 3, 4, 5, 6, 7}, {0, 1, 2, 3, 4, 5, 6, 8}, {-1, 1, 2, 3, 4, 5, 6, 7}, nil} {
		if result, e := source.ExpandVariables(8, mapping); result != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if result, e := source.Stretch(7); result != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if result, e := source.Stretch(37); result != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	limited, _ := NewVariableTruthTable(1, 0, 1)
	if result, e := limited.Stretch(7); result != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if !slices.Equal(before, source.Words()) {
		t.Fatal("source mutation")
	}
	identity, e := source.ProjectVariables([]int{0, 1, 2, 3, 4, 5, 6, 7})
	if e != nil || !identity.Equal(source) {
		t.Fatal(e)
	}
	identity.Set(0, true)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("identity alias")
	}
	var nilTruth *TruthTable
	if result, e := nilTruth.CompactSupport(); result.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzTruthSupport(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 64 {
			data = data[:64]
		}
		variables := int(flags % 8)
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
		reduced, e := source.CompactSupport()
		if e != nil {
			t.Fatal(e)
		}
		restored, e := reduced.Table.ExpandVariables(variables, reduced.SourceVariables)
		if e != nil || !restored.Equal(source) {
			t.Fatal("round trip", variables, reduced.SourceVariables, e)
		}
		for assignment := 0; assignment < n; assignment++ {
			target := 0
			for i, original := range reduced.SourceVariables {
				if assignment&(1<<uint(original)) != 0 {
					target |= 1 << uint(i)
				}
			}
			got, e := reduced.Table.Get(uint64(target))
			if e != nil || got != model[assignment] {
				t.Fatal("projection", assignment)
			}
		}
		targetVariables := variables + int(data[0]%4)
		positions := make([]int, targetVariables)
		for i := range positions {
			positions[i] = i
		}
		for i := len(positions) - 1; i > 0; i-- {
			j := int(data[i%len(data)]) % (i + 1)
			positions[i], positions[j] = positions[j], positions[i]
		}
		mapping := slices.Clone(positions[:variables])
		mappingBefore := slices.Clone(mapping)
		expanded, e := source.ExpandVariables(targetVariables, mapping)
		if e != nil {
			t.Fatal(e)
		}
		for assignment := 0; assignment < 1<<uint(targetVariables); assignment++ {
			old := 0
			for i, target := range mapping {
				if assignment&(1<<uint(target)) != 0 {
					old |= 1 << uint(i)
				}
			}
			got, e := expanded.Get(uint64(assignment))
			if e != nil || got != model[old] {
				t.Fatal("embedding", variables, mapping, assignment, old)
			}
		}
		if !slices.Equal(mapping, mappingBefore) || !slices.Equal(source.Words(), words) {
			t.Fatal("argument mutation")
		}
	})
}
