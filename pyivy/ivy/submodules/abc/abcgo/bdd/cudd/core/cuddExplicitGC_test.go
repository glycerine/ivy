package core

import (
	"math"
	"math/rand"
	"testing"
)

func TestGCLeavesNonConstantCacheResultForLiveOperands(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	x, y := dd.Vars[0], dd.Vars[1]
	dead := Cudd_bddAnd(dd, x, y)
	Cudd_CacheInsert(dd, 101, x, y, nil, DD_NON_CONSTANT, 2)
	Cudd_CacheInsert(dd, 102, dead, y, nil, DD_NON_CONSTANT, 2)
	if Cudd_GarbageCollect(dd, true) != 1 {
		t.Fatal("expected one swept AND")
	}
	if result := Cudd_CacheLookup(dd, 101, x, y, nil, 2); result != DD_NON_CONSTANT {
		t.Fatal("collection discarded a live nonconstant query")
	}
	if result := Cudd_CacheLookup(dd, 102, dead, y, nil, 2); result != nil {
		t.Fatal("cache retained dead operand")
	}
}

func TestExplicitGCNeedsNoHeapScratchForLiveKeys(t *testing.T) {
	dd := NewBDDManager(33)
	defer Cudd_Quit(dd)
	for i := range 32 {
		for j := range 32 {
			Cudd_Ref(dd.node(0, dd.Vars[1+i], Cudd_Not(dd.Vars[1+j])))
		}
	}
	if n := testing.AllocsPerRun(10, func() {
		if removed := Cudd_GarbageCollect(dd, true); removed != 0 {
			t.Fatal("removed live keys", removed)
		}
	}); n != 0 {
		t.Fatalf("collection allocated %.0f scratch objects for 1024 live keys", n)
	}
}

func BenchmarkExplicitGCMixedKeys(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		dd := NewBDDManager(65)
		Cudd_DisableGarbageCollection(dd)
		for i := range 64 {
			for j := range 64 {
				key := dd.node(0, dd.Vars[i+1], Cudd_Not(dd.Vars[j+1]))
				Cudd_Ref(key)
				if (i+j)%4 != 0 {
					Cudd_RecursiveDeref(dd, key)
				}
			}
		}
		b.StartTimer()
		if n := Cudd_GarbageCollect(dd, true); n != 3072 {
			b.Fatalf("collected %d keys", n)
		}
		b.StopTimer()
		Cudd_Quit(dd)
		b.StartTimer()
	}
}

func TestExplicitGCMatchesReachabilityAcrossShuffledFamilies(t *testing.T) {
	for seed := int64(0); seed < 12; seed++ {
		dd := Cudd_Init(6, 6, 64, 256, 0)
		Cudd_DisableGarbageCollection(dd)
		// Establish the order before creating the universe and fixture keys.
		FreeZDDUniverse(dd)
		Cudd_GarbageCollect(dd, true)
		rng := rand.New(rand.NewSource(seed))
		if Cudd_ShuffleHeap(dd, rng.Perm(6)) == 0 || Cudd_zddShuffleHeap(dd, rng.Perm(6)) == 0 {
			t.Fatal("shuffle")
		}
		if InitializeZDDUniverse(dd) != 1 {
			t.Fatal("create shuffled universe")
		}
		bdds := []*DdNode{dd.One, dd.LogicZero}
		adds := []*DdNode{dd.One, dd.Zero, Cudd_addConst(dd, -3), Cudd_addConst(dd, 5)}
		zdds := []*DdNode{dd.One, dd.Zero}
		choose := func(pool []*DdNode) *DdNode { return pool[rng.Intn(len(pool))] }
		for level := 5; level >= 0; level-- {
			// Pick children only from deeper levels. The shuffled numeric
			// indices intentionally do not give a topological sweep order.
			b, a, z := bdds, adds, zdds
			for range 24 {
				bdds = append(bdds, dd.node(uint32(dd.InvPerm[level]), choose(b), Cudd_Not(choose(b))))
				adds = append(adds, dd.addNode(uint32(dd.InvPerm[level]), choose(a), choose(a)))
				zdds = append(zdds, dd.zddNode(uint32(dd.InvPermZ[level]), choose(z), choose(z)))
			}
		}
		roots := []*DdNode{dd.One, dd.Zero, dd.PlusInfinity, dd.MinusInfinity}
		roots = append(roots, dd.Vars...)
		roots = append(roots, dd.Univ...)
		for _, pool := range [][]*DdNode{bdds, adds, zdds} {
			for range 6 {
				n := choose(pool)
				Cudd_Ref(n)
				roots = append(roots, n)
			}
			n := choose(pool)
			Cudd_Ref(n)
			Cudd_RecursiveDeref(dd, n)
		}
		background := adds[len(adds)-1]
		Cudd_SetBackground(dd, background)
		saturated := Cudd_Regular(bdds[len(bdds)-1])
		saturated.Ref = math.MaxUint32
		roots = append(roots, background, saturated)
		want := map[*DdNode]bool{}
		for len(roots) > 0 {
			n := Cudd_Regular(roots[len(roots)-1])
			roots = roots[:len(roots)-1]
			if n != nil && !want[n] {
				want[n] = true
				roots = append(roots, n.T, n.E)
			}
		}
		var before []*DdNode
		for _, table := range []*uniqueTable{&dd.unique, &dd.addUnique, &dd.zddUnique} {
			for _, n := range table.entries {
				before = append(before, n)
			}
		}
		for _, n := range dd.constants {
			before = append(before, n)
		}
		if got := int(Cudd_GarbageCollect(dd, true)); got != len(before)-len(want) {
			t.Fatalf("seed %d collected %d, want %d", seed, got, len(before)-len(want))
		}
		for _, n := range before {
			if (n.manager == dd) != want[n] {
				t.Fatalf("seed %d node %d live %t, want %t", seed, n.ID, n.manager == dd, want[n])
			}
		}
		if saturated.Ref != math.MaxUint32 || Cudd_GarbageCollect(dd, true) != 0 {
			t.Fatalf("seed %d second collection", seed)
		}
		if err := Cudd_CheckConsistency(dd); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		Cudd_Quit(dd)
	}
}
