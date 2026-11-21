package core

import (
	"math"
	"testing"
)

func TestCuddTableGCReleasesUnreachableParentsAndPreservesRootedClosure(t *testing.T) {
	dd := NewBDDManager(3)
	leaf := Cudd_BddNode(dd, 1, dd.Vars[2], dd.LogicZero)
	root := Cudd_BddNode(dd, 0, leaf, dd.LogicZero)
	if leaf == nil || root == nil {
		t.Fatal("failed to build test graph")
	}
	rootKey := bddUniqueKey{root.Index, root.T, root.E}
	leafKey := bddUniqueKey{leaf.Index, leaf.T, leaf.E}
	Cudd_Ref(root)
	if deleted := Cudd_GarbageCollect(dd, true); deleted != 0 {
		t.Fatalf("GC removed %d nodes reachable from the referenced root", deleted)
	}
	if Cudd_ReadGarbageCollections(dd) != 0 {
		t.Fatal("GC count advanced when there was no garbage")
	}
	if dd.unique.get(rootKey) == nil || dd.unique.get(leafKey) == nil {
		t.Fatal("referenced graph was not retained")
	}
	Cudd_Deref(root)
	if deleted := Cudd_GarbageCollect(dd, true); deleted != 2 {
		t.Fatalf("GC removed %d nodes, want the unreachable root and child", deleted)
	}
	if dd.unique.get(rootKey) != nil || dd.unique.get(leafKey) != nil {
		t.Fatal("unreachable nodes remain in the unique table")
	}
	if got := Cudd_ReadGarbageCollections(dd); got != 1 {
		t.Fatalf("GC count = %d, want 1", got)
	}
	if got := dd.Vars[2].Ref; got != 0 {
		t.Fatalf("unreferenced projection refcount = %d, want 0", got)
	}
}

func TestCuddTableGCRecursiveDerefAndDeadCacheFiltering(t *testing.T) {
	dd := NewBDDManager(3)
	leaf := Cudd_BddNode(dd, 1, dd.Vars[2], dd.LogicZero)
	root := Cudd_BddNode(dd, 0, leaf, dd.LogicZero)
	dead := Cudd_BddNode(dd, 0, dd.Vars[1], dd.LogicZero)
	if leaf == nil || root == nil || dead == nil {
		t.Fatal("failed to build test graph")
	}
	Cudd_Ref(root)
	Cudd_CacheInsert(dd, 1, root, dd.Vars[1], nil, root, 0)
	Cudd_CacheInsert(dd, 2, dead, dd.Vars[1], nil, root, 0)
	Cudd_CacheInsert(dd, 3, dd.Vars[0], dd.Vars[1], nil, dd.Vars[0], 0)
	Cudd_RecursiveDeref(dd, root)
	if root.Ref != 0 || !root.childrenReleased || leaf.Ref != 0 || !leaf.childrenReleased {
		t.Fatalf("recursive dereference did not release child references: root=%+v leaf=%+v", root, leaf)
	}
	if deleted := Cudd_GarbageCollect(dd, true); deleted != 3 {
		t.Fatalf("GC removed %d nodes, want 3", deleted)
	}
	if Cudd_CacheSize(dd) != 1 {
		t.Fatalf("dead-cache filtering retained %d entries, want one live entry", Cudd_CacheSize(dd))
	}
	if dd.unique.get(bddUniqueKey{root.Index, root.T, root.E}) != nil || dd.unique.get(bddUniqueKey{leaf.Index, leaf.T, leaf.E}) != nil {
		t.Fatal("recursive dereference graph remains in unique table")
	}
}

func TestCuddTableGCPreHookCancellationDoesNotCountCollection(t *testing.T) {
	dd := NewBDDManager(2)
	dead := Cudd_BddNode(dd, 0, dd.Vars[1], dd.One)
	if dead == nil {
		t.Fatal("failed to create dead node")
	}
	if Cudd_AddHook(dd, func(*DdManager, string, any) int { return 0 }, CUDD_PRE_GC_HOOK) != 1 {
		t.Fatal("failed to register pre-GC hook")
	}
	if got := Cudd_GarbageCollect(dd, true); got != 0 || Cudd_ReadGarbageCollections(dd) != 0 {
		t.Fatalf("cancelled GC returned %d with count %d", got, Cudd_ReadGarbageCollections(dd))
	}
}

