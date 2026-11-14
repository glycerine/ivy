package core

import (
	"reflect"
	"testing"
)

func TestZddShuffleHeapOrderAndFailureBoundaries(t *testing.T) {
	dd := NewBDDManager(4)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	f := Cudd_zddUnion(dd,
		Cudd_zddChange(dd, dd.One, 0),
		Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 2), Cudd_zddChange(dd, dd.One, 3)))
	if f == nil {
		t.Fatal("failed to construct test family")
	}
	Cudd_Ref(f)
	wantCount, err := dd.CountZDD(f)
	if err != nil {
		t.Fatal(err)
	}
	wantRef := f.Ref
	order := []int{3, 1, 0, 2}
	if Cudd_zddShuffleHeap(dd, order) != 1 {
		t.Fatal("ZDD shuffle failed")
	}
	if !reflect.DeepEqual(dd.InvPermZ, order) {
		t.Fatalf("inverse permutation = %v, want %v", dd.InvPermZ, order)
	}
	for level, index := range order {
		if dd.PermZ[index] != level {
			t.Fatalf("permutation mismatch at index %d", index)
		}
	}
	gotCount, err := dd.CountZDD(f)
	if err != nil || gotCount.Cmp(wantCount) != 0 || f.Ref != wantRef {
		t.Fatal("shuffle changed a live family or its reference count")
	}
	before := append([]int(nil), dd.InvPermZ...)
	if Cudd_zddShuffleHeap(dd, []int{0, 0, 2, 3}) != 0 || Cudd_zddShuffleHeap(dd, []int{0}) != 0 {
		t.Fatal("accepted an invalid ZDD permutation")
	}
	if !reflect.DeepEqual(before, dd.InvPermZ) {
		t.Fatal("invalid permutation modified ZDD order")
	}
	if Cudd_zddShuffleHeap(dd, []int{0, 1, 2, 3, 99}) != 1 {
		t.Fatal("rejected a permutation buffer with a valid required prefix")
	}
}

func TestZddReduceHeapSiftBoundariesAndSemantics(t *testing.T) {
	dd := NewBDDManager(3)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	f := Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 2))
	Cudd_Ref(f)
	want, err := dd.CountZDD(f)
	if err != nil {
		t.Fatal(err)
	}
	order := append([]int(nil), dd.InvPermZ...)
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, len(dd.zddUnique)+1) != 1 || !reflect.DeepEqual(order, dd.InvPermZ) || Cudd_ReadReorderings(dd) != 0 {
		t.Fatal("minsize boundary did not skip reordering")
	}
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_NONE, 0) != 1 || Cudd_ReadReorderings(dd) != 0 {
		t.Fatal("NONE should be a successful no-op")
	}
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, -1) != 1 || Cudd_ReadReorderings(dd) != 0 {
		t.Fatal("negative unsigned minsize should skip reordering")
	}
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, 0) != 1 {
		t.Fatal("ZDD sift failed")
	}
	got, err := dd.CountZDD(f)
	if err != nil || got.Cmp(want) != 0 || Cudd_ReadReorderings(dd) != 1 {
		t.Fatal("sift changed the family or missed its reorder count")
	}
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_GROUP_SIFT, 0) != 0 {
		t.Fatal("unsupported ZDD heuristic should fail")
	}
}

func TestZddReduceHeapLinearModesPreserveFamily(t *testing.T) {
	for _, method := range []Cudd_ReorderingType{CUDD_REORDER_LINEAR, CUDD_REORDER_LINEAR_CONVERGE} {
		dd := NewBDDManager(4)
		if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
			t.Fatal("failed to initialize ZDD variables")
		}
		f := Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 2), Cudd_zddChange(dd, dd.One, 3)))
		Cudd_Ref(f)
		want, err := dd.CountZDD(f)
		if err != nil {
			t.Fatal(err)
		}
		if Cudd_zddReduceHeap(dd, method, 0) != 1 {
			t.Fatalf("ZDD linear reorder failed for method %d", method)
		}
		got, err := dd.CountZDD(f)
		if err != nil || got.Cmp(want) != 0 || !validZddOrder(dd) {
			t.Fatalf("linear reorder changed family for method %d", method)
		}
	}
}

