package core

import "testing"

func TestShuffledZDDUniverseKeepsExactlyOneExternalReference(t *testing.T) {
	dd := Cudd_Init(0, 6, 64, 256, 0)
	defer Cudd_Quit(dd)
	root := dd.Univ[0]
	Cudd_Ref(root)
	for _, order := range [][]int{{4, 3, 1, 5, 2, 0}, {0, 1, 2, 3, 4, 5}} {
		if Cudd_zddShuffleHeap(dd, order) == 0 || dd.Univ[0] != root {
			t.Fatal("universe root moved")
		}
		for _, node := range dd.zddUnique.entries {
			want := uint32(0)
			if node == root {
				want = 1
			}
			if got := Cudd_ExternalRefCount(dd, node); got != want {
				t.Errorf("external refs for node %d = %d, want %d", node.ID, got, want)
			}
			if node.inverse.Index != node.Index {
				t.Errorf("inverse of %d has index %d, want %d", node.ID, node.inverse.Index, node.Index)
			}
		}
		Cudd_GarbageCollect(dd, true)
		if dd.zddUnique.count != 6 {
			t.Fatalf("pure universe retained %d decision keys after shuffle", dd.zddUnique.count)
		}
	}
	Cudd_RecursiveDeref(dd, root)
	FreeZDDUniverse(dd)
	Cudd_GarbageCollect(dd, true)
	if dd.zddUnique.count != 0 {
		t.Fatal("former universe kept keys", dd.zddUnique.count)
	}
}

func TestZddSwapPreservesExternalPinsAndReleasedChildren(t *testing.T) {
	dd := Cudd_Init(0, 4, 64, 256, 0)
	defer Cudd_Quit(dd)
	low := dd.zddNode(3, dd.One, dd.Zero)
	y := dd.zddNode(1, low, dd.One)
	f := dd.zddNode(0, y, dd.Zero)
	garbage := dd.zddNode(0, low, y)
	Cudd_Ref(f)
	Cudd_Ref(y)
	Cudd_Ref(garbage)
	Cudd_RecursiveDeref(dd, garbage)
	if !garbage.childrenReleased || !swapZddAdjacentInPlace(dd, 0, 1) {
		t.Fatal("fixture")
	}
	if Cudd_ExternalRefCount(dd, f) != 1 || Cudd_ExternalRefCount(dd, y) != 1 || !garbage.childrenReleased {
		t.Fatal("owned and released nodes changed their pins")
	}
	// The family f contains {0,1,3} and {0}, and y contains {1,3} and {}.
	contains := func(n *DdNode, mask int) bool {
		for !Cudd_IsConstant(n) {
			bit := 1 << n.Index
			if mask&bit != 0 {
				mask ^= bit
				n = n.T
			} else {
				n = n.E
			}
		}
		return n == dd.One && mask == 0
	}
	Cudd_GarbageCollect(dd, true)
	for mask := range 16 {
		if contains(f, mask) != (mask == 11 || mask == 1) || contains(y, mask) != (mask == 10 || mask == 0) {
			t.Fatalf("set %d", mask)
		}
	}
	Cudd_RecursiveDeref(dd, f)
	Cudd_RecursiveDeref(dd, y)
	Cudd_GarbageCollect(dd, true)
	if dd.zddUnique.count != 4 {
		t.Fatalf("released all external pins, retained %d universe keys", dd.zddUnique.count)
	}
}

func BenchmarkShuffleZDDTripleSets(b *testing.B) {
	const width = 10
	order := make([]int, width)
	for i := range order {
		order[i] = width - 1 - i
	}
	b.ReportAllocs()
	for b.Loop() {
		dd := Cudd_Init(0, width, 64, 256, 0)
		Cudd_DisableGarbageCollection(dd)
		for i := range width {
			for j := i + 1; j < width; j++ {
				for k := j + 1; k < width; k++ {
					leaf := dd.zddNode(uint32(k), dd.One, dd.Zero)
					middle := dd.zddNode(uint32(j), leaf, dd.Zero)
					Cudd_Ref(dd.zddNode(uint32(i), middle, dd.Zero))
				}
			}
		}
		if Cudd_zddShuffleHeap(dd, order) == 0 || dd.InvPermZ[0] != 9 || dd.InvPermZ[9] != 0 {
			b.Fatal("shuffle")
		}
		Cudd_Quit(dd)
	}
}

func TestZDDReorderPreprocessesReleasedKeysAndCache(t *testing.T) {
	for _, mode := range []string{"identity shuffle", "heap reduction"} {
		t.Run(mode, func(t *testing.T) {
			dd := Cudd_Init(2, 3, 32, 64, 0)
			defer Cudd_Quit(dd)
			z := dd.zddNode(1, dd.One, dd.Zero)
			b := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
			for _, n := range []*DdNode{z, b} {
				Cudd_Ref(n)
				Cudd_RecursiveDeref(dd, n)
			}
			tmp := dd.zddNode(0, dd.One, dd.Zero)
			Cudd_CacheInsert(dd, 5, dd.Vars[0], dd.One, nil, dd.Vars[0], 0)
			pre, post := 0, 0
			Cudd_AddHook(dd, func(*DdManager, string, any) int { pre++; return 1 }, CUDD_PRE_GC_HOOK)
			Cudd_AddHook(dd, func(*DdManager, string, any) int { post++; return 1 }, CUDD_POST_GC_HOOK)
			ok := 0
			if mode == "identity shuffle" {
				ok = Cudd_zddShuffleHeap(dd, []int{0, 1, 2})
			} else {
				ok = Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, 0)
			}
			if ok == 0 || pre == 0 || pre != post || Cudd_CacheSize(dd) != 0 || z.manager != nil || b.manager != nil || tmp.manager != dd {
				t.Fatalf("returned %d hooks %d/%d cache %d retired %t/%t retained temporary %t", ok, pre, post, Cudd_CacheSize(dd), z.manager == nil, b.manager == nil, tmp.manager == dd)
			}
		})
	}
}

