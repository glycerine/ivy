package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSymmetryNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_symmetry.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, IncludePhase, Code int
		Source                        []uint64
		First, Second                 []int
		Phases                        []bool
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 3664 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		code, e := source.GroupSymmetry(row.First, row.Second, row.Phases, row.IncludePhase != 0)
		if e != nil || int(code) != row.Code {
			t.Fatal(index, row, code, e)
		}
		if len(row.First) == 1 {
			code, e = source.InputSymmetry(row.First[0], row.Second[0], row.IncludePhase != 0)
			if e != nil || int(code) != row.Code {
				t.Fatal("single", index, code, e)
			}
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthSymmetryKindsAndDimensions(t *testing.T) {
	// AND, x & !y, XOR, x: direct, phased, both, neither.
	for _, tc := range []struct {
		word uint64
		want TruthSymmetry
	}{{0x8888888888888888, TruthSymmetryDirect}, {0x2222222222222222, TruthSymmetryPhased}, {0x6666666666666666, TruthSymmetryBoth}, {0xaaaaaaaaaaaaaaaa, TruthSymmetryNone}} {
		table, e := TruthTableFromWords(2, []uint64{tc.word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		code, e := table.InputSymmetry(0, 1, true)
		if e != nil || code != tc.want {
			t.Fatal(tc, code, e)
		}
		direct, e := table.InputSymmetry(1, 0, false)
		if e != nil || direct != tc.want&TruthSymmetryDirect {
			t.Fatal(tc, direct, e)
		}
	}
	wide, _ := NewVariableTruthTable(18, 17, 0)
	if code, e := wide.InputSymmetry(16, 17, true); e != nil || code != TruthSymmetryNone {
		t.Fatal("wide nonconstant", code, e)
	}
	if code, e := wide.GroupSymmetry([]int{0, 5, 6}, []int{17, 16, 10}, nil, true); e != nil || code != TruthSymmetryNone {
		t.Fatal("wide group", code, e)
	}
	table, _ := NewConstantTruthTable(18, true, 0)
	if code, e := table.GroupSymmetry([]int{0, 5, 6}, []int{17, 16, 10}, nil, true); e != nil || code != TruthSymmetryBoth {
		t.Fatal(code, e)
	}
	for _, tc := range []struct {
		a, b   []int
		phases []bool
	}{
		{nil, nil, nil}, {[]int{0}, nil, nil}, {[]int{-1}, []int{1}, nil}, {[]int{0}, []int{18}, nil},
		{[]int{0}, []int{0}, nil}, {[]int{0, 0}, []int{1, 2}, nil}, {[]int{0}, []int{1}, []bool{true}},
	} {
		if _, e := table.GroupSymmetry(tc.a, tc.b, tc.phases, true); !errors.Is(e, ErrDimension) {
			t.Fatal(tc, e)
		}
	}
	if _, e := (*TruthTable)(nil).InputSymmetry(0, 1, true); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzTruthSymmetry(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(128), []byte{255, 0})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 256 {
			data = data[:256]
		}
		variables := 2 + int(flags%11)
		n := 1 << uint(variables)
		model := make([]bool, n)
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		for i := range model {
			model[i] = data[(i/8)%len(data)]&(1<<uint(i%8)) != 0
		}
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
		for i := range order {
			order[i] = i
		}
		for i := variables - 1; i > 0; i-- {
			j := int(data[i%len(data)]) % (i + 1)
			order[i], order[j] = order[j], order[i]
		}
		size := 1 + int(data[0])%(variables/2)
		a, b := order[:size], order[size:2*size]
		phases := make([]bool, variables)
		for i := range phases {
			phases[i] = data[i%len(data)]&1 != 0
		}
		if flags&64 != 0 {
			phases = nil
		}
		include := flags&128 != 0
		want := TruthSymmetryBoth
		for assignment := 0; assignment < n; assignment++ {
			swap := func(x int) int {
				for i, v := range a {
					if ((x>>uint(v))^(x>>uint(b[i])))&1 != 0 {
						x ^= (1 << uint(v)) | (1 << uint(b[i]))
					}
				}
				return x
			}
			if model[assignment] != model[swap(assignment)] {
				want &^= TruthSymmetryDirect
			}
			if include {
				old := assignment
				for _, group := range [][]int{a, b} {
					for _, v := range group {
						if phases == nil || phases[v] {
							old ^= 1 << uint(v)
						}
					}
				}
				if model[assignment] != model[swap(old)] {
					want &^= TruthSymmetryPhased
				}
			}
		}
		if !include {
			want &^= TruthSymmetryPhased
		}
		code, e := source.GroupSymmetry(a, b, phases, include)
		if e != nil || code != want {
			t.Fatal(variables, a, b, phases, include, code, want, e)
		}
		if size == 1 && phases == nil {
			code, e = source.InputSymmetry(a[0], b[0], include)
			if e != nil || code != want {
				t.Fatal("single", code, want, e)
			}
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source changed")
		}
	})
}
