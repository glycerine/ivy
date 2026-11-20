package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestComputedCacheLookupUsesTheInsertingOperation(t *testing.T) {
	dd := Cudd_Init(2, 2, 8, 64, 0)
	defer Cudd_Quit(dd)
	f, g, h := dd.Univ[0], dd.Univ[1], dd.One
	cuddCacheInsert(dd, 0x22, f, g, h, f)
	if got := cuddCacheLookupZdd(dd, 0x22, f, g, h); got != f {
		t.Error("three-operand ZDD lookup missed an inserted result")
	}
	if got := cuddConstantLookup(dd, 0x22, f, g, h); got != f {
		t.Error("constant-only lookup missed the same entry")
	}
	binary := DD_CTFP(core.Cudd_zddUnion)
	cuddCacheInsert2(dd, binary, f, g, f)
	if got := cuddCacheLookup2Zdd(dd, binary, f, g); got != f {
		t.Error("two-operand ZDD lookup missed an inserted result")
	}
	unary := DD_CTFP1(func(_ *DdManager, n *DdNode) *DdNode { return n })
	cuddCacheInsert1(dd, unary, f, f)
	if got := cuddCacheLookup1Zdd(dd, unary, f); got != f {
		t.Error("one-operand ZDD lookup missed an inserted result")
	}
}

func TestComputedCacheInitResizeAndOccupancy(t *testing.T) {
	dd := Cudd_Init(0, 0, 0, 0, 0)
	defer Cudd_Quit(dd)
	if cuddInitCache(dd, 0, 1) != 1 || Cudd_ReadCacheSlots(dd) != 4 {
		t.Errorf("zero-sized request should select four slots, got %d", Cudd_ReadCacheSlots(dd))
	}
	if cuddInitCache(dd, 9, 32) != 1 || Cudd_ReadCacheSlots(dd) != 8 {
		t.Fatalf("nine-slot request should select eight slots, got %d", Cudd_ReadCacheSlots(dd))
	}
	cuddCacheInsert(dd, 0x22, dd.One, dd.One, dd.One, dd.One)
	if used := Cudd_ReadCacheUsedSlots(dd); used != 1.0/8 {
		t.Errorf("one occupied slot gives %g, want 1/8", used)
	}
	cuddCacheResize(dd)
	if slots := Cudd_ReadCacheSlots(dd); slots != 16 || cuddCacheLookup(dd, 0x22, dd.One, dd.One, dd.One) != dd.One {
		t.Fatalf("resize lost the entry or failed to double capacity, slots %d", slots)
	}
	cuddCacheFlush(dd)
	if cuddCacheLookup(dd, 0x22, dd.One, dd.One, dd.One) != nil || Cudd_ReadCacheUsedSlots(dd) != 1.0/16 {
		t.Fatal("flush must invalidate the value and preserve the used-slot statistic")
	}
}

func TestComputedCacheEvictsSlotCollisions(t *testing.T) {
	dd := Cudd_Init(0, 0, 0, 0, 63)
	defer Cudd_Quit(dd)
	var results [65]*DdNode
	for i := uintptr(1); i <= 64; i++ {
		results[i] = core.Cudd_addConst(dd, float64(i))
		core.Cudd_CacheInsert(dd, i, dd.One, dd.One, nil, results[i], 0)
	}
	size := core.Cudd_CacheSize(dd)
	if size < 1 || size > 4 {
		t.Fatalf("64 distinct insertions occupy %d entries in a four-slot cache", size)
	}
	hits := 0
	for i := uintptr(1); i <= 64; i++ {
		if n := core.Cudd_CacheLookup(dd, i, dd.One, dd.One, nil, 0); n != nil {
			hits++
			if n != results[i] {
				t.Fatalf("operation %d returned another operation's result", i)
			}
		}
	}
	if hits != size {
		t.Errorf("found %d valid entries, occupancy says %d", hits, size)
	}
}

func TestComputedCacheFlushAndReorderInvalidateAllOperations(t *testing.T) {
	dd := Cudd_Init(3, 0, 16, 256, 0)
	defer Cudd_Quit(dd)
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	for _, n := range dd.Vars {
		Cudd_Ref(n)
		Cudd_Ref(n)
	}
	core.Cudd_bddAnd(dd, x, y)
	core.Cudd_bddIte(dd, x, y, z)
	cuddCacheFlush(dd)
	hits := Cudd_ReadCacheHits(dd)
	core.Cudd_bddAnd(dd, x, y)
	core.Cudd_bddIte(dd, x, y, z)
	if Cudd_ReadCacheHits(dd) != hits {
		t.Error("public flush left Boolean results in the cache")
	}
	cuddCacheInsert(dd, 0x22, x, y, z, x)
	if core.Cudd_ShuffleHeap(dd, []int{2, 1, 0}) != 1 {
		t.Fatal("shuffle failed")
	}
	if got := cuddCacheLookup(dd, 0x22, x, y, z); got != nil {
		t.Error("shuffle retained a computed entry from the old variable order")
	}
}

func TestComputedCacheGrowsOnHitsAndKeepsLifetimeCounters(t *testing.T) {
	dd := Cudd_Init(0, 0, 8, 4, 0)
	defer Cudd_Quit(dd)
	cuddCacheInsert(dd, 0x22, dd.One, dd.One, dd.One, dd.One)
	for i := 0; i < 10; i++ {
		if cuddCacheLookup(dd, 0x22, dd.One, dd.One, dd.One) != dd.One {
			t.Fatal("missing cached result")
		}
	}
	if cuddCacheLookup(dd, 0x26, dd.One, dd.One, dd.One) != nil {
		t.Fatal("false hit")
	}
	if slots, hits, lookups := Cudd_ReadCacheSlots(dd), Cudd_ReadCacheHits(dd), Cudd_ReadCacheLookUps(dd); slots != 8 || hits != 10 || lookups != 11 {
		t.Fatalf("after triggering resize slots/hits/lookups = %d/%g/%g, want 8/10/11", slots, hits, lookups)
	}
	if cuddCacheLookup(dd, 0x22, dd.One, dd.One, dd.One) != dd.One {
		t.Fatal("resize lost live entry")
	}
}

func TestComputedCacheReleasesEntriesOnQuit(t *testing.T) {
	dd := Cudd_Init(2, 0, 8, 4, 0)
	f, g := dd.Vars[0], dd.Vars[1]
	core.Cudd_CacheInsert(dd, 7, f, g, nil, f, 0)
	Cudd_Quit(dd)
	core.Cudd_CacheInsert(dd, 8, f, g, nil, f, 0)
	if core.Cudd_CacheSize(dd) != 0 || core.Cudd_CacheLookup(dd, 7, f, g, nil, 0) != nil {
		t.Fatal("closed manager retains computed results")
	}
}
