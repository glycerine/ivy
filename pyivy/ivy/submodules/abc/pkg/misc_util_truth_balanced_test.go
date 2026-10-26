package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthBalancedNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_balanced.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Levels, Algorithm, Flag, Round, Mode, Variant, Evaluations, Selected, Hit, Full, Broadened int
		Source, Result                                                                                        []uint64
		Sizes                                                                                                 []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 4612 {
		t.Fatal(len(rows))
	}
	var cache *TruthHierarchyCache
	previousVars, previousLevels, previousAlgorithm, previousFlag := -1, -1, -1, -1
	selected, balanced, hits, hard := 0, 0, 0, 0
	for index, row := range rows {
		if row.Variables != previousVars || row.Levels != previousLevels || row.Algorithm != previousAlgorithm || row.Flag != previousFlag {
			cache, e = NewTruthHierarchyCache(row.Variables, row.Levels, WordMemoryOptions{})
			if e != nil {
				t.Fatal(e)
			}
			previousVars, previousLevels, previousAlgorithm, previousFlag = row.Variables, row.Levels, row.Algorithm, row.Flag
		}
		options := TruthBalancedOptions{Algorithm: TruthCanonicalAlgorithm(row.Algorithm), Hierarchical: TruthHierarchicalOptions{Exact: row.Flag != 0}, Adaptive: TruthAdaptiveOptions{Threshold: row.Flag}, CostAware: TruthCostAwareOptions{BroadenSearch: row.Flag != 0}}
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeBalanced(cache, options)
		if e != nil || r.Evaluations != row.Evaluations || r.ComplementSelected != (row.Selected != 0) || (row.Algorithm != 0 && r.CacheHit != (row.Hit != 0)) || r.FullEnumeration != (row.Full != 0) || r.Broadened != (row.Broadened != 0) || !slices.Equal(r.Table.Words(), row.Result) || !slices.Equal(cache.LevelSizes(), row.Sizes) {
			t.Fatal(index, row, r, e)
		}
		if r.ComplementSelected {
			selected++
		}
		if r.Evaluations == 2 {
			balanced++
		}
		if r.CacheHit {
			hits++
		}
		if r.Broadened {
			hard++
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
	if selected == 0 || balanced == 0 || hits == 0 || hard == 0 {
		t.Fatal("missing native branches", selected, balanced, hits, hard)
	}
}
func TestTruthBalancedLimitsCancellationAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(1, 0, 0)
	options := TruthBalancedOptions{Hierarchical: TruthHierarchicalOptions{Exact: true, MaxCandidates: 4, MaxTransitions: 2}}
	r, e := source.CanonicalizeBalanced(nil, options)
	if e != nil || r.Evaluations != 2 || r.ComplementSelected || r.ExactCandidates != 8 || r.Transitions != 4 {
		t.Fatal("per-polarity bounds or tie", r, e)
	}
	options.Hierarchical.MaxCandidates = 3
	if r, e := source.CanonicalizeBalanced(nil, options); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
	for _, options := range []TruthBalancedOptions{
		{Algorithm: 3}, {Hierarchical: TruthHierarchicalOptions{MaxRefinements: -1}},
		{Algorithm: TruthCanonicalAdaptive, Adaptive: TruthAdaptiveOptions{Threshold: -1}},
		{Algorithm: TruthCanonicalCostAware, CostAware: TruthCostAwareOptions{MaxConfigurations: -1}},
	} {
		if r, e := source.CanonicalizeBalanced(nil, options); r.Table != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthTable)(nil).CanonicalizeBalanced(nil, TruthBalancedOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	input, _ := NewVariableTruthTable(3, 0, 0)
	baselineCache, _ := NewTruthHierarchyCache(3, 5, WordMemoryOptions{})
	baseline := 0
	if _, e := input.CanonicalizeHierarchical(baselineCache, TruthHierarchicalOptions{Stop: func() bool { baseline++; return false }}); e != nil {
		t.Fatal(e)
	}
	cache, _ := NewTruthHierarchyCache(3, 5, WordMemoryOptions{})
	calls := 0
	r, e = input.CanonicalizeBalanced(cache, TruthBalancedOptions{Hierarchical: TruthHierarchicalOptions{Stop: func() bool { calls++; return calls == baseline+2 }}})
	if r.Table != nil || !errors.Is(e, ErrCancelled) {
		t.Fatal("second-run cancellation", baseline, calls, r, e)
	}
	r, e = input.CanonicalizeBalanced(cache, TruthBalancedOptions{})
	if e != nil || r.Evaluations != 2 {
		t.Fatal("retry", r, e)
	}
	expected := r.Table.Words()
	r.Table.Set(0, true)
	fresh, e := input.CanonicalizeBalanced(cache, TruthBalancedOptions{})
	if e != nil || !slices.Equal(fresh.Table.Words(), expected) {
		t.Fatal("result alias", e)
	}
	constant, _ := NewConstantTruthTable(0, true, 0)
	for algorithm := TruthCanonicalHierarchical; algorithm <= TruthCanonicalCostAware; algorithm++ {
		r, e := constant.CanonicalizeBalanced(nil, TruthBalancedOptions{Algorithm: algorithm})
		if e != nil || r.Evaluations != 1 || !r.Table.IsConstant(false) || r.ComplementSelected {
			t.Fatal("unbalanced constant", algorithm, r, e)
		}
	}
	wrong, _ := NewTruthHierarchyCache(4, 5, WordMemoryOptions{})
	if r, e := input.CanonicalizeBalanced(wrong, TruthBalancedOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal("width mismatch", r, e)
	}
}
func FuzzTruthBalancedCanonicalization(f *testing.F) {
	f.Add(uint16(4), uint64(0x96254718abcdef31))
	f.Add(uint16(128), uint64(0))
	f.Fuzz(func(t *testing.T, flags uint16, word uint64) {
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
		options := TruthBalancedOptions{Algorithm: TruthCanonicalAlgorithm((flags / 5) % 3), Hierarchical: TruthHierarchicalOptions{Exact: flags&128 != 0}, Adaptive: TruthAdaptiveOptions{Threshold: 85}, CostAware: TruthCostAwareOptions{BroadenSearch: flags&128 != 0}}
		cache, _ := NewTruthHierarchyCache(variables, int((flags/15)%6), WordMemoryOptions{MaxEntries: 64, MaxWords: 64})
		orbit := truthSmallNPNOrbit(variables, word)
		pop, e := source.CofactorPopulations()
		if e != nil {
			t.Fatal(e)
		}
		for i := 0; i < 3; i++ {
			r, e := source.CanonicalizeBalanced(cache, options)
			if e != nil {
				t.Fatal(e)
			}
			expectedEvaluations := 1
			if pop.Ones == len(source.words)*32 {
				expectedEvaluations = 2
			}
			if r.Evaluations != expectedEvaluations || !orbit[r.Table.words[0]] {
				t.Fatal("balanced class", r)
			}
			if source.words[0] != word {
				t.Fatal("source changed")
			}
		}
		// With no cache, compare wrapper choice to two direct
		// selected-algorithm runs, independently of the wrapper's run dispatch.
		var a, b *TruthTable
		switch options.Algorithm {
		case TruthCanonicalHierarchical:
			x, e := source.CanonicalizeHierarchical(nil, options.Hierarchical)
			if e != nil {
				t.Fatal(e)
			}
			a = x.Table
			y, e := source.Not().CanonicalizeHierarchical(nil, options.Hierarchical)
			if e != nil {
				t.Fatal(e)
			}
			b = y.Table
		case TruthCanonicalAdaptive:
			x, e := source.CanonicalizeAdaptive(options.Adaptive)
			if e != nil {
				t.Fatal(e)
			}
			a = x.Table
			y, e := source.Not().CanonicalizeAdaptive(options.Adaptive)
			if e != nil {
				t.Fatal(e)
			}
			b = y.Table
		case TruthCanonicalCostAware:
			x, e := source.CanonicalizeCostAware(options.CostAware)
			if e != nil {
				t.Fatal(e)
			}
			a = x.Table
			y, e := source.Not().CanonicalizeCostAware(options.CostAware)
			if e != nil {
				t.Fatal(e)
			}
			b = y.Table
		}
		r, e := source.CanonicalizeBalanced(nil, options)
		if e != nil {
			t.Fatal(e)
		}
		want := a
		selected := false
		if pop.Ones == len(source.words)*32 && b.words[0] < a.words[0] {
			want = b
			selected = true
		}
		if !r.Table.Equal(want) || r.ComplementSelected != selected {
			t.Fatal("polarity minimum", r, want.Words())
		}
	})
}