func TestZDDReduceHeapThresholdCountsOnlyLiveKeys(t *testing.T) {
	dd := Cudd_Init(0, 3, 32, 64, 0)
	defer Cudd_Quit(dd)
	z := dd.zddNode(1, dd.One, dd.Zero)
	Cudd_Ref(z)
	Cudd_RecursiveDeref(dd, z)
	keys := dd.zddUnique.count
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, keys) != 1 || dd.reorderings != 0 || z.manager != dd {
		t.Fatalf("released key triggered reorder at threshold %d: reorders %d", keys, dd.reorderings)
	}
	Cudd_Ref(z)
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, keys) != 1 || dd.reorderings != 1 || z.manager != dd {
		t.Fatal("revived key should reach threshold")
	}
}

func TestAdjacentZDDShuffleScoresOnlyLiveUniverseKeys(t *testing.T) {
	dd := Cudd_Init(0, 6, 32, 64, 0)
	defer Cudd_Quit(dd)
	if Cudd_zddShuffleHeap(dd, []int{4, 3, 1, 5, 2, 0}) == 0 {
		t.Fatal("shuffle failed")
	}
	if size := zddHeapSize(dd); size != 6 || dd.liveZDDKeys() != 6 {
		t.Fatalf("shuffle scored %d keys, %d live; native universe has six", size, dd.liveZDDKeys())
	}
}

func TestAdjacentZDDSwapMovesLogicalSubtableWidths(t *testing.T) {
	dd := Cudd_Init(0, 3, 8, 32, 0)
	defer Cudd_Quit(dd)
	Cudd_RehashZddSubtable(dd, 0)
	Cudd_RehashZddSubtable(dd, 0)
	Cudd_RehashZddSubtable(dd, 1)
	total := Cudd_ReadSlots(dd)
	if !swapZddAdjacentInPlace(dd, 0, 1) || dd.zddTableSlots[0] != 16 || dd.zddTableSlots[1] != 32 || Cudd_ReadSlots(dd) != total {
		t.Fatalf("swapped subtable capacities %v, total %d", dd.zddTableSlots, Cudd_ReadSlots(dd))
	}
}

func TestZDDLiveCountAcrossReclaimSwapAndCollection(t *testing.T) {
	dd := Cudd_Init(0, 8, 32, 64, 0)
	defer Cudd_Quit(dd)
	check := func() {
		t.Helper()
		live := 0
		for _, n := range dd.zddUnique.entries {
			if !n.childrenReleased {
				live++
			}
		}
		if got := Cudd_zddReadNodeCount(dd); got != live+2 || zddHeapSize(dd) != live {
			t.Fatalf("node count %d, heap score %d, walked %d decisions", got, zddHeapSize(dd), live)
		}
	}
	var nodes []*DdNode
	for i := range 7 {
		for j := i + 1; j < 8; j++ {
			n := dd.zddNode(uint32(i), dd.zddNode(uint32(j), dd.One, dd.Zero), dd.Zero)
			Cudd_Ref(n)
			nodes = append(nodes, n)
		}
	}
	check()
	for i, n := range nodes {
		if i%2 == 0 {
			Cudd_RecursiveDeref(dd, n)
		}
	}
	check()
	for i, n := range nodes {
		if i%4 == 0 {
			Cudd_Ref(n)
		}
	}
	check()
	if Cudd_zddShuffleHeap(dd, []int{4, 3, 7, 2, 5, 1, 0, 6}) != 1 {
		t.Fatal("shuffle")
	}
	check()
	Cudd_GarbageCollect(dd, true)
	check()
	for i, n := range nodes {
		if i%2 != 0 || i%4 == 0 {
			Cudd_RecursiveDeref(dd, n)
		}
	}
	check()
	FreeZDDUniverse(dd)
	check()
	Cudd_GarbageCollect(dd, true)
	check()
}

func BenchmarkZDDReadMixedLiveCount(b *testing.B) {
	dd := Cudd_Init(0, 32, 64, 256, 0)
	defer Cudd_Quit(dd)
	Cudd_DisableGarbageCollection(dd)
	for i := range 31 {
		for j := i + 1; j < 32; j++ {
			n := dd.zddNode(uint32(i), dd.zddNode(uint32(j), dd.One, dd.Zero), dd.Zero)
			Cudd_Ref(n)
			if j%2 == 0 {
				Cudd_RecursiveDeref(dd, n)
			}
		}
	}
	want := Cudd_zddReadNodeCount(dd)
	if want != 306 {
		b.Fatal("fixture", want)
	}
	b.ReportAllocs()
	for b.Loop() {
		if Cudd_zddReadNodeCount(dd) != want {
			b.Fatal("count")
		}
	}
}
