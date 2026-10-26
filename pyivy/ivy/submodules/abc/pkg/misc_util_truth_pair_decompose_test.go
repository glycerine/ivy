package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func evaluateTruthPair(operator TruthPairOperator, a, b bool) bool {
	if operator == TruthPairXor {
		return a != b
	}
	return (a != (operator&1 != 0)) && (b != (operator&2 != 0))
}
func TestTruthPairDecompositionNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/util_truth_pair_decompose.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, First, Second, Mode, Kind int
		Source, Remainder                    uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 210 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, []uint64{row.Source}, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, found, e := source.DecomposeInputPair(row.First, row.Second)
		if e != nil || found != (row.Kind >= 0) {
			t.Fatal(e)
		}
		if !found {
			continue
		}
		if int(r.Operator) != row.Kind || !slices.Equal(r.Remainder.Words(), []uint64{row.Remainder}) {
			t.Fatal(row, r)
		}
		for assignment := 0; assignment < 1<<uint(row.Variables); assignment++ {
			a, b := assignment&(1<<uint(row.First)) != 0, assignment&(1<<uint(row.Second)) != 0
			target := assignment &^ (1 << uint(row.First))
			if evaluateTruthPair(r.Operator, a, b) {
				target |= 1 << uint(row.First)
			}
			got, e := r.Remainder.Get(uint64(target))
			want, _ := source.Get(uint64(assignment))
			if e != nil || got != want {
				t.Fatal(row, assignment)
			}
		}
	}
}
func TestTruthPairDecompositionWideAndDegenerate(t *testing.T) {
	a, _ := NewVariableTruthTable(13, 12, 0)
	b, _ := NewVariableTruthTable(13, 0, 0)
	c, _ := NewVariableTruthTable(13, 6, 0)
	pair, _ := a.And(b, false, true)
	source, _ := pair.Xor(c, false)
	before := source.Words()
	r, found, e := source.DecomposeInputPair(12, 0)
	if e != nil || !found || r.Operator != TruthPairNotSecondAnd {
		t.Fatal(found, e)
	}
	expected, _ := a.Xor(c, false)
	if !r.Remainder.Equal(expected) {
		t.Fatal("wide remainder")
	}
	r.Remainder.Set(0, true)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("alias")
	}
	constant, _ := NewConstantTruthTable(3, true, 0)
	r, found, e = constant.DecomposeInputPair(0, 1)
	if e != nil || !found || r.Operator != TruthPairAnd || !r.Remainder.IsConstant(true) {
		t.Fatal("precedence", found, e)
	}
	for _, indices := range [][2]int{{-1, 0}, {0, 13}, {0, 0}} {
		r, found, e := source.DecomposeInputPair(indices[0], indices[1])
		if r.Remainder != nil || found || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if r, found, e := (*TruthTable)(nil).DecomposeInputPair(0, 1); r.Remainder != nil || found || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzTruthPairDecomposition(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 96 {
			data = data[:96]
		}
		variables := 2 + int(flags%8)
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
		first := int(data[0]) % variables
		second := (first + 1 + int(data[len(data)-1])%(variables-1)) % variables
		cofs := [4][]bool{}
		for i := range cofs {
			cofs[i] = make([]bool, n)
		}
		for assignment := 0; assignment < n; assignment++ {
			base := assignment &^ (1 << uint(first)) &^ (1 << uint(second))
			for bits := 0; bits < 4; bits++ {
				old := base
				if bits&2 != 0 {
					old |= 1 << uint(first)
				}
				if bits&1 != 0 {
					old |= 1 << uint(second)
				}
				cofs[bits][assignment] = model[old]
			}
		}
		equal := func(a, b int) bool { return slices.Equal(cofs[a], cofs[b]) }
		want := -1
		switch {
		case equal(0, 1) && equal(0, 2):
			want = 0
		case equal(3, 0) && equal(3, 2):
			want = 1
		case equal(3, 0) && equal(3, 1):
			want = 2
		case equal(3, 1) && equal(3, 2):
			want = 3
		case equal(0, 3) && equal(1, 2):
			want = 4
		}
		r, found, e := source.DecomposeInputPair(first, second)
		if e != nil || found != (want >= 0) {
			t.Fatal(found, want, e)
		}
		if !found {
			if r.Remainder != nil {
				t.Fatal("partial")
			}
			return
		}
		if int(r.Operator) != want || r.First != first || r.Second != second {
			t.Fatal(r, want)
		}
		used, e := r.Remainder.HasVariable(second)
		if e != nil || used {
			t.Fatal("second input remains")
		}
		for assignment := 0; assignment < n; assignment++ {
			a, b := assignment&(1<<uint(first)) != 0, assignment&(1<<uint(second)) != 0
			target := assignment &^ (1 << uint(first))
			if evaluateTruthPair(r.Operator, a, b) {
				target |= 1 << uint(first)
			}
			got, e := r.Remainder.Get(uint64(target))
			if e != nil || got != model[assignment] {
				t.Fatal("reconstruction", first, second, assignment)
			}
		}
		if !slices.Equal(words, source.Words()) {
			t.Fatal("mutation")
		}
	})
}
