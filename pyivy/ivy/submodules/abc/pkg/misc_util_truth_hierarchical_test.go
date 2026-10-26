package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthHierarchicalNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_hierarchical.json")
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
	if len(rows) != 1536 {
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
		r, e := source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{})
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
func TestTruthHierarchicalLimitsRetryAndOwnership(t *testing.T) {
	source, _ := TruthTableFromWords(4, []uint64{0x6996699669966996}, 0)
	cache, _ := NewTruthHierarchyCache(4, 5, WordMemoryOptions{MaxEntries: 128, MaxWords: 128})
	if r, e := source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{MaxRefinements: 1}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
	r, e := source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{})
	if e != nil || r.CacheHit {
		t.Fatal("retry pending", r, e)
	}
	expected, e := source.CanonicalizeNPN(TruthCanonicalOptions{})
	if e != nil || !expected.Table.Equal(r.Table) {
		t.Fatal("uncached parity", e)
	}
	r, e = source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{MaxRefinements: 1})
	if e != nil || !r.CacheHit || !r.Table.Equal(expected.Table) {
		t.Fatal("cached retry", r, e)
	}
	r.Table.Set(0, true)
	fresh, e := source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{})
	if e != nil || !fresh.Table.Equal(expected.Table) {
		t.Fatal("cache result alias", e)
	}
	for _, when := range []int{1, 3, 5} {
		abandoned, _ := NewTruthHierarchyCache(4, 5, WordMemoryOptions{MaxEntries: 128, MaxWords: 128})
		calls := 0
		r, e := source.CanonicalizeHierarchical(abandoned, TruthHierarchicalOptions{Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
		r, e = source.CanonicalizeHierarchical(abandoned, TruthHierarchicalOptions{})
		if e != nil || !r.Table.Equal(expected.Table) {
			t.Fatal("retry cancellation", when, r, e)
		}
	}
	if r, e := source.CanonicalizeHierarchical(nil, TruthHierarchicalOptions{MaxRefinements: -1}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	if r, e := (*TruthTable)(nil).CanonicalizeHierarchical(nil, TruthHierarchicalOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	wrong, _ := NewTruthHierarchyCache(3, 5, WordMemoryOptions{})
	if _, e := source.CanonicalizeHierarchical(wrong, TruthHierarchicalOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	cache.Close()
	disabled, e := source.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{})
	if e != nil || disabled.CacheHit || !disabled.Table.Equal(expected.Table) {
		t.Fatal("closed cache", e)
	}
	wide, _ := NewVariableTruthTable(18, 17, 0)
	wideCache, _ := NewTruthHierarchyCache(18, 5, WordMemoryOptions{MaxWords: 4096, MaxEntries: 1})
	a, e := wide.CanonicalizeHierarchical(wideCache, TruthHierarchicalOptions{})
	if e != nil {
		t.Fatal(e)
	}
	b, e := wide.CanonicalizeHierarchical(wideCache, TruthHierarchicalOptions{})
	if e != nil || !b.CacheHit || !b.Table.Equal(a.Table) {
		t.Fatal("wide", e)
	}
	alternate, _ := NewConstantTruthTable(18, false, 0)
	if r, e := alternate.CanonicalizeHierarchical(wideCache, TruthHierarchicalOptions{}); r.Table != nil || !errors.Is(e, ErrLimit) {
		t.Fatal("cache limit", r, e)
	}
	b, e = wide.CanonicalizeHierarchical(wideCache, TruthHierarchicalOptions{})
	if e != nil || !b.Table.Equal(a.Table) {
		t.Fatal("cache limit poisoned entries", e)
	}
}

// Enumerate the unrestricted small-input NPN orbit from Boolean assignments.
// The dictionary is lookup-only; generation/output never depends on map order.
func truthSmallNPNOrbit(variables int, word uint64) map[uint64]bool {
	result := map[uint64]bool{}
	order := make([]int, variables)
	for i := range order {
		order[i] = i
	}
	var permute func(int)
	permute = func(index int) {
		if index < variables {
			for j := index; j < variables; j++ {
				order[index], order[j] = order[j], order[index]
				permute(index + 1)
				order[index], order[j] = order[j], order[index]
			}
			return
		}
		for mask := 0; mask < 1<<uint(variables); mask++ {
			packed := uint64(0)
			for x := 0; x < 64; x++ {
				old := mask
				for target, source := range order {
					if x&(1<<uint(target)) != 0 {
						old ^= 1 << uint(source)
					}
				}
				if word&(uint64(1)<<uint(old)) != 0 {
					packed |= uint64(1) << uint(x)
				}
			}
			result[packed] = true
			result[^packed] = true
		}
	}
	permute(0)
	return result
}
func FuzzTruthHierarchical(f *testing.F) {
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
				r, e := input.CanonicalizeHierarchical(cache, TruthHierarchicalOptions{})
				if e != nil {
					t.Fatal(e)
				}
				if !orbit[r.Table.words[0]] {
					t.Fatal("representative outside NPN orbit", variables, r.Table.Words())
				}
				if !slices.Equal(input.Words(), before) {
					t.Fatal("source changed")
				}
				if e := r.Table.Check(); e != nil {
					t.Fatal(e)
				}
			}
		}
	})
}
