package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSCCEnumerationNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_scc_enum.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, High, Kind, Split, Phase int
		Source, Result                            []uint64
		Order                                     []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 1080 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{PhaseMode: TruthSCCPhaseMode(row.Kind), HighSymmetry: row.High != 0, SplitByCofactors: row.Split != 0})
		if e != nil {
			t.Fatal(e)
		}
		phase := 0
		for target, label := range row.Order {
			if r.Permutation[label] != target {
				t.Fatal("permutation", index, r.Permutation, row.Order)
			}
			if r.InputPhases[label] {
				phase |= 1 << uint(target)
			}
		}
		if r.OutputPhase {
			phase |= 1 << uint(row.Variables)
		}
		if phase != row.Phase || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(index, row.Variables, row.Mode, row.High, row.Kind, row.Split, phase, row.Phase, r.Table.Words(), row.Result)
		}
		expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
		if e != nil || !expected.Equal(r.Table) {
			t.Fatal("configuration", index, e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthSCCEnumerationLimits(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	for _, tc := range []struct {
		options TruthSCCEnumerationOptions
		want    error
	}{
		{TruthSCCEnumerationOptions{PhaseMode: 3}, ErrDimension}, {TruthSCCEnumerationOptions{MaxConfigurations: -1}, ErrDimension},
		{TruthSCCEnumerationOptions{MaxConfigurations: 1}, ErrLimit}, {TruthSCCEnumerationOptions{MaxConfigurations: 2}, ErrLimit},
		{TruthSCCEnumerationOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		r, e := source.EnumerateAdaptiveSCC(tc.options)
		if r.Table != nil || !errors.Is(e, tc.want) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthTable)(nil).EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, when := range []int{4, 7} {
		calls := 0
		r, e := source.EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	r, e := source.EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{PhaseMode: TruthSCCPhaseSignature, HighSymmetry: true, SplitByCofactors: true, MaxConfigurations: 3})
	if e != nil || r.Configurations != 3 || r.PhaseConfigurations != 2 || r.FilteredPhases != 1 || r.Permutations != 1 {
		t.Fatal(r, e)
	}
	expected, e := source.TransformNPN(r.Permutation, r.InputPhases, r.OutputPhase)
	if e != nil || !expected.Equal(r.Table) {
		t.Fatal("wide", e)
	}
	r.Table.Set(0, false)
	r.Permutation[0] = 99
	r.InputPhases[0] = true
	fresh, e := source.EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{})
	if e != nil || fresh.Permutation[0] == 99 || fresh.InputPhases[0] || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	for _, value := range []bool{false, true} {
		source, _ := NewConstantTruthTable(0, value, 0)
		r, e := source.EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{})
		if e != nil || !r.Table.IsConstant(false) || r.OutputPhase != value || r.Configurations != 2 || r.Permutations != 1 {
			t.Fatal("zero input", r, e)
		}
	}
}

// Independently build the filtered orbit using Boolean assignments and
// recursive permutations; this does not use mutable permutation state,
// Johnson-Trotter traversal or the implementation's SCC splitting routine.
func checkTruthSCCEnumerationOrbit(t *testing.T, source *TruthTable, options TruthSCCEnumerationOptions) {
	t.Helper()
	reduction, e := source.ReduceInputSymmetry(options.HighSymmetry)
	if e != nil {
		t.Fatal(e)
	}
	filter, e := source.FilterAdaptivePhases(TruthSCCPhaseOptions{Mode: options.PhaseMode, HighSymmetry: options.HighSymmetry})
	if e != nil {
		t.Fatal(e)
	}
	actual, e := source.EnumerateAdaptiveSCC(options)
	if e != nil {
		t.Fatal(e)
	}
	masks := filter.Masks
	if reduction.PhasesDetermined {
		masks = []uint64{0}
	}
	expand := func(labels []int) []int {
		order := []int{}
		for _, head := range labels {
			for v := head; v >= 0; v = reduction.Links[v] {
				order = append(order, v)
			}
		}
		for v := reduction.Links[source.variables]; v >= 0; v = reduction.Links[v] {
			order = append(order, v)
		}
		return order
	}
	value := func(x int, order []int, flips int) bool {
		old := flips
		for target, v := range order {
			if x&(1<<uint(target)) != 0 {
				old ^= 1 << uint(v)
			}
		}
		return reduction.Table.words[0]&(uint64(1)<<uint(old)) != 0
	}
	minimum := ^uint64(0)
	visits := 0
	for _, mask := range masks {
		flips := 0
		for i, head := range filter.PhaseGroups {
			if mask&(uint64(1)<<uint(i)) != 0 {
				for v := head; v >= 0; v = reduction.Links[v] {
					if reduction.PhaseWeights[v] != 0 {
						flips ^= 1 << uint(v)
					}
				}
			}
		}
		labels := slices.Clone(reduction.Representatives)
		groups := slices.Clone(reduction.Groups)
		if options.SplitByCofactors {
			order := expand(labels)
			zero := make([]uint64, source.variables)
			for x := 0; x < 64; x++ {
				if !value(x, order, flips) {
					continue
				}
				weight := uint64(1)
				for bit := 0; bit < 6; bit++ {
					if x&(1<<uint(bit)) != 0 {
						weight *= 2
					}
				}
				for bit := range zero {
					if x&(1<<uint(bit)) == 0 {
						zero[bit] += weight
					}
				}
			}
			weights := make([]uint64, source.variables)
			position := 0
			for _, head := range labels {
				weights[head] = zero[position]
				for v := head; v >= 0; v = reduction.Links[v] {
					position++
				}
			}
			split := []TruthVariableGroup{}
			for _, g := range groups {
				part := labels[g.Start : g.Start+g.Size]
				slices.SortStableFunc(part, func(a, b int) int {
					if weights[a] < weights[b] {
						return -1
					}
					if weights[a] > weights[b] {
						return 1
					}
					return 0
				})
				for begin := g.Start; begin < g.Start+g.Size; {
					end := begin + 1
					for end < g.Start+g.Size && weights[labels[end]] == weights[labels[begin]] {
						end++
					}
					split = append(split, TruthVariableGroup{begin, end - begin})
					begin = end
				}
			}
			groups = split
		}
		var visit func(int)
		visit = func(index int) {
			if index == len(groups) {
				order := expand(labels)
				word := uint64(0)
				for x := 0; x < 64; x++ {
					if value(x, order, flips) {
						word |= uint64(1) << uint(x)
					}
				}
				if word < minimum {
					minimum = word
				}
				visits++
				return
			}
			g := groups[index]
			var permute func(int)
			permute = func(start int) {
				if start == g.Start+g.Size {
					visit(index + 1)
					return
				}
				for j := start; j < g.Start+g.Size; j++ {
					labels[start], labels[j] = labels[j], labels[start]
					permute(start + 1)
					labels[start], labels[j] = labels[j], labels[start]
				}
			}
			permute(g.Start)
		}
		visit(0)
	}
	if actual.Table.words[0] != minimum || actual.Permutations != visits || actual.Configurations != filter.Visited+visits {
		t.Fatal("SCC orbit", options, actual, minimum, visits)
	}
	expected, e := source.TransformNPN(actual.Permutation, actual.InputPhases, actual.OutputPhase)
	if e != nil || !expected.Equal(actual.Table) {
		t.Fatal("forward metadata", e)
	}
}
func FuzzTruthSCCEnumeration(f *testing.F) {
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
		checkTruthSCCEnumerationOrbit(t, source, TruthSCCEnumerationOptions{PhaseMode: TruthSCCPhaseMode((flags / 6) % 3), HighSymmetry: flags&128 != 0, SplitByCofactors: flags&64 != 0})
		if source.words[0] != word {
			t.Fatal("source changed")
		}
	})
}
