package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthAdaptiveEnumerationNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_adaptive_enum.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, High, Phase int
		Source, Result               []uint64
		Expanded                     []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 288 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{HighSymmetry: row.High != 0})
		if e != nil {
			t.Fatal(e)
		}
		phase := 0
		for target, label := range row.Expanded {
			if r.Permutation[label] != target {
				t.Fatal("permutation", index, r.Permutation, row.Expanded)
			}
			if r.InputPhases[label] {
				phase |= 1 << uint(target)
			}
		}
		if r.OutputPhase {
			phase |= 1 << uint(row.Variables)
		}
		if phase != row.Phase || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(index, row.Variables, row.Mode, row.High, phase, row.Phase, r.Table.Words(), row.Result)
		}
		expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
		if e != nil || !expected.Equal(r.Table) {
			t.Fatal("forward metadata", index, e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}

func TestTruthAdaptivePermutationTraversal(t *testing.T) {
	for size := 2; size <= 8; size++ {
		source, _ := NewConstantTruthTable(size, false, 0)
		s := &truthAdaptiveState{table: source, order: make([]int, size), position: make([]int, size), representatives: make([]int, size), links: make([]int, size+1), groups: []TruthVariableGroup{{0, size}}}
		directions := make([]int, size)
		for i := range s.links {
			s.links[i] = -1
		}
		for i := 0; i < size; i++ {
			s.order[i] = i
			s.position[i] = i
			s.representatives[i] = i
			directions[i] = -1
		}
		seen := map[string]bool{}
		count := 0
		for {
			key := make([]byte, size)
			for i, v := range s.representatives {
				key[i] = byte(v)
			}
			if seen[string(key)] {
				t.Fatal("duplicate", size, s.representatives)
			}
			seen[string(key)] = true
			count++
			before := slices.Clone(s.representatives)
			if !s.nextPermutation(directions) {
				break
			}
			changed := []int{}
			for i, v := range s.representatives {
				if v != before[i] {
					changed = append(changed, i)
				}
			}
			if len(changed) != 2 || changed[1] != changed[0]+1 {
				t.Fatal("nonadjacent step", changed)
			}
		}
		factorial := 1
		for i := 2; i <= size; i++ {
			factorial *= i
		}
		if count != factorial {
			t.Fatal(size, count, factorial)
		}
		for i, v := range s.representatives {
			if v != i || s.order[i] != i || directions[i] != -1 {
				t.Fatal("terminal restoration", size, s.representatives, directions)
			}
		}
	}
}
func TestTruthAdaptiveEnumerationLimits(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	for _, tc := range []struct {
		options TruthAdaptiveEnumerationOptions
		want    error
	}{
		{TruthAdaptiveEnumerationOptions{MaxConfigurations: -1}, ErrDimension},
		{TruthAdaptiveEnumerationOptions{MaxConfigurations: 1}, ErrLimit},
		{TruthAdaptiveEnumerationOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		r, e := source.EnumerateAdaptiveGroups(tc.options)
		if r.Table != nil || !errors.Is(e, tc.want) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthTable)(nil).EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	calls := 0
	r, e := source.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{Stop: func() bool { calls++; return calls == 4 }})
	if r.Table != nil || !errors.Is(e, ErrCancelled) {
		t.Fatal(r, e)
	}
	r, e = source.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{HighSymmetry: true})
	if e != nil || r.Configurations != 2 || r.Permutations != 1 {
		t.Fatal(r, e)
	}
	expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
	if e != nil || !expected.Equal(r.Table) {
		t.Fatal("wide metadata", e)
	}
	r.Table.Set(0, false)
	r.Permutation[0] = 99
	r.InputPhases[0] = true
	fresh, e := source.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{HighSymmetry: true})
	if e != nil || fresh.Permutation[0] == 99 || fresh.InputPhases[0] || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	for _, value := range []bool{false, true} {
		source, _ := NewConstantTruthTable(0, value, 0)
		r, e := source.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{})
		if e != nil || !r.Table.IsConstant(false) || r.OutputPhase != value || r.Configurations != 1 || r.Permutations != 1 {
			t.Fatal("zero input", r, e)
		}
	}
}