func TestCuddUniqueConstantUsesManagerEpsilon(t *testing.T) {
	dd := NewBDDManager(0)
	if Cudd_addConst(dd, 0.5e-12) != dd.Zero {
		t.Fatal("constant below epsilon was not canonicalized to zero")
	}
	base := Cudd_addConst(dd, 7)
	if got := Cudd_addConst(dd, 7+0.5e-12); got != base {
		t.Fatal("constants within epsilon were not interned together")
	}
	if got := Cudd_addConst(dd, 7+2e-12); got == base {
		t.Fatal("constants outside epsilon were incorrectly merged")
	}
	plusInfinity := Cudd_addConst(dd, math.Inf(1))
	if Cudd_addConst(dd, math.Inf(1)) != plusInfinity {
		t.Fatal("equal infinities were not interned together")
	}
	if Cudd_ReadEpsilon(dd) != 1e-12 {
		t.Fatalf("default epsilon = %g, want 1e-12", Cudd_ReadEpsilon(dd))
	}
	Cudd_SetEpsilon(dd, 0)
	if Cudd_ReadEpsilon(dd) != 0 {
		t.Fatal("epsilon setter did not update the manager")
	}
	if Cudd_addConst(dd, 0.5e-12) == dd.Zero {
		t.Fatal("zero epsilon still collapsed a nonzero constant to zero")
	}
}

func TestCuddReorderingCycleManagerField(t *testing.T) {
	dd := NewBDDManager(0)
	if Cudd_ReadReorderingCycle(dd) != 0 {
		t.Fatal("reordering cycle default is not zero")
	}
	Cudd_SetReorderingCycle(dd, 17)
	if Cudd_ReadReorderingCycle(dd) != 17 || dd.ReorderCycle != 17 {
		t.Fatal("reordering cycle setter did not update manager state")
	}
}

func TestCuddTableManagerParameterDefaultsAndAccessors(t *testing.T) {
	dd := NewBDDManager(0)
	if Cudd_ReadGroupcheck(dd) != CUDD_GROUP_CHECK7 || Cudd_GarbageCollectionEnabled(dd) != 1 || Cudd_DeadAreCounted(dd) != 0 {
		t.Fatal("CUDD table-manager defaults differ from cuddInitTable")
	}
	Cudd_SetGroupcheck(dd, CUDD_GROUP_CHECK5)
	Cudd_DisableGarbageCollection(dd)
	Cudd_TurnOnCountDead(dd)
	Cudd_SetRecomb(dd, 3)
	Cudd_SetSymmviolation(dd, 4)
	Cudd_SetArcviolation(dd, 5)
	Cudd_SetPopulationSize(dd, 6)
	Cudd_SetNumberXovers(dd, 7)
	if Cudd_ReadGroupcheck(dd) != CUDD_GROUP_CHECK5 || Cudd_GarbageCollectionEnabled(dd) != 0 || Cudd_DeadAreCounted(dd) != 1 ||
		Cudd_ReadRecomb(dd) != 3 || Cudd_ReadSymmviolation(dd) != 4 || Cudd_ReadArcviolation(dd) != 5 || Cudd_ReadPopulationSize(dd) != 6 || Cudd_ReadNumberXovers(dd) != 7 {
		t.Fatal("table-manager accessors failed to preserve state")
	}
	Cudd_EnableGarbageCollection(dd)
	Cudd_TurnOffCountDead(dd)
	if Cudd_GarbageCollectionEnabled(dd) != 1 || Cudd_DeadAreCounted(dd) != 0 {
		t.Fatal("GC/count-dead toggles failed to restore defaults")
	}
}

func TestCuddTableCapacityGrowthAndLimits(t *testing.T) {
	dd := Cudd_InitTable(3, 2, 9, 128)
	if dd == nil || Cudd_ReadSlots(dd) != 96 {
		t.Fatalf("initial slots = %d, want 96", Cudd_ReadSlots(dd))
	}
	if !Cudd_RehashBDDSubtable(dd, 0) || Cudd_ReadSlots(dd) != 112 {
		t.Fatalf("BDD rehash slots = %d, want 112", Cudd_ReadSlots(dd))
	}
	if !Cudd_ShrinkBDDSubtable(dd, 0) || Cudd_ReadSlots(dd) != 96 {
		t.Fatalf("BDD shrink slots = %d, want 96", Cudd_ReadSlots(dd))
	}
	Cudd_FixTableLimits(dd)
	if dd.minDead != uint32(DD_GC_FRAC_HI*float64(dd.slotCount)) {
		t.Fatal("table limit repair did not apply the current GC fraction")
	}
	Cudd_SlowTableGrowth(dd)
	if dd.bddDensityFactor(0) != 4 || dd.maxCacheHard != dd.cacheSlots-1 || dd.gcFraction != DD_GC_FRAC_MIN ||
		dd.minDead != uint32(DD_GC_FRAC_MIN*float64(dd.slotCount)) || dd.cacheSlack != -int32(dd.cacheSlots+1) {
		t.Fatal("slow-table-growth thresholds do not match cuddSlowTableGrowth")
	}
}

