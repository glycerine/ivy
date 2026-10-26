package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthAdaptiveCachedNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_adaptive_cached.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Levels, Round, Mode, Variant, Threshold, Hit, Full int
		Source, Result                                                []uint64
		Sizes                                                         []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 4608 {
		t.Fatal(len(rows))
	}
	var cache *TruthHierarchyCache
	previousVars, previousLevels, previousThreshold := -1, -1, -1
	hits := 0
	for index, row := range rows {
		if row.Variables != previousVars || row.Levels != previousLevels || row.Threshold != previousThreshold {
			cache, e = NewTruthHierarchyCache(row.Variables, row.Levels, WordMemoryOptions{})
			if e != nil {
				t.Fatal(e)
			}
			previousVars, previousLevels, previousThreshold = row.Variables, row.Levels, row.Threshold
		}
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{Threshold: row.Threshold})
		if e != nil || r.CacheHit != (row.Hit != 0) || r.FullEnumeration != (row.Full != 0) || !slices.Equal(r.Table.Words(), row.Result) || !slices.Equal(cache.LevelSizes(), row.Sizes) {
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
		t.Fatal("no cached hits")
	}
}
func TestTruthAdaptiveCachedLimitsRetryAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	cache, _ := NewTruthHierarchyCache(18, 5, WordMemoryOptions{MaxWords: 4096, MaxEntries: 1})
	if r, e := source.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{Threshold: 1085, MaxConfigurations: 39}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
	r, e := source.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{Threshold: 1085, MaxConfigurations: 40})
	if e != nil || r.CacheHit || r.Configurations != 40 || !r.FullEnumeration || !r.CostKnown || r.EnumerationCost != 18 {
		t.Fatal("retry", r, e)
	}
	plain, e := source.CanonicalizeAdaptive(TruthAdaptiveOptions{Threshold: 1085})
	if e != nil || !plain.Table.Equal(r.Table) {
		t.Fatal("uncached result", e)
	}
	r, e = source.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{Threshold: 1085, MaxConfigurations: 1})
	if e != nil || !r.CacheHit || r.HitLevel != 0 || r.CostKnown || r.FullEnumeration || r.Configurations != 0 {
		t.Fatal("early hit", r, e)
	}
	r.Table.Set(0, true)
	again, e := source.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{Threshold: 1085})
	if e != nil || !again.Table.Equal(plain.Table) || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	alternate, _ := NewConstantTruthTable(18, false, 0)
	if r, e := alternate.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal("cache limit", r, e)
	}
	for _, when := range []int{1, 4, 12} {
		abandoned, _ := NewTruthHierarchyCache(18, 5, WordMemoryOptions{MaxWords: 4096, MaxEntries: 1})
		calls := 0
		r, e := source.CanonicalizeAdaptiveCached(abandoned, TruthAdaptiveOptions{Threshold: 1085, Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
		r, e = source.CanonicalizeAdaptiveCached(abandoned, TruthAdaptiveOptions{Threshold: 1085})
		if e != nil || !r.Table.Equal(plain.Table) {
			t.Fatal("retry cancellation", when, r, e)
		}
	}
	if r, e := source.CanonicalizeAdaptiveCached(nil, TruthAdaptiveOptions{Threshold: -1}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	if r, e := source.CanonicalizeAdaptiveCached(nil, TruthAdaptiveOptions{MaxConfigurations: -1}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	if r, e := (*TruthTable)(nil).CanonicalizeAdaptiveCached(nil, TruthAdaptiveOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	wrong, _ := NewTruthHierarchyCache(3, 5, WordMemoryOptions{})
	if _, e := source.CanonicalizeAdaptiveCached(wrong, TruthAdaptiveOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	cache.Close()
	disabled, e := source.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{Threshold: 1085})
	if e != nil || disabled.CacheHit || !disabled.Table.Equal(plain.Table) {
		t.Fatal("closed cache", e)
	}
	// Live one-level caches intentionally stay empty in C's uncached branch.
	single, _ := NewTruthHierarchyCache(18, 1, WordMemoryOptions{MaxWords: 4096, MaxEntries: 1})
	r, e = source.CanonicalizeAdaptiveCached(single, TruthAdaptiveOptions{Threshold: 1085})
	if e != nil || r.CacheHit || !r.Table.Equal(plain.Table) || !slices.Equal(single.LevelSizes(), []int{0}) {
		t.Fatal("one level", r, e)
	}
	for _, variables := range []int{0, 3} {
		for _, value := range []bool{false, true} {
			constant, _ := NewConstantTruthTable(variables, value, 0)
			c, _ := NewTruthHierarchyCache(variables, 5, WordMemoryOptions{})
			r, e := constant.CanonicalizeAdaptiveCached(c, TruthAdaptiveOptions{Threshold: 1085})
			if e != nil || !r.Table.IsConstant(false) {
				t.Fatal("constant extension", variables, r, e)
			}
		}
	}
}
func FuzzTruthAdaptiveCached(f *testing.F) {
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
		orbit := truthSmallNPNOrbit(variables, word)
		thresholds := [4]int{0, 85, 1000, 1085}
		threshold := thresholds[(flags/5)%4]
		cache, e := NewTruthHierarchyCache(variables, int((flags/20)%6), WordMemoryOptions{MaxEntries: 64, MaxWords: 64})
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
				r, e := input.CanonicalizeAdaptiveCached(cache, TruthAdaptiveOptions{Threshold: threshold})
				if e != nil {
					t.Fatal(e)
				}
				if !orbit[r.Table.words[0]] {
					t.Fatal("representative outside NPN orbit", variables, r.Table.Words())
				}
				if !slices.Equal(input.Words(), before) {
					t.Fatal("source changed")
				}
				if r.CostKnown && variables > 0 && r.FullEnumeration != (threshold%100 > 84 || r.EnumerationCost < threshold%100) {
					t.Fatal("path status")
				}
				if e := r.Table.Check(); e != nil {
					t.Fatal(e)
				}
			}
		}
	})
}