func TestZddRandomReorderingPreservesFamily(t *testing.T) {
	for _, method := range []Cudd_ReorderingType{CUDD_REORDER_RANDOM, CUDD_REORDER_RANDOM_PIVOT} {
		dd := NewBDDManager(4)
		if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
			t.Fatal("failed to initialize ZDD variables")
		}
		f := Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 2), Cudd_zddChange(dd, dd.One, 3)))
		Cudd_Ref(f)
		want, err := dd.CountZDD(f)
		if err != nil {
			t.Fatal(err)
		}
		Cudd_Srandom(1)
		if Cudd_zddReduceHeap(dd, method, 0) != 1 {
			t.Fatalf("random ZDD reorder failed for method %d", method)
		}
		got, err := dd.CountZDD(f)
		if err != nil || got.Cmp(want) != 0 || !validZddOrder(dd) {
			t.Fatalf("random reorder changed family for method %d", method)
		}
	}
}

func TestZddSymmetryCheckUsesBothCofactorEqualitiesAndArcs(t *testing.T) {
	dd := NewBDDManager(2)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	if !zddSymmCheck(dd, 0, 1) {
		t.Fatal("universe variables should be symmetric")
	}
	Cudd_Ref(dd.Univ[1])
	if zddSymmCheck(dd, 0, 1) {
		t.Fatal("external reference omitted from symmetry arc check")
	}
}

func TestZddAlignToBddPreservesMultiplicityBlocks(t *testing.T) {
	dd := NewBDDManager(3)
	if Cudd_zddVarsFromBddVars(dd, 2) != 1 {
		t.Fatal("failed to initialize paired ZDD variables")
	}
	if Cudd_zddShuffleHeap(dd, []int{4, 5, 0, 1, 2, 3}) != 1 || Cudd_ShuffleHeap(dd, []int{2, 0, 1}) != 1 {
		t.Fatal("failed to perturb BDD/ZDD orders")
	}
	if Cudd_zddAlignToBdd(dd) != 1 {
		t.Fatal("failed to align ZDD variables to BDD order")
	}
	for bddLevel, bddIndex := range dd.InvPerm {
		for offset := 0; offset < 2; offset++ {
			zddLevel := bddLevel*2 + offset
			want := bddIndex*2 + offset
			if dd.InvPermZ[zddLevel] != want {
				t.Fatalf("ZDD level %d index=%d, want %d", zddLevel, dd.InvPermZ[zddLevel], want)
			}
		}
	}
}

func TestBddAlignToZddUsesBlockOrder(t *testing.T) {
	dd := NewBDDManager(3)
	if Cudd_zddVarsFromBddVars(dd, 2) != 1 || Cudd_zddShuffleHeap(dd, []int{4, 5, 0, 1, 2, 3}) != 1 {
		t.Fatal("failed to prepare ZDD block order")
	}
	if Cudd_bddAlignToZdd(dd) != 1 || !reflect.DeepEqual(dd.InvPerm, []int{2, 0, 1}) {
		t.Fatalf("BDD order was not aligned to ZDD blocks: %v", dd.InvPerm)
	}
}

func TestZddReorderingHookOrderAndAbort(t *testing.T) {
	dd := NewBDDManager(2)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	var calls []string
	pre := DD_HFP(func(_ *DdManager, graph string, data any) int {
		if graph != "ZDD" || data != CUDD_REORDER_SIFT {
			t.Fatal("unexpected pre-hook arguments", graph, data)
		}
		calls = append(calls, "pre")
		return 1
	})
	post := DD_HFP(func(_ *DdManager, graph string, data any) int {
		if graph != "ZDD" {
			t.Fatal("unexpected post-hook graph", graph)
		}
		calls = append(calls, "post")
		return 1
	})
	if Cudd_AddHook(dd, pre, CUDD_PRE_REORDERING_HOOK) != 1 || Cudd_AddHook(dd, post, CUDD_POST_REORDERING_HOOK) != 1 {
		t.Fatal("hook registration failed")
	}
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, 0) != 1 || !reflect.DeepEqual(calls, []string{"pre", "post"}) {
		t.Fatal("reorder hook order")
	}
	if Cudd_RemoveHook(dd, pre, CUDD_PRE_REORDERING_HOOK) != 1 {
		t.Fatal("hook removal failed")
	}
	calls = nil
	abort := DD_HFP(func(_ *DdManager, _ string, _ any) int { return 0 })
	if Cudd_AddHook(dd, abort, CUDD_PRE_REORDERING_HOOK) != 1 || Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, 0) != 0 || len(calls) != 0 {
		t.Fatal("pre-hook abort semantics")
	}
}
