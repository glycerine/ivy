package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthHierarchyCacheNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_hierarchy_cache.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Levels, Run, Level, Status int
		Source, Representative                []uint64
		Sizes                                 []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 306 {
		t.Fatal(len(rows))
	}
	var cache *TruthHierarchyCache
	previousVars, previousLevels, previousRun := -1, -1, -1
	for index, row := range rows {
		if row.Variables != previousVars || row.Levels != previousLevels {
			cache, e = NewTruthHierarchyCache(row.Variables, row.Levels, WordMemoryOptions{})
			if e != nil {
				t.Fatal(e)
			}
			previousVars, previousLevels, previousRun = row.Variables, row.Levels, -1
		}
		if row.Run != previousRun {
			if e = cache.Begin(); e != nil {
				t.Fatal(e)
			}
			previousRun = row.Run
		}
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := cache.RetrieveOrInsert(row.Level, source)
		if e != nil || int(r.Status) != row.Status || !slices.Equal(cache.LevelSizes(), row.Sizes) {
			t.Fatal(index, row, r, e)
		}
		if len(row.Representative) == 0 {
			if r.Representative != nil || r.RepresentativeID != -1 {
				t.Fatal("unexpected representative", index, r)
			}
		} else if r.Representative == nil || !slices.Equal(r.Representative.Words(), row.Representative) || r.RepresentativeID < 0 {
			t.Fatal("representative", index, r)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthHierarchyPendingCloneAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(3, 0, 0)
	first, _ := NewVariableTruthTable(3, 1, 0)
	second, _ := NewVariableTruthTable(3, 2, 0)
	cache, e := NewTruthHierarchyCache(3, 3, WordMemoryOptions{MaxEntries: 4, MaxWords: 4})
	if e != nil {
		t.Fatal(e)
	}
	cache.Begin()
	r, e := cache.RetrieveOrInsert(0, source)
	if e != nil || r.Representative != nil {
		t.Fatal(r, e)
	}
	// An abandoned pending entry is reused rather than read as a resolved link.
	cache.Begin()
	r, e = cache.RetrieveOrInsert(0, source)
	if e != nil || r.Status != TruthHierarchyInserted || !slices.Equal(cache.LevelSizes(), []int{1, 0, 0}) {
		t.Fatal(r, e)
	}
	clone := cache.Clone()
	r, e = clone.RetrieveOrInsert(-1, first)
	if e != nil || r.RepresentativeID != 0 {
		t.Fatal(r, e)
	}
	r, e = cache.RetrieveOrInsert(-1, second)
	if e != nil || r.RepresentativeID != 0 {
		t.Fatal(r, e)
	}
	for _, tc := range []struct {
		cache *TruthHierarchyCache
		want  *TruthTable
	}{{cache, second}, {clone, first}} {
		tc.cache.Begin()
		r, e := tc.cache.RetrieveOrInsert(0, source)
		if e != nil || r.Status != TruthHierarchyHit || !r.Representative.Equal(tc.want) {
			t.Fatal(r, e)
		}
		r.Representative.Set(0, true)
		tc.cache.Begin()
		fresh, e := tc.cache.RetrieveOrInsert(0, source)
		if e != nil || !fresh.Representative.Equal(tc.want) {
			t.Fatal("result alias", e)
		}
	}
	if !slices.Equal(cache.LevelSizes(), []int{1, 0, 1}) {
		t.Fatal("unvisited stage was populated")
	}
	cache.Close()
	cache.Close()
	if cache.Clone() != nil || cache.LevelSizes() != nil || !errors.Is(cache.Begin(), ErrDimension) {
		t.Fatal("close")
	}
	if r, e := cache.RetrieveOrInsert(0, nil); e != nil || r.Status != TruthHierarchyUnavailable {
		t.Fatal(r, e)
	}
	clone.Begin()
	if r, e := clone.RetrieveOrInsert(0, source); e != nil || !r.Representative.Equal(first) {
		t.Fatal("clone released with source", e)
	}
}
func TestTruthHierarchyCacheBounds(t *testing.T) {
	for _, tc := range []struct {
		variables, levels int
		options           WordMemoryOptions
		want              error
	}{
		{-1, 1, WordMemoryOptions{}, ErrDimension}, {3, -1, WordMemoryOptions{}, ErrDimension}, {3, 6, WordMemoryOptions{}, ErrDimension},
		{3, 1, WordMemoryOptions{MaxWords: -1}, ErrDimension}, {18, 1, WordMemoryOptions{MaxWords: 1}, ErrLimit},
	} {
		if r, e := NewTruthHierarchyCache(tc.variables, tc.levels, tc.options); r != nil || !errors.Is(e, tc.want) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthHierarchyCache)(nil).RetrieveOrInsert(0, nil); e != nil || r.Status != TruthHierarchyUnavailable {
		t.Fatal(r, e)
	}
	disabled, _ := NewTruthHierarchyCache(3, 0, WordMemoryOptions{})
	if r, e := disabled.RetrieveOrInsert(0, nil); e != nil || r.Status != TruthHierarchyUnavailable {
		t.Fatal(r, e)
	}
	cache, _ := NewTruthHierarchyCache(3, 2, WordMemoryOptions{MaxEntries: 1, MaxWords: 1})
	source, _ := NewVariableTruthTable(3, 0, 0)
	other, _ := NewVariableTruthTable(3, 1, 0)
	if _, e := cache.RetrieveOrInsert(0, source); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	cache.Begin()
	if _, e := cache.RetrieveOrInsert(0, nil); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	wrong, _ := NewConstantTruthTable(4, false, 0)
	if _, e := cache.RetrieveOrInsert(0, wrong); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if r, e := cache.RetrieveOrInsert(-3, nil); e != nil || r.Status != TruthHierarchyUnavailable {
		t.Fatal(r, e)
	}
	cache.RetrieveOrInsert(0, source)
	cache.RetrieveOrInsert(1, other)
	cache.Begin()
	if _, e := cache.RetrieveOrInsert(0, other); !errors.Is(e, ErrLimit) || !slices.Equal(cache.LevelSizes(), []int{1, 1}) {
		t.Fatal("limit atomicity", e)
	}
	cache.Begin()
	if r, e := cache.RetrieveOrInsert(0, source); e != nil || !r.Representative.Equal(other) {
		t.Fatal("limit poisoned cache", e)
	}
	wide, e := NewTruthHierarchyCache(18, 1, WordMemoryOptions{MaxWords: 4096, MaxEntries: 2})
	if e != nil {
		t.Fatal(e)
	}
	a, _ := NewConstantTruthTable(18, false, 0)
	b, _ := NewConstantTruthTable(18, true, 0)
	wide.Begin()
	if _, e := wide.RetrieveOrInsert(0, a); e != nil {
		t.Fatal(e)
	}
	wide.Begin()
	if _, e := wide.RetrieveOrInsert(0, b); !errors.Is(e, ErrLimit) {
		t.Fatal("word bound", e)
	}
}
func FuzzTruthHierarchyCache(f *testing.F) {
	f.Add(uint8(4), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(128), []byte{255, 0})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 160 {
			data = data[:160]
		}
		levels := 1 + int(flags%5)
		cache, e := NewTruthHierarchyCache(6, levels, WordMemoryOptions{MaxEntries: 64, MaxWords: 64})
		if e != nil {
			t.Fatal(e)
		}
		type entry struct {
			resolved       bool
			representative uint64
		}
		model := make([]map[uint64]entry, levels)
		for i := range model {
			model[i] = map[uint64]entry{}
		}
		for run := 0; run < min(24, len(data)); run++ {
			cache.Begin()
			path := make([]uint64, levels)
			for level := 0; level < levels; level++ {
				key := uint64(data[(run+level)%len(data)]) * 0x0101010101010101
				path[level] = key
				item, found := model[level][key]
				if !found {
					model[level][key] = entry{}
				}
				status := TruthHierarchyInserted
				resolved := false
				representative := uint64(0)
				if level == levels-1 {
					representative = key
					resolved = true
				} else if item.resolved {
					status = TruthHierarchyHit
					representative = item.representative
					resolved = true
				}
				if resolved {
					for earlier := 0; earlier < level; earlier++ {
						model[earlier][path[earlier]] = entry{true, representative}
					}
				}
				table, _ := TruthTableFromWords(6, []uint64{key}, 0)
				r, e := cache.RetrieveOrInsert(level-levels, table)
				if e != nil || r.Status != status {
					t.Fatal(run, level, r, e)
				}
				if resolved {
					if r.Representative == nil || r.Representative.words[0] != representative {
						t.Fatal("model representative", run, level, r, representative)
					}
				} else if r.Representative != nil {
					t.Fatal("pending representative")
				}
				sizes := cache.LevelSizes()
				for i := range model {
					if sizes[i] != len(model[i]) {
						t.Fatal("sizes")
					}
				}
				if resolved || flags&64 != 0 && data[run]&1 != 0 {
					break
				}
			}
		}
	})
}
