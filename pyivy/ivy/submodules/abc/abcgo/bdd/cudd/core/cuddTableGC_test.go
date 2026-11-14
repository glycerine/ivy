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
	if dd.unique[rootKey] == nil || dd.unique[leafKey] == nil {
		t.Fatal("referenced graph was not retained")
	}
	Cudd_Deref(root)
	if deleted := Cudd_GarbageCollect(dd, true); deleted != 2 {
		t.Fatalf("GC removed %d nodes, want the unreachable root and child", deleted)
	}
	if dd.unique[rootKey] != nil || dd.unique[leafKey] != nil {
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
	if len(dd.computed) != 1 {
		t.Fatalf("dead-cache filtering retained %d entries, want one live entry", len(dd.computed))
	}
	if dd.unique[bddUniqueKey{root.Index, root.T, root.E}] != nil || dd.unique[bddUniqueKey{leaf.Index, leaf.T, leaf.E}] != nil {
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
	if dd.tableGrowthFactor != 4 || dd.maxCacheHard != dd.cacheSlots-1 || dd.gcFraction != DD_GC_FRAC_MIN ||
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