func TestCuddTableSubtableInsertAndDestroy(t *testing.T) {
	dd := NewBDDManager(2)
	if !Cudd_InsertBDDSubtables(dd, 1, 1) {
		t.Fatal("failed to insert a BDD subtable")
	}
	if Cudd_ReadSize(dd) != 3 || dd.InvPerm[0] != 0 || dd.InvPerm[1] != 2 || dd.InvPerm[2] != 1 || Cudd_ReadSlots(dd) != 1024 {
		t.Fatalf("insert state: size=%d order=%v slots=%d", Cudd_ReadSize(dd), dd.InvPerm, Cudd_ReadSlots(dd))
	}
	Cudd_Ref(dd.Vars[2])
	if Cudd_DestroyBDDSubtables(dd, 1) {
		t.Fatal("destroyed a projection variable with an external reference")
	}
	Cudd_Deref(dd.Vars[2])
	if !Cudd_DestroyBDDSubtables(dd, 1) {
		t.Fatal("failed to destroy the most recently inserted subtable")
	}
	if Cudd_ReadSize(dd) != 2 || dd.InvPerm[0] != 0 || dd.InvPerm[1] != 1 || Cudd_ReadSlots(dd) != 768 {
		t.Fatalf("destroy state: size=%d order=%v slots=%d", Cudd_ReadSize(dd), dd.InvPerm, Cudd_ReadSlots(dd))
	}
}

func TestCuddResizeTablesCreatesProjectionAndUniverseState(t *testing.T) {
	bdd := NewBDDManager(0)
	if !Cudd_ResizeBDDTable(bdd, 2) || Cudd_ReadSize(bdd) != 3 || Cudd_ReadSlots(bdd) != 1024 {
		t.Fatalf("BDD resize state: size=%d slots=%d", Cudd_ReadSize(bdd), Cudd_ReadSlots(bdd))
	}
	zdd := NewBDDManager(0)
	if !Cudd_ResizeZddTable(zdd, 3) || zdd.ZddSize != 4 || len(zdd.Univ) != 4 || Cudd_ReadSlots(zdd) != 1280 {
		t.Fatalf("ZDD resize state: size=%d universe=%d slots=%d", zdd.ZddSize, len(zdd.Univ), Cudd_ReadSlots(zdd))
	}
}

func TestCuddLiveCountsAcrossReleaseAndReclaim(t *testing.T) {
	dd := Cudd_Init(3, 2, 16, 64, 0)
	defer Cudd_Quit(dd)
	check := func(label string, keys, dead uint32, live, zlive, peak int) {
		t.Helper()
		got := [...]int{int(Cudd_ReadKeys(dd)), int(Cudd_ReadDead(dd)), int(Cudd_ReadMinDead(dd)), Cudd_ReadNodeCount(dd), Cudd_zddReadNodeCount(dd), Cudd_ReadPeakLiveNodeCount(dd)}
		want := [...]int{int(keys), int(dead), 384, live, zlive, peak}
		if got != want {
			t.Errorf("%s [keys dead threshold BDD/ADD ZDD peak] = %v, want %v", label, got, want)
		}
	}
	check("initial", 7, 0, 1, 4, 7)
	root := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	Cudd_Ref(root)
	check("live conjunction", 8, 0, 3, 4, 8)
	Cudd_RecursiveDeref(dd, root)
	check("released conjunction", 8, 1, 1, 4, 8)
	if got := Cudd_bddAnd(dd, dd.Vars[1], dd.Vars[0]); got != root {
		t.Fatal("reclaimed conjunction changed identity")
	}
	Cudd_Ref(root)
	check("reclaimed conjunction", 8, 0, 3, 4, 8)
	Cudd_RecursiveDeref(dd, root)
	Cudd_GarbageCollect(dd, true)
	check("after collection", 7, 0, 1, 4, 8)
	z := Cudd_zddChange(dd, dd.One, 0)
	Cudd_Ref(z)
	check("ZDD singleton uses ADD zero", 7, 0, 2, 5, 8)
	Cudd_RecursiveDeref(dd, z)
	check("released ZDD singleton", 7, 0, 1, 4, 8)
	if got := Cudd_zddChange(dd, dd.One, 0); got != z {
		t.Fatal("reclaimed ZDD singleton changed identity")
	}
	Cudd_Ref(z)
	check("reclaimed ZDD singleton", 7, 0, 2, 5, 8)
	Cudd_RecursiveDeref(dd, z)
	Cudd_GarbageCollect(dd, true)
	check("after ZDD collection", 7, 0, 1, 4, 8)
}

