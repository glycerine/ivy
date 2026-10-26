package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthShiftedCoefficientsNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_scc.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode int
		Total           uint64
		Source, Zero    []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 104 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.ShiftedCofactorCoefficients()
		if e != nil || r.Total != row.Total || !slices.Equal(r.Zero, row.Zero) {
			t.Fatal(row, r, e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthShiftedGroupCoefficientsNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_group_scc.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, High        int
		Source, Coefficients, Sorted []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 288 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		reduction, e := source.ReduceInputSymmetry(row.High != 0)
		if e != nil {
			t.Fatal(e)
		}
		state := newTruthAdaptiveState(reduction)
		before := state.table.Words()
		labels := slices.Clone(state.representatives)
		got, e := state.shiftedGroupCoefficients(false)
		if e != nil || !slices.Equal(got, row.Coefficients) {
			t.Fatal(row, got, e)
		}
		got, e = state.shiftedGroupCoefficients(true)
		if e != nil || !slices.Equal(got, row.Sorted) {
			t.Fatal(row, got, e)
		}
		if !slices.Equal(state.table.Words(), before) || !slices.Equal(state.representatives, labels) {
			t.Fatal("sort mutated state")
		}
	}
}
func TestTruthShiftedCoefficientBounds(t *testing.T) {
	if r, e := (*TruthTable)(nil).ShiftedCofactorCoefficients(); r.Zero != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	power := func(n int) uint64 {
		value := uint64(1)
		for i := 0; i < n; i++ {
			value *= 3
		}
		return value
	}
	for _, variables := range []int{0, 1, 5, 6, 7, 18, 20} {
		source, _ := NewConstantTruthTable(variables, true, 0)
		r, e := source.ShiftedCofactorCoefficients()
		if e != nil || r.Total != power(max(6, variables)) {
			t.Fatal(variables, r, e)
		}
		for _, value := range r.Zero {
			if value != power(max(6, variables)-1) {
				t.Fatal("zero cofactor", variables, value)
			}
		}
		if variables == 20 && r.Total <= 1<<31 {
			t.Fatal("wide signature truncated")
		}
		if variables > 0 {
			r.Zero[0] = 0
			fresh, e := source.ShiftedCofactorCoefficients()
			if e != nil || fresh.Zero[0] == 0 {
				t.Fatal("ownership", e)
			}
		}
	}
}
func FuzzTruthShiftedCoefficients(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 256 {
			data = data[:256]
		}
		variables := int(width % 13)
		n := 1 << uint(variables)
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		for x := 0; x < count*64; x++ {
			i := x % n
			if data[(i/8)%len(data)]&(1<<uint(i%8)) != 0 {
				words[x/64] |= uint64(1) << uint(x%64)
			}
		}
		source, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.ShiftedCofactorCoefficients()
		if e != nil {
			t.Fatal(e)
		}
		total := uint64(0)
		zero := make([]uint64, variables)
		factor := uint64(1)
		for v := variables; v < 6; v++ {
			factor *= 3
		}
		for x := 0; x < n; x++ {
			if data[(x/8)%len(data)]&(1<<uint(x%8)) == 0 {
				continue
			}
			weight := factor
			for v := 0; v < variables; v++ {
				if x&(1<<uint(v)) != 0 {
					weight *= 2
				}
			}
			total += weight
			for v := range zero {
				if x&(1<<uint(v)) == 0 {
					zero[v] += weight
				}
			}
		}
		if r.Total != total || !slices.Equal(r.Zero, zero) {
			t.Fatal(variables, r, total, zero)
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source changed")
		}
	})
}
