package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSmartRefinementNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_refine_smart.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, First, Mode, OnlyPermutation, RawConfiguration, Configuration int
		Source, RawResult, Result                                                []uint64
		Compare, CompareRev                                                      []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 1056 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		raw := source.refineAdjacentSmartCandidate(row.First, row.OnlyPermutation != 0)
		r, e := source.RefineAdjacentCofactorsSmart(row.First, row.OnlyPermutation != 0)
		if e != nil || int(raw.Configuration) != row.RawConfiguration || int(r.Configuration) != row.Configuration || !slices.Equal(raw.Table.Words(), row.RawResult) || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(row.Variables, row.First, row.Mode, row.OnlyPermutation, raw.Configuration, row.RawConfiguration, r.Configuration, row.Configuration, e)
		}
		i := 0
		for a := 0; a < 4; a++ {
			for b := a + 1; b < 4; b++ {
				forward, e := source.CompareAdjacentCofactors(row.First, a, b, false)
				if e != nil {
					t.Fatal(e)
				}
				reverse, e := source.CompareAdjacentCofactors(row.First, a, b, true)
				if e != nil || forward != row.Compare[i] || reverse != row.CompareRev[i] {
					t.Fatal("pair comparison", row.Variables, row.First, a, b, forward, reverse, e)
				}
				i++
			}
		}
	}
}
func TestTruthSmartRefinementWideAndDimensions(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	r, e := source.RefineAdjacentCofactorsSmart(16, false)
	if e != nil {
		t.Fatal(e)
	}
	order := make([]int, 18)
	phases := make([]bool, 18)
	for i := range order {
		order[i] = i
	}
	phases[16], phases[17] = r.Configuration&1 != 0, r.Configuration&2 != 0
	if r.Configuration&4 != 0 {
		order[16], order[17] = order[17], order[16]
	}
	expected, e := source.TransformNPN(order, phases, false)
	if e != nil || !expected.Equal(r.Table) {
		t.Fatal("wide configuration", e)
	}
	r.Table.Set(0, false)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("alias")
	}
	for _, args := range [][3]int{{-1, 0, 1}, {17, 0, 1}, {0, -1, 1}, {0, 0, 4}} {
		if _, e := source.CompareAdjacentCofactors(args[0], args[1], args[2], false); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if r, e := source.RefineAdjacentCofactorsSmart(17, false); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if r, e := (*TruthTable)(nil).RefineAdjacentCofactorsSmart(0, false); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if cmp, e := source.CompareAdjacentCofactors(16, 2, 2, true); e != nil || cmp != 0 {
		t.Fatal(cmp, e)
	}
}
func FuzzTruthSmartRefinement(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(128), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 96 {
			data = data[:96]
		}
		variables := 2 + int(flags%11)
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
		first := int(data[0]) % (variables - 1)
		cofs := [4][]uint64{}
		for code := 0; code < 4; code++ {
			cofs[code] = make([]uint64, count)
			for bit := 0; bit < count*64; bit++ {
				old := (bit % n) &^ (1 << uint(first)) &^ (1 << uint(first+1))
				if code&1 != 0 {
					old |= 1 << uint(first)
				}
				if code&2 != 0 {
					old |= 1 << uint(first+1)
				}
				if model[old] {
					cofs[code][bit/64] |= uint64(1) << uint(bit%64)
				}
			}
		}
		for a := 0; a < 4; a++ {
			for b := 0; b < 4; b++ {
				for _, reverse := range []bool{false, true} {
					want := 0
					for i := 0; i < count; i++ {
						index := i
						if reverse {
							index = count - 1 - i
						}
						if cofs[a][index] < cofs[b][index] {
							want = -1
							break
						}
						if cofs[a][index] > cofs[b][index] {
							want = 1
							break
						}
					}
					got, e := source.CompareAdjacentCofactors(first, a, b, reverse)
					if e != nil || got != want {
						t.Fatal("cofactor order", variables, first, a, b, reverse, got, want, e)
					}
				}
			}
		}
		only := flags&128 != 0
		raw := source.refineAdjacentSmartCandidate(first, only)
		r, e := source.RefineAdjacentCofactorsSmart(first, only)
		if e != nil {
			t.Fatal(e)
		}
		for _, candidate := range []TruthPairRefinement{raw, r} {
			if only && candidate.Configuration != 0 && candidate.Configuration != 4 {
				t.Fatal("swap only")
			}
			for assignment := 0; assignment < n; assignment++ {
				old := assignment
				if candidate.Configuration&4 != 0 {
					a, b := (assignment>>uint(first))&1, (assignment>>uint(first+1))&1
					if a != b {
						old ^= (1 << uint(first)) | (1 << uint(first+1))
					}
				}
				if candidate.Configuration&1 != 0 {
					old ^= 1 << uint(first)
				}
				if candidate.Configuration&2 != 0 {
					old ^= 1 << uint(first+1)
				}
				got, e := candidate.Table.Get(uint64(assignment))
				if e != nil || got != model[old] {
					t.Fatal("configuration", variables, first, candidate.Configuration, assignment)
				}
			}
		}
		cmp, e := raw.Table.Compare(source, true)
		if e != nil {
			t.Fatal(e)
		}
		if cmp > 0 {
			if r.Configuration != 0 || !r.Table.Equal(source) {
				t.Fatal("rollback")
			}
		} else if r.Configuration != raw.Configuration || !r.Table.Equal(raw.Table) {
			t.Fatal("accepted candidate")
		}
		cmp, e = r.Table.Compare(source, true)
		if e != nil || cmp > 0 {
			t.Fatal("worse result")
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source mutation")
		}
	})
}
