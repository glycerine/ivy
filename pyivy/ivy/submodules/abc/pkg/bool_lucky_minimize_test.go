package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthExactMinimizationNative(t *testing.T) {
	for _, grouped := range []bool{false, true} {
		name := "truth_minimize"
		expectedRows := 40
		if grouped {
			name = "truth_group_minimize"
			expectedRows = 320
		}
		data, e := os.ReadFile("test_vectors/oracle/" + name + ".json")
		if e != nil {
			t.Fatal(e)
		}
		var rows []struct {
			Variables, Output, Input int
			Groups                   []int
			Source, Result           []uint64
		}
		if e = json.Unmarshal(data, &rows); e != nil {
			t.Fatal(e)
		}
		if len(rows) != expectedRows {
			t.Fatal(name, len(rows))
		}
		for index, row := range rows {
			source, e := TruthTableFromWords(row.Variables, row.Source, 0)
			if e != nil {
				t.Fatal(e)
			}
			var r TruthExactMinimization
			if grouped {
				r, e = source.MinimizeGroups(row.Groups, row.Output != 0, row.Input != 0, TruthExactOptions{})
			} else {
				r, e = source.MinimizeNPN(TruthExactOptions{})
			}
			if e != nil || !slices.Equal(r.Table.Words(), row.Result) {
				t.Fatal(name, index, row, r, e)
			}
			if !slices.Equal(source.Words(), row.Source) {
				t.Fatal("source changed")
			}
		}
	}
}
func TestTruthExactMinimizationLimitsAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(3, 0, 0)
	before := source.Words()
	for _, tc := range []struct {
		options TruthExactOptions
		want    error
	}{
		{TruthExactOptions{MaxCandidates: -1}, ErrDimension}, {TruthExactOptions{MaxScheduleSteps: -1}, ErrDimension},
		{TruthExactOptions{MaxScheduleSteps: 11}, ErrLimit}, {TruthExactOptions{MaxCandidates: 95}, ErrLimit},
		{TruthExactOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		r, e := source.MinimizeNPN(tc.options)
		if r.Table != nil || !errors.Is(e, tc.want) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthTable)(nil).MinimizeNPN(TruthExactOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, when := range []int{3, 20} {
		calls := 0
		r, e := source.MinimizeNPN(TruthExactOptions{Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	r, e := source.MinimizeNPN(TruthExactOptions{MaxCandidates: 96, MaxScheduleSteps: 12})
	if e != nil || r.Candidates != 96 {
		t.Fatal("exact budget", r, e)
	}
	expected := r.Table.Words()
	r.Table.Set(0, true)
	fresh, e := source.MinimizeNPN(TruthExactOptions{})
	if e != nil || !slices.Equal(fresh.Table.Words(), expected) || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	for _, groups := range [][]int{{}, {0, 3}, {-1, 4}, {2}, {4}, {1, 3}} {
		if r, e := source.MinimizeGroups(groups, true, true, TruthExactOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(groups, r, e)
		}
	}
	if r, e := source.MinimizeGroups([]int{2, 1}, true, true, TruthExactOptions{MaxScheduleSteps: 4}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal("shared schedule cap", r, e)
	}
	if r, e := source.MinimizeGroups([]int{2, 1}, true, true, TruthExactOptions{MaxCandidates: 1}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal("group candidate cap", r, e)
	}
	for _, value := range []bool{false, true} {
		zero, _ := NewConstantTruthTable(0, value, 0)
		r, e := zero.MinimizeNPN(TruthExactOptions{})
		if e != nil || !r.Table.IsConstant(false) || r.Candidates != 2 {
			t.Fatal("zero-input exact", r, e)
		}
	}
	one, _ := NewVariableTruthTable(1, 0, 0)
	r, e = one.MinimizeNPN(TruthExactOptions{})
	if e != nil || r.Table.words[0] != 0x5555555555555555 || r.Candidates != 4 {
		t.Fatal("single-input dummy-swap correction", r, e)
	}
}
func truthBruteGroupMinimum(source *TruthTable, groups []int, output, singletons bool) uint64 {
	n := source.variables
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	allow := 0
	start := 0
	for _, size := range groups {
		if size > 1 || singletons {
			for i := start; i < start+size; i++ {
				allow |= 1 << uint(i)
			}
		}
		start += size
	}
	minimum := ^uint64(0)
	var visit func(int, int)
	visit = func(group, start int) {
		if group < len(groups) {
			end := start + groups[group]
			var permute func(int)
			permute = func(i int) {
				if i == end {
					visit(group+1, end)
					return
				}
				for j := i; j < end; j++ {
					order[i], order[j] = order[j], order[i]
					permute(i + 1)
					order[i], order[j] = order[j], order[i]
				}
			}
			permute(start)
			return
		}
		for mask := 0; mask < 1<<uint(n); mask++ {
			if mask&^allow != 0 {
				continue
			}
			word := uint64(0)
			for x := 0; x < 64; x++ {
				old := mask
				for target, v := range order {
					if x&(1<<uint(target)) != 0 {
						old ^= 1 << uint(v)
					}
				}
				if source.words[0]&(uint64(1)<<uint(old)) != 0 {
					word |= uint64(1) << uint(x)
				}
			}
			if word < minimum {
				minimum = word
			}
			if output && ^word < minimum {
				minimum = ^word
			}
		}
	}
	visit(0, 0)
	return minimum
}
func FuzzTruthExactMinimization(f *testing.F) {
	f.Add(uint8(4), uint64(0x96254718abcdef31))
	f.Add(uint8(128), uint64(0))
	f.Fuzz(func(t *testing.T, flags uint8, word uint64) {
		variables := int(flags % 5)
		width := 1 << uint(variables)
		word &= (uint64(1) << uint(width)) - 1
		for shift := width; shift < 64; shift *= 2 {
			word |= word << uint(shift)
		}
		source, e := TruthTableFromWords(variables, []uint64{word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		groups := []int{}
		for left := variables; left > 0; {
			size := left
			switch (flags / 5) % 3 {
			case 1:
				size = 1
			case 2:
				size = min(2, left)
			}
			groups = append(groups, size)
			left -= size
		}
		output, singletons := flags&64 != 0, flags&128 != 0
		grouped, e := source.MinimizeGroups(groups, output, singletons, TruthExactOptions{})
		if e != nil || grouped.Table.words[0] != truthBruteGroupMinimum(source, groups, output, singletons) {
			t.Fatal("group minimum", variables, groups, output, singletons, grouped, e)
		}
		full, e := source.MinimizeNPN(TruthExactOptions{})
		all := []int{}
		if variables > 0 {
			all = []int{variables}
		}
		if e != nil || full.Table.words[0] != truthBruteGroupMinimum(source, all, true, true) {
			t.Fatal("global minimum", variables, full, e)
		}
		if source.words[0] != word {
			t.Fatal("source changed")
		}
	})
}