func TestRecursiveDerefPreservesManagerPins(t *testing.T) {
	dd := Cudd_Init(3, 2, 16, 64, 0)
	defer Cudd_Quit(dd)
	for _, pin := range []*DdNode{dd.Vars[1], dd.Zero, dd.PlusInfinity, dd.MinusInfinity, dd.Univ[0]} {
		childRefs := dd.One.Ref
		Cudd_Ref(pin)
		Cudd_RecursiveDeref(dd, pin)
		if pin.childrenReleased || dd.One.Ref != childRefs {
			t.Errorf("releasing pin %d left released=%v one refs=%d, want false and %d", pin.ID, pin.childrenReleased, dd.One.Ref, childRefs)
		}
	}
}

func TestExactConstantLookupReclaimsDeadEntry(t *testing.T) {
	dd := NewBDDManager(0)
	defer Cudd_Quit(dd)
	n := Cudd_addConst(dd, 12.5)
	Cudd_Ref(n)
	Cudd_RecursiveDeref(dd, n)
	if !n.childrenReleased {
		t.Fatal("released constant stayed live")
	}
	if got := Cudd_addConst(dd, 12.5); got != n || n.childrenReleased || n.Ref != 0 {
		t.Errorf("exact constant lookup got %p, want %p with released=false and ref=0, got %v and %d", got, n, n.childrenReleased, n.Ref)
	}
}

func TestPeakLiveCountSurvivesUnobservedRelease(t *testing.T) {
	dd := NewBDDManager(2)
	defer Cudd_Quit(dd)
	root := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	Cudd_Ref(root)
	Cudd_RecursiveDeref(dd, root)
	Cudd_GarbageCollect(dd, true)
	if got := Cudd_ReadPeakLiveNodeCount(dd); got != 7 {
		t.Errorf("peak after releasing the seventh key = %d, want 7", got)
	}
}

func TestDestroySubtablesChecksADDAndReleasesDeadParents(t *testing.T) {
	t.Run("live ADD", func(t *testing.T) {
		dd := NewBDDManager(4)
		defer Cudd_Quit(dd)
		live := Cudd_addIthVar(dd, 2)
		Cudd_Ref(live)
		slots, keys := Cudd_ReadSlots(dd), Cudd_ReadKeys(dd)
		last := dd.Vars[3]
		if Cudd_DestroyBDDSubtables(dd, 2) {
			t.Fatal("removed the subtable of a referenced ADD")
		}
		if len(dd.Vars) != 4 || dd.Vars[3] != last || Cudd_ReadSlots(dd) != slots || Cudd_ReadKeys(dd) != keys {
			t.Fatal("rejecting removal changed the manager")
		}
		Cudd_RecursiveDeref(dd, live)
		if !Cudd_DestroyBDDSubtables(dd, 2) || dd.addUnique.count != 0 {
			t.Fatal("released ADD prevented removal")
		}
	})
	t.Run("dead parent", func(t *testing.T) {
		dd := NewBDDManager(5)
		defer Cudd_Quit(dd)
		root := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[4])
		Cudd_Ref(root)
		if Cudd_DestroyBDDSubtables(dd, 1) {
			t.Fatal("removed a projection with a live parent")
		}
		Cudd_RecursiveDeref(dd, root)
		if !Cudd_DestroyBDDSubtables(dd, 1) || len(dd.Vars) != 4 || Cudd_ReadDead(dd) != 0 {
			t.Fatal("a released parent blocked removal of its former child")
		}
		if dd.unique.get(bddUniqueKey{root.Index, root.T, root.E}) != nil {
			t.Fatal("a dangling canonical key survived removal")
		}
	})
}

