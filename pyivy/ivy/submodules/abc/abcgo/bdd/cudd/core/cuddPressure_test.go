package core

import "testing"

func buildPressureBDD() (*DdManager, *DdNode) {
	dd := Cudd_Init(24, 0, 8, 256, 0)
	// Go operations may hold a just-computed unreferenced result locally.
	temporary := dd.node(14, dd.Vars[18], dd.Vars[21])
	for level := range 14 {
		for i := range 4 {
			for j := range 8 {
				root := dd.node(uint32(level), dd.Vars[14+i], Cudd_Not(dd.Vars[14+j]))
				Cudd_Ref(root)
				Cudd_RecursiveDeref(dd, root)
			}
		}
	}
	return dd, temporary
}

func TestPressureCollectsReleasedBDDKeysAndPreservesTemporaries(t *testing.T) {
	dd, temporary := buildPressureBDD()
	defer Cudd_Quit(dd)
	if dd.dead != 448 || dd.minDead != 800 {
		t.Fatalf("setup: %d dead, minimum %d", dd.dead, dd.minDead)
	}
	arg := dd.unique.get(bddUniqueKey{13, dd.Vars[14], Cudd_Not(dd.Vars[15])})
	dead := dd.unique.get(bddUniqueKey{0, dd.Vars[14], Cudd_Not(dd.Vars[16])})
	Cudd_CacheInsert(dd, 100, temporary, nil, nil, temporary, 0)
	Cudd_CacheInsert(dd, 101, dead, nil, nil, dead, 0)
	// The new key consumes a previously released child. The pressure sweep
	// must preserve it along with local results from the current operation.
	root := dd.node(0, arg, temporary)
	if Cudd_ReadGarbageCollections(dd) != 1 || dd.bddTableSlots[0] != 8 || dd.dead != 0 || dd.unique.count != 27 {
		t.Errorf("after pressure: collections %d, width %d, dead %d, resident %d", dd.garbageCollections, dd.bddTableSlots[0], dd.dead, dd.unique.count)
	}
	if root.T != arg || root.E != temporary || arg.manager != dd || temporary.manager != dd ||
		dd.node(13, dd.Vars[14], Cudd_Not(dd.Vars[15])) != arg || dd.node(14, dd.Vars[18], dd.Vars[21]) != temporary ||
		Cudd_CacheLookup(dd, 100, temporary, nil, nil, 0) != temporary || Cudd_CacheLookup(dd, 101, dead, nil, nil, 0) != nil {
		t.Error("pressure collection lost an input, temporary or live cache entry, or retained a dead cache entry")
	}
}

func TestPressureChecksZddAndConstantsBeforeLookup(t *testing.T) {
	t.Run("ZDD", func(t *testing.T) {
		dd := Cudd_Init(0, 9, 8, 64, 0)
		defer Cudd_Quit(dd)
		for i := 1; i <= 4; i++ {
			for j := 1; j <= 8; j++ {
				root := dd.zddNode(0, dd.Univ[i], dd.Univ[j])
				if root == dd.Univ[0] {
					root = dd.zddNode(0, dd.Univ[5], dd.Zero)
				}
				Cudd_Ref(root)
				Cudd_RecursiveDeref(dd, root)
			}
		}
		if got := dd.zddNode(0, dd.Univ[1], dd.Univ[1]); got != dd.Univ[0] ||
			dd.garbageCollections != 1 || dd.zddTableSlots[0] != 8 || dd.zddUnique.count != 9 {
			t.Errorf("ZDD hit: collections %d width %d resident %d", dd.garbageCollections, dd.zddTableSlots[0], dd.zddUnique.count)
		}
	})
	for _, disabled := range []bool{false, true} {
		dd := Cudd_Init(0, 0, 8, 64, 0)
		if disabled {
			Cudd_DisableGarbageCollection(dd)
		}
		for i := 2; i <= 62; i++ {
			root := Cudd_addConst(dd, float64(i))
			Cudd_Ref(root)
			Cudd_RecursiveDeref(dd, root)
		}
		got := Cudd_addConst(dd, 0)
		wantWidth, wantKeys, wantGC := uint32(16), 4, uint64(1)
		if disabled {
			wantWidth, wantKeys, wantGC = 32, 65, 0
		}
		if got != dd.Zero || dd.garbageCollections != wantGC || dd.constantTableWidth() != wantWidth || len(dd.constants) != wantKeys {
			t.Errorf("constant hit (GC disabled %t): collections %d width %d resident %d", disabled, dd.garbageCollections, dd.constantTableWidth(), len(dd.constants))
		}
		Cudd_Quit(dd)
	}
}

