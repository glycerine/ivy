package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthTiedGroupsNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_groups.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, Phase, Phased, Ones int
		Source, Result                       []uint64
		Order, Counts                        []int
		Groups                               []TruthVariableGroup
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 96 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CreateTiedGroups()
		if e != nil {
			t.Fatal(e)
		}
		phase := 0
		for i, v := range r.InputPhases {
			if v {
				phase |= 1 << uint(i)
			}
		}
		if r.OutputPhase {
			phase |= 1 << uint(row.Variables)
		}
		if !slices.Equal(r.Table.Words(), row.Result) || !slices.Equal(r.InputOrder, row.Order) || !slices.Equal(r.ZeroCofactorCounts, row.Counts) || !slices.Equal(r.Groups, row.Groups) || r.PackedOnes != row.Ones || r.PhasesDetermined != (row.Phased != 0) || phase != row.Phase {
			t.Fatal(row.Variables, row.Mode, r, row)
		}
		expected, e := source.TransformNPN(nil, r.InputPhases, r.OutputPhase)
		if e != nil || !expected.Equal(r.Table) {
			t.Fatal("phase metadata", e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthTiedGroupStableSplit(t *testing.T) {
	order := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	groups := []TruthVariableGroup{{0, 2}, {2, 6}, {8, 2}}
	coefficients := []int{3, 1, 2, 1, 3, 2}
	got := truthSplitTiedGroup(order, groups, 1, coefficients)
	if !slices.Equal(order, []int{9, 8, 6, 4, 5, 2, 7, 3, 1, 0}) || !slices.Equal(coefficients, []int{1, 1, 2, 2, 3, 3}) || !slices.Equal(got, []TruthVariableGroup{{0, 2}, {2, 2}, {4, 2}, {6, 2}, {8, 2}}) {
		t.Fatal(order, coefficients, got)
	}
	if !slices.Equal(groups, []TruthVariableGroup{{0, 2}, {2, 6}, {8, 2}}) {
		t.Fatal("group aliases")
	}
}
func TestTruthTiedGroupsBoundaries(t *testing.T) {
	if r, e := (*TruthTable)(nil).CreateTiedGroups(); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, value := range []bool{false, true} {
		table, _ := NewConstantTruthTable(0, value, 0)
		r, e := table.CreateTiedGroups()
		if e != nil || len(r.Groups) != 0 || len(r.InputOrder) != 0 || r.OutputPhase != value || !r.Table.IsConstant(false) || r.PackedOnes != 0 || r.PhasesDetermined {
			t.Fatal(r, e)
		}
	}
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	r, e := source.CreateTiedGroups()
	if e != nil || r.OutputPhase || !r.InputPhases[17] || r.InputOrder[17] != 17 || len(r.Groups) != 2 || r.Groups[0].Size != 17 || r.PhasesDetermined {
		t.Fatal(r, e)
	}
	expected, _ := source.Not().CreateTiedGroups()
	if !expected.Table.Equal(r.Table) {
		t.Fatal("phase normalized variable")
	}
	r.Table.Set(0, false)
	r.InputOrder[0] = 99
	r.InputPhases[0] = true
	r.Groups[0].Size = 1
	r.ZeroCofactorCounts[0] = 0
	fresh, e := source.CreateTiedGroups()
	if e != nil || !slices.Equal(source.Words(), before) || fresh.InputOrder[0] != 0 || fresh.InputPhases[0] || fresh.Groups[0].Size != 17 || fresh.ZeroCofactorCounts[0] == 0 {
		t.Fatal("ownership", e)
	}
}
func FuzzTruthTiedGroups(f *testing.F) {
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
		model := make([]bool, n)
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
		r, e := source.CreateTiedGroups()
		if e != nil {
			t.Fatal(e)
		}
		ones := 0
		for _, v := range model {
			if v {
				ones++
			}
		}
		output := ones > n/2 // n==1 requires its own exact half test.
		if variables == 0 {
			output = model[0]
		}
		if output {
			ones = n - ones
		}
		factor := max(1, 64/n)
		if r.OutputPhase != output || r.PackedOnes != ones*factor {
			t.Fatal("output", r)
		}
		phases := make([]bool, variables)
		counts := make([]int, variables)
		for v := 0; v < variables; v++ {
			for x, b := range model {
				if x&(1<<uint(v)) == 0 && b != output {
					counts[v]++
				}
			}
			phases[v] = counts[v] < ones-counts[v]
			if phases[v] {
				counts[v] = ones - counts[v]
			}
			counts[v] *= factor
		}
		if !slices.Equal(phases, r.InputPhases) {
			t.Fatal("phases")
		}
		order := make([]int, variables)
		for i := range order {
			order[i] = i
		}
		// Independent stable insertion by adjacent exchanges.
		for i := 1; i < variables; i++ {
			for j := i; j > 0 && counts[order[j-1]] > counts[order[j]]; j-- {
				order[j], order[j-1] = order[j-1], order[j]
			}
		}
		if !slices.Equal(order, r.InputOrder) {
			t.Fatal("stable order", order, r.InputOrder)
		}
		mask := 0
		for i, v := range phases {
			if v {
				mask |= 1 << uint(i)
			}
		}
		for x := 0; x < n; x++ {
			got, e := r.Table.Get(uint64(x))
			if e != nil || got != (model[x^mask] != output) {
				t.Fatal("truth", x, e)
			}
		}
		start := 0
		for _, group := range r.Groups {
			if group.Start != start || group.Size <= 0 {
				t.Fatal("partition")
			}
			for j := start; j < start+group.Size; j++ {
				if r.ZeroCofactorCounts[j] != counts[order[j]] || counts[order[j]] != counts[order[start]] {
					t.Fatal("signature")
				}
			}
			start += group.Size
			if start < variables && counts[order[start]] == counts[order[start-1]] {
				t.Fatal("unmerged tie")
			}
		}
		if start != variables || r.PhasesDetermined != (variables > 0 && counts[order[0]]*2 != ones*factor) {
			t.Fatal("groups")
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source changed")
		}
	})
}
