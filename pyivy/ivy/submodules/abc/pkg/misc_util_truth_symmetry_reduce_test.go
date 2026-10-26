package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSymmetryReductionNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_symmetry_reduce.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, High, Phase, Phased, Supported int
		Source, Result                                  []uint64
		Representatives, Expanded, Links, Weights       []int
		Groups                                          []TruthVariableGroup
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 288 {
		t.Fatal(len(rows))
	}
	highDifference := false
	for i, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.ReduceInputSymmetry(row.High != 0)
		if e != nil {
			t.Fatal(e)
		}
		phase := 0
		for v, b := range r.InputPhases {
			if b {
				phase |= 1 << uint(v)
			}
		}
		if r.OutputPhase {
			phase |= 1 << uint(row.Variables)
		}
		if phase != row.Phase || r.PhasesDetermined != (row.Phased != 0) || r.SupportedInputs != row.Supported || !slices.Equal(r.Table.Words(), row.Result) || !slices.Equal(r.Representatives, row.Representatives) || !slices.Equal(r.ExpandedOrder, row.Expanded) || !slices.Equal(r.Links, row.Links) || !slices.Equal(r.PhaseWeights, row.Weights) || !slices.Equal(r.Groups, row.Groups) {
			t.Fatal(i, row, r)
		}
		expected, e := source.TransformNPN(nil, r.InputPhases, r.OutputPhase)
		if e != nil || !expected.Equal(r.Table) {
			t.Fatal("phase metadata", e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
		if row.High != 0 && len(rows[i-1].Representatives) != len(row.Representatives) {
			highDifference = true
		}
	}
	if !highDifference {
		t.Fatal("native cases do not exercise higher-order merging")
	}
}
func TestTruthSymmetryReductionBoundaries(t *testing.T) {
	if r, e := (*TruthTable)(nil).ReduceInputSymmetry(true); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, value := range []bool{false, true} {
		source, _ := NewConstantTruthTable(0, value, 0)
		r, e := source.ReduceInputSymmetry(true)
		if e != nil || !r.Table.IsConstant(false) || r.OutputPhase != value || len(r.Representatives) != 0 || len(r.ExpandedOrder) != 0 || len(r.Groups) != 0 || !slices.Equal(r.Links, []int{-1}) {
			t.Fatal(r, e)
		}
	}
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	r, e := source.ReduceInputSymmetry(true)
	if e != nil || r.SupportedInputs != 1 || !slices.Equal(r.Representatives, []int{17}) || r.ExpandedOrder[0] != 17 || r.ExpandedOrder[1] != 16 || r.Links[18] != 16 {
		t.Fatal(r, e)
	}
	r.Table.Set(0, false)
	r.InputPhases[17] = false
	r.Representatives[0] = 0
	r.Links[18] = -1
	r.Groups[0].Size = 0
	r.PhaseWeights[17] = 0
	r.ExpandedOrder[0] = 0
	fresh, e := source.ReduceInputSymmetry(true)
	if e != nil || !slices.Equal(source.Words(), before) || fresh.Representatives[0] != 17 || fresh.Links[18] != 16 || fresh.PhaseWeights[17] != 1 || !fresh.InputPhases[17] || fresh.Groups[0].Size != 1 {
		t.Fatal("ownership", e)
	}
}
func FuzzTruthSymmetryReduction(f *testing.F) {
	f.Add(uint8(4), []byte{0x88, 0x66})
	f.Add(uint8(128), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 256 {
			data = data[:256]
		}
		variables := int(flags % 11)
		n := 1 << uint(variables)
		count, _ := TruthTableWordCount(variables)
		model := make([]bool, n)
		words := make([]uint64, count)
		for x := range model {
			model[x] = data[(x/8)%len(data)]&(1<<uint(x%8)) != 0
		}
		for bit := 0; bit < count*64; bit++ {
			if model[bit%n] {
				words[bit/64] |= uint64(1) << uint(bit%64)
			}
		}
		source, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.ReduceInputSymmetry(flags&128 != 0)
		if e != nil {
			t.Fatal(e)
		}
		mask := 0
		for v, b := range r.InputPhases {
			if b {
				mask |= 1 << uint(v)
			}
		}
		for x := range model {
			got, e := r.Table.Get(uint64(x))
			if e != nil || got != (model[x^mask] != r.OutputPhase) {
				t.Fatal("phase transform", x, e)
			}
		}
		supported := make([]bool, variables)
		total := 0
		for v := range supported {
			for x := range model {
				if model[x] != model[x^(1<<uint(v))] {
					supported[v] = true
					total++
					break
				}
			}
		}
		if total != r.SupportedInputs || len(r.ExpandedOrder) != variables {
			t.Fatal("support")
		}
		seen := make([]bool, variables)
		for i, v := range r.ExpandedOrder {
			if v < 0 || v >= variables || seen[v] || supported[v] != (i < r.SupportedInputs) {
				t.Fatal("expansion", r)
			}
			seen[v] = true
		}
		rebuilt := []int{}
		heads := append(slices.Clone(r.Representatives), r.Links[variables])
		for _, head := range heads {
			steps := 0
			for v := head; v >= 0; v = r.Links[v] {
				if v >= variables || steps >= variables {
					t.Fatal("chain cycle")
				}
				rebuilt = append(rebuilt, v)
				steps++
			}
		}
		if !slices.Equal(rebuilt, r.ExpandedOrder) {
			t.Fatal("chains")
		}
		start := 0
		for _, group := range r.Groups {
			if group.Start != start || group.Size <= 0 {
				t.Fatal("groups")
			}
			start += group.Size
		}
		if start != len(r.Representatives) {
			t.Fatal("partition")
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source changed")
		}
	})
}