func TestSlowGrowthTargetsExistingBDDSubtables(t *testing.T) {
	dd := Cudd_Init(12, 12, 8, 64, 0)
	defer Cudd_Quit(dd)
	Cudd_SlowTableGrowth(dd)
	for i := range 40 {
		j, k := 1+i/8, 1+i%8
		Cudd_Ref(dd.node(0, dd.Vars[j], Cudd_Not(dd.Vars[k])))
		Cudd_Ref(dd.zddNode(0, dd.Univ[j], dd.Univ[k]))
		Cudd_Ref(Cudd_addConst(dd, float64(i+2)))
	}
	if dd.bddTableSlots[0] != 8 || dd.zddTableSlots[0] != 16 || dd.constantTableWidth() != 16 {
		t.Errorf("after slowdown: BDD %d ZDD %d constants %d", dd.bddTableSlots[0], dd.zddTableSlots[0], dd.constantTableWidth())
	}
	Cudd_RehashBDDSubtable(dd, 0)
	for i := 40; i < 70; i++ {
		j, k := 1+i/8, 1+i%8
		Cudd_Ref(dd.node(0, dd.Vars[j], Cudd_Not(dd.Vars[k])))
	}
	if dd.bddTableSlots[0] != 32 {
		t.Errorf("rehashed BDD retained its old density threshold: width %d", dd.bddTableSlots[0])
	}
	Cudd_InsertBDDSubtables(dd, 1, 0)
	for i := range 40 {
		j, k := 1+i/8, 1+i%8
		Cudd_Ref(dd.node(12, dd.Vars[j], Cudd_Not(dd.Vars[k])))
	}
	if dd.bddTableSlots[0] != 16 {
		t.Errorf("new BDD subtable inherited a slowdown: width %d", dd.bddTableSlots[0])
	}
}

func BenchmarkCollectUnderTablePressure(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		dd, temporary := buildPressureBDD()
		arg := dd.node(13, dd.Vars[14], Cudd_Not(dd.Vars[15]))
		root := dd.node(0, arg, temporary)
		if root == nil || dd.garbageCollections != 1 || Cudd_ReadKeys(dd) != 31 {
			b.Fatal("pressure did not recover the expected resident graph")
		}
		Cudd_Quit(dd)
	}
}

func TestPressureCollectionReprobesAfterHook(t *testing.T) {
	dd, temporary := buildPressureBDD()
	defer Cudd_Quit(dd)
	var hooked *DdNode
	Cudd_AddHook(dd, func(*DdManager, string, any) int {
		hooked = dd.node(0, dd.Vars[22], temporary)
		Cudd_Ref(hooked)
		return 1
	}, CUDD_PRE_GC_HOOK)
	got := dd.node(0, dd.Vars[22], temporary)
	if got != hooked || dd.unique.count != 26 || dd.garbageCollections != 1 {
		t.Fatalf("after hook insertion: same=%v keys=%d collections=%d", got == hooked, dd.unique.count, dd.garbageCollections)
	}
}

func TestBoundedOperationPreservesReleasedKeysUnderPressure(t *testing.T) {
	dd, temporary := buildPressureBDD()
	defer Cudd_Quit(dd)
	oldKeys, oldDead, oldSlots := dd.unique.count, dd.dead, dd.slotCount
	run := func() (result *DdNode) {
		defer dd.beginAllocationAttempt(1).finish(&result)
		dd.node(0, dd.Vars[22], temporary)
		return dd.node(0, dd.Vars[23], temporary)
	}
	if run() != nil || dd.unique.count != oldKeys || dd.dead != oldDead || dd.slotCount != oldSlots || dd.garbageCollections != 0 {
		t.Errorf("failed bounded operation changed keys=%d dead=%d slots=%d collections=%d", dd.unique.count, dd.dead, dd.slotCount, dd.garbageCollections)
	}
}

func TestCollectionCompactsSparseProbeArrays(t *testing.T) {
	dd := NewBDDManager(28)
	defer Cudd_Quit(dd)
	Cudd_DisableGarbageCollection(dd)
	keep := make([]*DdNode, 0, 6)
	for i := 1; i < 12; i++ {
		for j := 13; j < 27; j++ {
			for index := range 2 {
				n := dd.node(uint32(index), dd.Vars[i+1], Cudd_Not(dd.Vars[j]))
				Cudd_Ref(n)
				if index == 1 && i == 1 && j <= 18 {
					keep = append(keep, n)
				} else {
					Cudd_RecursiveDeref(dd, n)
				}
			}
		}
	}
	Cudd_GarbageCollect(dd, true)
	if width := len(dd.unique.levels[0].slots); width != 0 {
		t.Errorf("single projection retained %d probe slots", width)
	}
	if width := len(dd.unique.levels[1].slots); width > 4*(len(keep)+1) {
		t.Errorf("seven live nodes retained %d probe slots", width)
	}
	for _, n := range keep {
		if dd.node(n.Index, n.T, n.E) != n {
			t.Fatal("compaction lost a surviving node")
		}
	}
}
