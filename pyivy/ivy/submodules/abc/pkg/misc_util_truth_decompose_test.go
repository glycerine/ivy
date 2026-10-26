package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func evaluateOutputDecomposition(operator TruthOutputOperator, x, g bool) bool {
	switch operator {
	case TruthOutputAnd:
		return x && g
	case TruthOutputNotAnd:
		return !x && g
	case TruthOutputNotOr:
		return !x || g
	case TruthOutputOr:
		return x || g
	case TruthOutputXor:
		return x != g
	}
	panic("invalid operator")
}
func TestTruthOutputDecompositionNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/util_truth_decompose.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, Kind int
		Source, Remainder     []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 36 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, found, e := source.DecomposeVariable(row.Variables - 1)
		if e != nil || found != (row.Kind >= 0) {
			t.Fatal(e)
		}
		if found && (int(r.Operator) != row.Kind || r.Variable != row.Variables-1 || !slices.Equal(r.Remainder.Words(), row.Remainder)) {
			t.Fatal(row, r)
		}
		if found {
			for assignment := 0; assignment < 1<<uint(row.Variables); assignment++ {
				g, e := r.Remainder.Get(uint64(assignment))
				if e != nil {
					t.Fatal(e)
				}
				want, _ := source.Get(uint64(assignment))
				if evaluateOutputDecomposition(r.Operator, assignment&(1<<uint(r.Variable)) != 0, g) != want {
					t.Fatal("reconstruction")
				}
			}
		}
	}
}
func TestTruthOutputDecompositionUnusedAndLarge(t *testing.T) {
	constant, _ := NewConstantTruthTable(3, true, 0)
	if r, found, e := constant.DecomposeVariable(1); r.Remainder != nil || found || e != nil {
		t.Fatal(found, e)
	}
	x, _ := NewVariableTruthTable(13, 12, 0)
	y, _ := NewVariableTruthTable(13, 0, 0)
	source, _ := x.Xor(y, false)
	before := source.Words()
	r, found, e := source.DecomposeVariable(12)
	if e != nil || !found || r.Operator != TruthOutputXor || !r.Remainder.Equal(y) {
		t.Fatal(found, e)
	}
	r.Remainder.Set(0, true)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("alias")
	}
	if _, _, e := source.DecomposeVariable(13); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, _, e := (*TruthTable)(nil).DecomposeVariable(0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	// Literal functions satisfy several cofactor rules; C's first matching rule wins.
	literal, _ := NewVariableTruthTable(1, 0, 0)
	r, found, e = literal.DecomposeVariable(0)
	if e != nil || !found || r.Operator != TruthOutputAnd || !r.Remainder.IsConstant(true) {
		t.Fatal(r, found, e)
	}
	r, found, e = literal.Not().DecomposeVariable(0)
	if e != nil || !found || r.Operator != TruthOutputNotAnd {
		t.Fatal(r, found, e)
	}
}
func FuzzTruthOutputDecomposition(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 96 {
			data = data[:96]
		}
		variables := 1 + int(flags%9)
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
		variable := int(data[0]) % variables
		same, zero0, zero1, one0, one1, opposite := true, true, true, true, true, true
		for assignment := 0; assignment < n; assignment++ {
			base := assignment &^ (1 << uint(variable))
			a, b := model[base], model[base|(1<<uint(variable))]
			same = same && a == b
			zero0 = zero0 && !a
			zero1 = zero1 && !b
			one0 = one0 && a
			one1 = one1 && b
			opposite = opposite && (a != b)
		}
		want := -1
		if !same {
			switch {
			case zero0:
				want = 0
			case zero1:
				want = 1
			case one0:
				want = 2
			case one1:
				want = 3
			case opposite:
				want = 4
			}
		}
		r, found, e := source.DecomposeVariable(variable)
		if e != nil || found != (want >= 0) {
			t.Fatal(want, found, e)
		}
		if !found {
			if r.Remainder != nil {
				t.Fatal("partial")
			}
			return
		}
		if int(r.Operator) != want {
			t.Fatal(r.Operator, want)
		}
		used, e := r.Remainder.HasVariable(variable)
		if e != nil || used {
			t.Fatal("remaining variable")
		}
		for assignment := 0; assignment < n; assignment++ {
			g, e := r.Remainder.Get(uint64(assignment))
			if e != nil || evaluateOutputDecomposition(r.Operator, assignment&(1<<uint(variable)) != 0, g) != model[assignment] {
				t.Fatal("formula", variable, assignment)
			}
		}
		if !slices.Equal(words, source.Words()) {
			t.Fatal("source mutation")
		}
	})
}
