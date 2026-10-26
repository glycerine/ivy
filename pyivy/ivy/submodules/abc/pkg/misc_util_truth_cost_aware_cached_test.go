package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthCostAwareCachedNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_cost_aware_cached.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Levels, Round, Mode, Variant, Enabled, Hit, Full, Broadened int
		Source, Result                                                         []uint64
		Sizes                                                                  []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 2312 {
		t.Fatal(len(rows))
	}
	var cache *TruthHierarchyCache
	previousVars, previousLevels, previousEnabled := -1, -1, -1
	hits, broadened := 0, 0
	for index, row := range rows {
		if row.Variables != previousVars || row.Levels != previousLevels || row.Enabled != previousEnabled {
			cache, e = NewTruthHierarchyCache(row.Variables, row.Levels, WordMemoryOptions{})
			if e != nil {
				t.Fatal(e)
			}
			previousVars, previousLevels, previousEnabled = row.Variables, row.Levels, row.Enabled
		}
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CanonicalizeCostAwareCached(cache, TruthCostAwareOptions{BroadenSearch: row.Enabled != 0})
		if e != nil || r.CacheHit != (row.Hit != 0) || r.FullEnumeration != (row.Full != 0) || r.Broadened != (row.Broadened != 0) || !slices.Equal(r.Table.Words(), row.Result) || !slices.Equal(cache.LevelSizes(), row.Sizes) {
			t.Fatal(index, row, r, e)
		}
		if r.CacheHit {
			hits++
			if r.HitLevel < 0 || r.HitLevel >= row.Levels-1 || r.EstimatesKnown {
				t.Fatal("hit metadata", r)
			}
		} else if r.HitLevel != -1 {
			t.Fatal("nonhit level", r)
		}
		if r.Broadened {
			broadened++
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
	if hits == 0 || broadened == 0 {
		t.Fatal("missing native branches", hits, broadened)
	}
}
func TestTruthCostAwareCachedLimitsRetry(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	options := TruthCostAwareOptions{BroadenSearch: true}
	cache, _ := NewTruthHierarchyCache(18, 5, WordMemoryOptions{MaxWords: 4096, MaxEntries: 1})
	if r, e := source.CanonicalizeCostAwareCached(cache, TruthCostAwareOptions{MaxConfigurations: 37}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
	r, e := source.CanonicalizeCostAwareCached(cache, TruthCostAwareOptions{BroadenSearch: true, MaxConfigurations: 38})
	if e != nil || r.CacheHit || r.Configurations != 38 || !r.FullEnumeration || !r.EstimatesKnown {
		t.Fatal("retry", r, e)
	}
	plain, e := source.CanonicalizeCostAware(options)
	if e != nil || !r.Table.Equal(plain.Table) || r.EnumerationEstimate != plain.EnumerationEstimate || r.PhaseEstimate != plain.PhaseEstimate {
		t.Fatal("uncached parity", e)
	}
	r, e = source.CanonicalizeCostAwareCached(cache, TruthCostAwareOptions{MaxConfigurations: 1})
	if e != nil || !r.CacheHit || r.HitLevel != 0 || r.Configurations != 0 || r.EstimatesKnown || r.FullEnumeration || r.Broadened {
		t.Fatal("early hit", r, e)
	}
	r.Table.Set(0, true)
	again, e := source.CanonicalizeCostAwareCached(cache, options)
	if e != nil || !again.Table.Equal(plain.Table) || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	alternate, _ := NewConstantTruthTable(18, false, 0)
	if r, e := alternate.CanonicalizeCostAwareCached(cache, options); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal("cache limit", r, e)
	}
	for _, when := range []int{1, 4, 12} {
		c, _ := NewTruthHierarchyCache(18, 5, WordMemoryOptions{MaxWords: 4096, MaxEntries: 1})
		calls := 0
		r, e := source.CanonicalizeCostAwareCached(c, TruthCostAwareOptions{Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
		r, e = source.CanonicalizeCostAwareCached(c, options)
		if e != nil || !r.Table.Equal(plain.Table) {
			t.Fatal("retry cancellation", when, e)
		}
	}
	if r, e := source.CanonicalizeCostAwareCached(nil, TruthCostAwareOptions{MaxConfigurations: -1}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	if r, e := (*TruthTable)(nil).CanonicalizeCostAwareCached(nil, options); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	wrong, _ := NewTruthHierarchyCache(3, 5, WordMemoryOptions{})
	if _, e := source.CanonicalizeCostAwareCached(wrong, options); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	cache.Close()
	disabled, e := source.CanonicalizeCostAwareCached(cache, options)
	if e != nil || disabled.CacheHit || !disabled.Table.Equal(plain.Table) {
		t.Fatal("closed cache", e)
	}
	single, _ := NewTruthHierarchyCache(18, 1, WordMemoryOptions{MaxWords: 4096, MaxEntries: 1})
	for i := 0; i < 2; i++ {
		r, e := source.CanonicalizeCostAwareCached(single, options)
		if e != nil || r.CacheHit || !r.Table.Equal(plain.Table) || !slices.Equal(single.LevelSizes(), []int{1}) {
			t.Fatal("one level final storage", r, e)
		}
	}
	for _, variables := range []int{0, 3} {
		for _, value := range []bool{false, true} {
			constant, _ := NewConstantTruthTable(variables, value, 0)
			c, _ := NewTruthHierarchyCache(variables, 5, WordMemoryOptions{})
			r, e := constant.CanonicalizeCostAwareCached(c, options)
			if e != nil || !r.Table.IsConstant(false) {
				t.Fatal("constant extension", variables, r, e)
			}
		}
	}
}
func FuzzTruthCostAwareCached(f *testing.F) {
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
				r, e := input.CanonicalizeCostAwareCached(cache, TruthCostAwareOptions{BroadenSearch: flags&128 != 0})
				if e != nil {
					t.Fatal(e)
				}
				if !orbit[r.Table.words[0]] {
					t.Fatal("representative outside NPN orbit", variables, r.Table.Words())
				}
				if !slices.Equal(input.Words(), before) {
					t.Fatal("source changed")
				}
				if r.EstimatesKnown && r.Broadened != (flags&128 != 0 && r.EnumerationEstimate > r.PhaseEstimate) {
					t.Fatal("cost decision")
				}
				if e := r.Table.Check(); e != nil {
					t.Fatal(e)
				}
			}
		}
	})
}
