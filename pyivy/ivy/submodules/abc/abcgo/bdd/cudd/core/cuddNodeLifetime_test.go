package core

import (
	"runtime"
	"testing"
)

func TestCollectedHandlesCannotReenterCanonicalTables(t *testing.T) {
	dd := Cudd_Init(3, 2, 8, 256, 0)
	defer Cudd_Quit(dd)
	bdd := Cudd_bddAnd(dd, dd.Vars[1], dd.Vars[2])
	k := Cudd_addConst(dd, 9)
	add := Cudd_addIte(dd, Cudd_addIthVar(dd, 1), k, dd.Zero)
	zdd := dd.zddNode(1, dd.One, dd.Zero)
	for _, n := range []*DdNode{bdd, add, zdd} {
		Cudd_Ref(n)
		Cudd_RecursiveDeref(dd, n)
	}
	if removed := Cudd_GarbageCollect(dd, true); removed < 4 {
		t.Fatal("fixture collected", removed)
	}
	if dd.valid(bdd) || dd.valid(Cudd_Not(bdd)) || dd.validAdd(add, k) || dd.validZdd(zdd) {
		t.Error("a swept handle is still accepted by its manager")
	}
	for _, n := range []*DdNode{Cudd_Regular(bdd), Cudd_Not(bdd), add, k, zdd} {
		if n.manager != nil || n.T != nil || n.E != nil {
			t.Errorf("retired node %d retains manager or children", n.ID)
		}
	}
	before := dd.Vars[2].Ref
	Cudd_Ref(bdd)
	if dd.Vars[2].Ref != before {
		t.Error("pinning a retired handle revived live children")
	}
	fresh := Cudd_bddAnd(dd, dd.Vars[1], dd.Vars[2])
	if fresh == bdd || Cudd_bddAnd(dd, bdd, dd.Vars[0]) != nil {
		t.Error("retired BDD participated in a new canonical operation")
	}
}

func TestRolledBackHandleCannotAliasReusedID(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	var discarded *DdNode
	attempt := func() (result *DdNode) {
		defer dd.beginAllocationAttempt(1).finish(&result)
		discarded = dd.node(0, dd.Vars[1], dd.Vars[2])
		return dd.node(0, dd.Vars[2], dd.Vars[3])
	}
	if attempt() != nil || discarded == nil {
		t.Fatal("expected bounded failure after first key")
	}
	fresh := dd.node(0, dd.Vars[1], dd.Vars[2])
	if fresh == discarded || fresh.ID != discarded.ID || dd.valid(discarded) || dd.valid(Cudd_Not(discarded)) || discarded.T != nil || discarded.E != nil {
		t.Error("aborted handle can alias the new edge or retain its children")
	}
}

func TestRemovedVariableAndReorderedDeadKeyDropChildren(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	old := dd.Vars[3]
	if !Cudd_DestroyBDDSubtables(dd, 1) {
		t.Fatal("variable removal")
	}
	if old.T != nil || old.E != nil {
		t.Error("removed projection retains child pointers")
	}
	dead := Cudd_bddAnd(dd, dd.Vars[1], dd.Vars[2])
	Cudd_Ref(dead)
	Cudd_RecursiveDeref(dd, dead)
	if Cudd_ShuffleHeap(dd, []int{2, 1, 0}) == 0 {
		t.Fatal("reorder")
	}
	if Cudd_Regular(dead).manager != nil || Cudd_Regular(dead).T != nil || Cudd_Regular(dead).E != nil {
		t.Error("retired reorder handle retains children")
	}
}

func TestPressureAllocatesFewerObjectsThanDecisionKeys(t *testing.T) {
	alloc := testing.AllocsPerRun(10, func() {
		dd, tmp := buildPressureBDD()
		dd.node(0, dd.node(13, dd.Vars[14], Cudd_Not(dd.Vars[15])), tmp)
		Cudd_Quit(dd)
	})
	if alloc >= 448 {
		t.Fatalf("448 released decision keys used %.0f allocations", alloc)
	}
}

func TestRollbackOfReleasedNewKeyPreservesExternalPins(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	x, y, z := dd.Vars[1], dd.Vars[2], dd.Vars[3]
	Cudd_Ref(x)
	Cudd_Ref(y)
	xref, yref := x.Ref, y.Ref
	attempt := func() (result *DdNode) {
		defer dd.beginAllocationAttempt(1).finish(&result)
		first := dd.node(0, x, y)
		Cudd_Ref(first)
		Cudd_RecursiveDeref(dd, first)
		return dd.node(0, y, z)
	}
	if attempt() != nil || x.Ref != xref || y.Ref != yref {
		t.Fatalf("rollback of released key changed refs %d/%d to %d/%d", xref, yref, x.Ref, y.Ref)
	}
}

// Run with -benchtime=40x to retain 40 independently collected managers.
func BenchmarkRetainedBDDHeap(b *testing.B) {
	for _, mode := range []string{"Queens", "Strided"} {
		b.Run(mode, func(b *testing.B) {
			var managers [40]*DdManager
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			count := 0
			for b.Loop() {
				var dd *DdManager
				if mode == "Queens" {
					dd, _ = buildQueensBDD(6)
				} else {
					dd = NewBDDManager(65)
					for i := range 4096 {
						n := dd.node(0, dd.Vars[1+i/64], Cudd_Not(dd.Vars[1+i%64]))
						if i%64 == 3 {
							Cudd_Ref(n)
						}
					}
				}
				Cudd_GarbageCollect(dd, true)
				Cudd_InitCache(dd, 4096, 16384)
				managers[count%len(managers)] = dd
				count++
			}
			runtime.GC()
			runtime.ReadMemStats(&after)
			b.ReportMetric(float64(int64(after.HeapAlloc)-int64(before.HeapAlloc))/float64(min(count, len(managers))), "retained-B/manager")
			runtime.KeepAlive(managers)
		})
	}
}
