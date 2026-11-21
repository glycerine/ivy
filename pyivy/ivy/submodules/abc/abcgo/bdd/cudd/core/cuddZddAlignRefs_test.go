package core

import (
	"slices"
	"testing"
)

func TestZDDVariablesFromBDDRetainCoversAcrossGrowth(t *testing.T) {
	dd := Cudd_Init(4, 4, 32, 64, 0)
	defer Cudd_Quit(dd)
	oldUniverse := dd.Univ[0]
	f := dd.zddNode(0, dd.zddNode(2, dd.One, dd.Zero), dd.zddNode(1, dd.One, dd.Zero))
	Cudd_Ref(oldUniverse)
	Cudd_Ref(f)
	if Cudd_ShuffleHeap(dd, []int{1, 0, 3, 2}) != 1 || Cudd_zddVarsFromBddVars(dd, 2) != 1 {
		t.Fatal("alignment failed")
	}
	if !slices.Equal(dd.InvPermZ, []int{2, 3, 0, 1, 6, 7, 4, 5}) {
		t.Fatal(dd.InvPermZ)
	}
	for _, n := range dd.zddUnique.entries {
		if n.inverse == nil || n.inverse.Index != n.Index || n.inverse.inverse != n {
			t.Errorf("node %d missing its matching inverse", n.ID)
		}
		want := uint32(0)
		if n == oldUniverse || n == f {
			want = 1
		}
		if got := Cudd_ExternalRefCount(dd, n); got != want {
			t.Errorf("node %d with %d external references, want %d", n.ID, got, want)
		}
	}
	Cudd_GarbageCollect(dd, true)
	assignment := make([]bool, dd.ZddSize)
	for mask := range 1 << dd.ZddSize {
		for i := range assignment {
			assignment[i] = mask&(1<<i) != 0
		}
		if zddTestContains(dd, f, assignment) != (mask == 5 || mask == 2) || zddTestContains(dd, oldUniverse, assignment) != (mask < 16) || !zddTestContains(dd, dd.Univ[0], assignment) {
			t.Fatalf("membership of set %d after extending and aligning", mask)
		}
	}
	Cudd_RecursiveDeref(dd, f)
	Cudd_RecursiveDeref(dd, oldUniverse)
	Cudd_GarbageCollect(dd, true)
	if got := dd.zddUnique.count; got != 8 {
		t.Fatalf("after releasing both original covers got %d keys, want 8", got)
	}
}

func TestExistingZDDAlignmentRunsNativeShufflePreprocessing(t *testing.T) {
	dd := Cudd_Init(3, 0, 32, 64, 0)
	defer Cudd_Quit(dd)
	hooks := 0
	Cudd_AddHook(dd, func(*DdManager, string, any) int { hooks++; return 1 }, CUDD_PRE_GC_HOOK)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 || hooks != 0 {
		t.Fatalf("first alignment hooks = %d", hooks)
	}
	root := dd.Univ[0]
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 || hooks != 1 || root != dd.Univ[0] {
		t.Fatalf("repeat alignment hooks = %d", hooks)
	}
}

func BenchmarkAlignExistingZDDVars(b *testing.B) {
	const width = 12
	dd := Cudd_Init(width, width, 32, 128, 0)
	defer Cudd_Quit(dd)
	for i := 0; i < width; i++ {
		for j := i + 1; j < width; j++ {
			Cudd_Ref(dd.zddNode(uint32(i), dd.zddNode(uint32(j), dd.One, dd.Zero), dd.Zero))
		}
	}
	root := dd.Univ[0]
	b.ReportAllocs()
	for b.Loop() {
		if Cudd_zddVarsFromBddVars(dd, 1) == 0 || dd.Univ[0] != root {
			b.Fatal("alignment")
		}
	}
}

func TestExistingUnitMultiplicityAlignmentReusesBDDOrder(t *testing.T) {
	dd := Cudd_Init(12, 12, 32, 64, 0)
	defer Cudd_Quit(dd)
	allocs := testing.AllocsPerRun(50, func() {
		if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
			t.Fatal("alignment")
		}
	})
	if allocs != 0 {
		t.Fatalf("repeated twelve-variable alignment allocated %.0f times", allocs)
	}
}
