package core

import "testing"

func TestLimitedAndAbstractRestoresPartialAttempt(t *testing.T) {
	dd := NewBDDManager(5)
	defer Cudd_Quit(dd)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])
	g := Cudd_bddXor(dd, dd.Vars[1], dd.Vars[3])
	before, id := dd.unique.count, dd.nextID
	refs := make([]uint32, len(dd.Vars))
	for i, v := range dd.Vars {
		refs[i] = v.Ref
	}
	for attempt := range 2 {
		if got := Cudd_bddAndAbstractLimit(dd, f, g, dd.Vars[4], 1); got != nil || dd.unique.count != before || dd.nextID != id {
			t.Fatalf("attempt %d, got %p, keys %d/%d, next ID %d/%d", attempt, got, dd.unique.count, before, dd.nextID, id)
		}
		for i, v := range dd.Vars {
			if v.Ref != refs[i] {
				t.Fatal("changed reference count", i)
			}
		}
	}
	if Cudd_bddAndAbstractLimit(dd, f, g, dd.Vars[4], 20) != Cudd_bddAnd(dd, f, g) {
		t.Fatal("retry")
	}
}

func TestLimitedAndAbstractHonorsOuterCeiling(t *testing.T) {
	dd := NewBDDManager(2)
	defer Cudd_Quit(dd)
	dd.allocationLimited, dd.allocationCeiling = true, dd.unique.count
	if got := Cudd_bddAndAbstractLimit(dd, dd.Vars[0], dd.Vars[1], dd.One, 10); got != nil {
		t.Fatal("inner operation exceeded its parent's allocation ceiling")
	}
}

func TestBDDAbstractionReportsInvalidCube(t *testing.T) {
	dd := NewBDDManager(2)
	defer Cudd_Quit(dd)
	for name, abstract := range map[string]func(*DdManager, *DdNode, *DdNode) *DdNode{
		"exist": Cudd_bddExistAbstract, "univ": Cudd_bddUnivAbstract, "xor-exist": func(dd *DdManager, f, cube *DdNode) *DdNode { return Cudd_bddXorExistAbstract(dd, f, dd.One, cube) },
	} {
		Cudd_ClearErrorCode(dd)
		if got := abstract(dd, dd.One, Cudd_Not(dd.Vars[0])); got != nil || Cudd_ReadErrorCode(dd) != CUDD_INVALID_ARG {
			t.Errorf("%s got %p, error %d", name, got, Cudd_ReadErrorCode(dd))
		}
	}
}

func TestFusedAbstractionReusesCommutedResult(t *testing.T) {
	dd := NewBDDManager(5)
	defer Cudd_Quit(dd)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[3])
	g := Cudd_bddXor(dd, dd.Vars[1], dd.Vars[2])
	for range 2 {
		Cudd_Ref(f)
		Cudd_Ref(g)
	}
	cube := dd.Vars[2]
	first := Cudd_bddAndAbstract(dd, f, g, cube)
	hits, lookups := dd.cacheHits, dd.cacheLookups
	if Cudd_bddAndAbstract(dd, g, f, cube) != first || dd.cacheHits != hits+1 || dd.cacheLookups != lookups+1 {
		t.Fatalf("commuted query used %d lookups with %d hits", dd.cacheLookups-lookups, dd.cacheHits-hits)
	}
}

func TestBDDAbstractionTruthTables(t *testing.T) {
	dd := NewBDDManager(3)
	defer Cudd_Quit(dd)
	if Cudd_ShuffleHeap(dd, []int{2, 0, 1}) == 0 {
		t.Fatal("reorder")
	}
	for table := range 256 {
		f := bddFromTruth(dd, uint8(table), 3)
		for mask := range 8 {
			cube := dd.One
			for i := range 3 {
				if mask&(1<<i) != 0 {
					cube = Cudd_bddAnd(dd, cube, dd.Vars[i])
				}
			}
			exist, univ := Cudd_bddExistAbstract(dd, f, cube), Cudd_bddUnivAbstract(dd, f, cube)
			for input := range 8 {
				any, all := false, true
				for candidate := range 8 {
					if (candidate^input)&^mask != 0 {
						continue
					}
					set := table&(1<<candidate) != 0
					any, all = any || set, all && set
				}
				if bddTruth(t, dd, exist, input) != any || bddTruth(t, dd, univ, input) != all {
					t.Fatalf("table %02x mask %03b input %03b", table, mask, input)
				}
			}
		}
	}
}

func BenchmarkCommutedAndAbstract(b *testing.B) {
	dd := Cudd_Init(5, 0, 256, 4096, 0)
	defer Cudd_Quit(dd)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[3])
	g := Cudd_bddXor(dd, dd.Vars[1], dd.Vars[2])
	for range 2 {
		Cudd_Ref(f)
		Cudd_Ref(g)
	}
	cube := dd.Vars[2]
	if Cudd_bddAndAbstract(dd, f, g, cube) != f {
		b.Fatal("warmup")
	}
	b.ReportAllocs()
	for b.Loop() {
		if Cudd_bddAndAbstract(dd, g, f, cube) != f {
			b.Fatal("image")
		}
	}
}