// Enumerate the same group orbit recursively, independently of Johnson-Trotter
// directions, Gray traversal and carried phase state. Compare minimum words;
// native fixtures separately establish first-encounter metadata tie behavior.
func checkTruthAdaptiveGroupOrbit(t *testing.T, source *TruthTable, high bool) {
	t.Helper()
	reduction, e := source.ReduceInputSymmetry(high)
	if e != nil {
		t.Fatal(e)
	}
	actual, e := source.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{HighSymmetry: high})
	if e != nil {
		t.Fatal(e)
	}
	labels := slices.Clone(reduction.Representatives)
	minimum := ^uint64(0)
	visits := 0
	var visit func(int)
	visit = func(index int) {
		if index < len(reduction.Groups) {
			group := reduction.Groups[index]
			var permute func(int)
			permute = func(offset int) {
				if offset == group.Start+group.Size {
					visit(index + 1)
					return
				}
				for j := offset; j < group.Start+group.Size; j++ {
					labels[offset], labels[j] = labels[j], labels[offset]
					permute(offset + 1)
					labels[offset], labels[j] = labels[j], labels[offset]
				}
			}
			permute(group.Start)
			return
		}
		order := []int{}
		for _, head := range labels {
			for v := head; v >= 0; v = reduction.Links[v] {
				order = append(order, v)
			}
		}
		for v := reduction.Links[source.variables]; v >= 0; v = reduction.Links[v] {
			order = append(order, v)
		}
		phaseInputs := 0
		if !reduction.PhasesDetermined && len(reduction.Groups) > 0 {
			phaseInputs = reduction.Groups[0].Size
		}
		for mask := 0; mask < 1<<uint(phaseInputs); mask++ {
			sourceMask := 0
			for i := 0; i < phaseInputs; i++ {
				if mask&(1<<uint(i)) != 0 {
					for v := labels[i]; v >= 0; v = reduction.Links[v] {
						if reduction.PhaseWeights[v] != 0 {
							sourceMask ^= 1 << uint(v)
						}
					}
				}
			}
			word := uint64(0)
			for x := 0; x < 64; x++ {
				old := sourceMask
				for target, v := range order {
					if x&(1<<uint(target)) != 0 {
						old ^= 1 << uint(v)
					}
				}
				if reduction.Table.words[0]&(uint64(1)<<uint(old)) != 0 {
					word |= uint64(1) << uint(x)
				}
			}
			if word < minimum {
				minimum = word
			}
			visits++
		}
	}
	visit(0)
	if actual.Table.words[0] != minimum || actual.Configurations != visits {
		t.Fatal("group orbit", source.variables, high, actual.Table.Words(), minimum, actual.Configurations, visits)
	}
	expected, e := source.TransformNPN(actual.Permutation, actual.InputPhases, actual.OutputPhase)
	if e != nil || !expected.Equal(actual.Table) {
		t.Fatal("configuration", e)
	}
}
func TestTruthAdaptiveEnumerationIndependentOrbit(t *testing.T) {
	state := uint64(71)
	for variables := 0; variables <= 5; variables++ {
		for sample := 0; sample < 24; sample++ {
			state = state*6364136223846793005 + 1
			word := state
			if variables < 6 {
				width := 1 << uint(variables)
				mask := uint64(1)<<uint(width) - 1
				word &= mask
				for shift := width; shift < 64; shift *= 2 {
					word |= word << uint(shift)
				}
			}
			source, e := TruthTableFromWords(variables, []uint64{word}, 0)
			if e != nil {
				t.Fatal(e)
			}
			checkTruthAdaptiveGroupOrbit(t, source, sample&1 != 0)
		}
	}
}
func FuzzTruthAdaptiveEnumeration(f *testing.F) {
	f.Add(uint8(4), uint64(0x96254718abcdef31))
	f.Add(uint8(128), uint64(0))
	f.Fuzz(func(t *testing.T, flags uint8, word uint64) {
		variables := int(flags % 6)
		width := 1 << uint(variables)
		word &= (uint64(1) << uint(width)) - 1
		for shift := width; shift < 64; shift *= 2 {
			word |= word << uint(shift)
		}
		source, e := TruthTableFromWords(variables, []uint64{word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		checkTruthAdaptiveGroupOrbit(t, source, flags&128 != 0)
		if source.words[0] != word {
			t.Fatal("source changed")
		}
	})
}
