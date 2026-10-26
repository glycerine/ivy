package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthHierarchicalExactNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_hierarchical_exact.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Levels, Round, Mode, Variant int
		Source, Result                          []uint64
		Sizes                                   []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 1920 {
		t.Fatal(len(rows))
	}
	var cache *TruthHierarchyCache
	previousVars, previousLevels := -1, -1
	hits := 0
	for index, row := range rows {
		if row.Variables != previousVars || row.Levels != previousLevels {
			cache, e = NewTruthHierarchyCache(row.Variables, row.Levels, WordMemoryOptions{})
			if e != nil {
				t.Fatal(e)
			}
			previousVars, previousLevels = row.Variables, row.Levels
		}
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{Exact: true})
		if e != nil || !slices.Equal(r.Table.Words(), row.Result) || !slices.Equal(cache.LevelSizes(), row.Sizes) {
			t.Fatal(index, row, r, e)
		}
		if r.CacheHit {
			hits++
			if r.HitLevel < 0 || r.HitLevel >= row.Levels-1 {
				t.Fatal("hit level", r)
			}
		} else if r.HitLevel != -1 {
			t.Fatal("nonhit level", r)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
	if hits == 0 {
		t.Fatal("no native cache hits exercised")
	}
}

func TestTruthHierarchicalExactLimitsAndRetry(t *testing.T) {
	source, _ := TruthTableFromWords(4, []uint64{0x6996699669966996}, 0)
	before := source.Words()
	for _, options := range []TruthHierarchicalOptions{
		{Exact: true, MaxScheduleSteps: 1}, {Exact: true, MaxCandidates: 1}, {Exact: true, MaxTransitions: 1},
	} {
		cache, _ := NewTruthHierarchyCache(4, 5, WordMemoryOptions{MaxEntries: 64, MaxWords: 64})
		r, e := source.CanonicalizeHierarchical(cache, options)
		if r.Table != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(options, r, e)
		}
		r, e = source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{Exact: true})
		if e != nil || r.CacheHit || !r.ExactSearch || r.ExactCandidates == 0 || r.ExactTransitions == 0 {
			t.Fatal("pending exact retry", r, e)
		}
		saved := r.Table.Words()
		r.Table.Set(0, true)
		hit, e := source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{Exact: true, MaxCandidates: 1})
		if e != nil || !hit.CacheHit || hit.ExactSearch || !slices.Equal(hit.Table.Words(), saved) {
			t.Fatal("exact hit/ownership", hit, e)
		}
	}
	for _, options := range []TruthHierarchicalOptions{{Exact: true, MaxCandidates: -1}, {Exact: true, MaxScheduleSteps: -1}, {Exact: true, MaxTransitions: -1}} {
		if _, e := source.CanonicalizeHierarchical(nil, options); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, when := range []int{10, 40} {
		calls := 0
		r, e := source.CanonicalizeHierarchical(nil, TruthHierarchicalOptions{Exact: true, Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	if !slices.Equal(source.Words(), before) {
		t.Fatal("source changed")
	}
}
func TestTruthGroupedNoOpTransitionBudget(t *testing.T) {
	source, _ := NewConstantTruthTable(8, false, 0)
	groups := []int{1, 1, 1, 1, 1, 1, 1, 1}
	if r, e := source.MinimizeGroups(groups, false, false, TruthExactOptions{MaxCandidates: 1, MaxTransitions: 255}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal("unbounded no-op carry", r, e)
	}
	r, e := source.MinimizeGroups(groups, false, false, TruthExactOptions{MaxCandidates: 1, MaxTransitions: 256})
	if e != nil || r.Candidates != 1 || r.Transitions != 256 || !r.Table.IsConstant(false) {
		t.Fatal("transition boundary", r, e)
	}
	calls := 0
	r, e = source.MinimizeGroups(groups, false, false, TruthExactOptions{Stop: func() bool { calls++; return calls == 5 }})
	if r.Table != nil || !errors.Is(e, ErrCancelled) {
		t.Fatal("no-op cancellation", r, e)
	}
}
func checkTruthHierarchicalExactModel(t *testing.T, source *TruthTable) {
	t.Helper()
	actual, e := source.CanonicalizeHierarchical(nil, TruthHierarchicalOptions{Exact: true})
	if e != nil {
		t.Fatal(e)
	}
	if source.variables == 0 {
		if !actual.Table.IsConstant(false) {
			t.Fatal("zero input")
		}
		return
	}
	heuristic, e := source.CanonicalizeNPN(TruthCanonicalOptions{})
	if e != nil {
		t.Fatal(e)
	}
	groups := []int{}
	counts := heuristic.ZeroCofactorCounts
	for start := 0; start < source.variables; {
		end := start + 1
		for end < source.variables && counts[end] == counts[start] {
			end++
		}
		groups = append(groups, end-start)
		start = end
	}
	output := heuristic.PackedOnes == len(source.words)*32
	second := heuristic.PackedOnes
	if source.variables > 1 {
		second = counts[1]
	}
	singleton := output || counts[0] != second && counts[0] == heuristic.PackedOnes-counts[0]
	want := truthBruteGroupMinimum(heuristic.Table, groups, output, singleton)
	if actual.Table.words[0] != want || !actual.ExactSearch {
		t.Fatal("exact hierarchical group minimum", source.variables, actual.Table.Words(), want)
	}
}
func FuzzTruthHierarchicalExact(f *testing.F) {
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
		checkTruthHierarchicalExactModel(t, source)
		orbit := truthSmallNPNOrbit(variables, word)
		cache, e := NewTruthHierarchyCache(variables, int((flags/5)%6), WordMemoryOptions{MaxEntries: 64, MaxWords: 64})
		if e != nil {
			t.Fatal(e)
		}
		for run := 0; run < 2; run++ {
			for variant := 0; variant < 4; variant++ {
				input := source.Clone()
				switch variant {
				case 1:
					input = input.Not()
				case 2:
					if variables > 0 {
						input, _ = input.FlipVariable(0)
					}
				case 3:
					if variables > 1 {
						input, _ = input.SwapVariables(0, variables-1)
					}
				}
				before := input.Words()
				r, e := input.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{Exact: true})
				if e != nil {
					t.Fatal(e)
				}
				if !orbit[r.Table.words[0]] {
					t.Fatal("exact representative outside orbit")
				}
				if !slices.Equal(input.Words(), before) {
					t.Fatal("source changed")
				}
			}
		}
	})
}