func TestDestroyedVariablesLoseReorderingAttributes(t *testing.T) {
	dd := NewBDDManager(5)
	defer Cudd_Quit(dd)
	for _, index := range []int{1, 4} {
		Cudd_bddBindVar(dd, index)
		Cudd_bddSetVariableType(dd, index, CUDD_VAR_PRESENT_STATE)
		Cudd_bddSetPairIndex(dd, index, 0)
		Cudd_bddSetLazyGroup(dd, index, CUDD_LAZY_HARD_GROUP)
	}
	if !Cudd_DestroyBDDSubtables(dd, 1) || Cudd_bddIthVar(dd, 4) == nil {
		t.Fatal("failed to replace the last variable")
	}
	if Cudd_bddVarIsBound(dd, 4) != 0 || Cudd_bddVariableType(dd, 4) != int(CUDD_VAR_PRIMARY_INPUT) ||
		Cudd_bddReadPairIndex(dd, 4) != -1 || Cudd_bddLazyGroup(dd, 4) != int(CUDD_LAZY_NONE) {
		t.Error("replacement variable inherited the retired variable's attributes")
	}
	if Cudd_bddVarIsBound(dd, 1) != 1 || Cudd_bddVariableType(dd, 1) != int(CUDD_VAR_PRESENT_STATE) ||
		Cudd_bddReadPairIndex(dd, 1) != 0 || Cudd_bddLazyGroup(dd, 1) != int(CUDD_LAZY_HARD_GROUP) {
		t.Error("surviving variable lost its attributes")
	}
}

func BenchmarkInsertDestroySubtables(b *testing.B) {
	const keep, add, level = 256, 128, 64
	dd := Cudd_Init(keep, 0, 8, 4096, 0)
	defer Cudd_Quit(dd)
	baseSlots := Cudd_ReadSlots(dd)
	b.ReportAllocs()
	for b.Loop() {
		if !Cudd_InsertBDDSubtables(dd, add, level) || dd.InvPerm[level] != keep ||
			!Cudd_DestroyBDDSubtables(dd, add) || len(dd.Vars) != keep || dd.InvPerm[level] != level || Cudd_ReadSlots(dd) != baseSlots {
			b.Fatal("insert/destroy cycle")
		}
	}
}

func TestDestroySubtableCollectsRetiredProjectionsOnce(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	checks := 0
	Cudd_AddHook(dd, func(*DdManager, string, any) int { checks++; return 1 }, CUDD_POST_GC_HOOK)
	base := Cudd_ReadGarbageCollections(dd)
	if !Cudd_DestroyBDDSubtables(dd, 2) || Cudd_ReadGarbageCollections(dd) != base+1 || checks != 1 {
		t.Errorf("removed projections: collections %d, hooks %d", Cudd_ReadGarbageCollections(dd)-base, checks)
	}
	veto := func(*DdManager, string, any) int { return 0 }
	Cudd_AddHook(dd, veto, CUDD_PRE_GC_HOOK)
	project := dd.Vars[1]
	if Cudd_DestroyBDDSubtables(dd, 1) || len(dd.Vars) != 2 || dd.Vars[1] != project || project.manager != dd || checks != 1 {
		t.Error("collection veto left a partially destroyed subtable")
	}
}

func TestDestroySaturatedIsolatedProjection(t *testing.T) {
	dd := NewBDDManager(3)
	defer Cudd_Quit(dd)
	project := dd.Vars[2]
	project.Ref = math.MaxUint32
	root := Cudd_bddAnd(dd, dd.Vars[0], project)
	Cudd_Ref(root)
	if Cudd_DestroyBDDSubtables(dd, 1) {
		t.Fatal("accepted saturated projection with a live parent")
	}
	Cudd_RecursiveDeref(dd, root)
	if !Cudd_DestroyBDDSubtables(dd, 1) {
		t.Fatal("rejected a saturated isolated projection")
	}
}

func TestProjectionBatchPreservesExistingCanonicalPairs(t *testing.T) {
	dd := Cudd_InitTable(260, 0, 8, 4096)
	defer Cudd_Quit(dd)
	e := dd.node(128, dd.One, dd.LogicZero)
	if Cudd_bddIthVar(dd, 257) == nil || dd.Vars[128] != e || dd.unique.count != 258 {
		t.Fatal("materializing projections replaced or duplicated a canonical pair")
	}
	for i, v := range dd.Vars {
		if int(v.Index) != i || v.manager != dd || v.T != dd.One || v.E != dd.LogicZero ||
			Cudd_Regular(Cudd_Not(v)) != v || dd.unique.get(bddUniqueKey{uint32(i), dd.One, dd.LogicZero}) != v {
			t.Fatalf("wrong canonical pair at %d", i)
		}
	}
	old := dd.Vars[129]
	if !Cudd_DestroyBDDSubtables(dd, 129) || dd.Vars[128] != e || old.manager != nil ||
		Cudd_bddIthVar(dd, 129) == old || Cudd_Not(e).manager != dd {
		t.Fatal("removing one part of a batch affected surviving handles")
	}
	if len(dd.bddTableSlots) != 131 || Cudd_ReadSlots(dd) != 132*8 {
		t.Fatalf("reserved subtable capacity after shrink and re-extension: %d slots", Cudd_ReadSlots(dd))
	}
}
