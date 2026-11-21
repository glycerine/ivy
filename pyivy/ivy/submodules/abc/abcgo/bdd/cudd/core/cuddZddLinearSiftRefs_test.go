package core

import (
	"slices"
	"sync"
	"testing"
)

var nativeLinearSifts = []struct {
	input, result uint32
	order         [5]int
	keys          int
}{
	{0x420b6e3c, 0x7b86081e, [5]int{4, 1, 3, 2, 0}, 14},
	{0xff000111, 0xff000111, [5]int{3, 4, 0, 1, 2}, 7},
	{0x007dc06a, 0x0caf00a6, [5]int{0, 2, 4, 3, 1}, 12},
	{0xa78f2469, 0xeb8127c5, [5]int{4, 1, 3, 0, 2}, 16},
}

func shuffleFiveVarCover(input uint32) (*DdManager, *DdNode) {
	dd := Cudd_Init(0, 5, 32, 128, 0)
	var build func(int, int) *DdNode
	build = func(level, set int) *DdNode {
		if level == 5 {
			if input&(1<<set) != 0 {
				return dd.One
			}
			return dd.Zero
		}
		return dd.zddNode(uint32(level), build(level+1, set|(1<<level)), build(level+1, set))
	}
	f := build(0, 0)
	Cudd_Ref(f)
	if Cudd_zddShuffleHeap(dd, []int{3, 0, 4, 1, 2}) == 0 {
		panic("shuffle")
	}
	return dd, f
}

func siftFiveVarCover(input uint32) (*DdManager, *DdNode) {
	dd, f := shuffleFiveVarCover(input)
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_LINEAR, 0) == 0 {
		panic("sift")
	}
	return dd, f
}

func TestLinearSiftDiscardsUnreachableTrialNodes(t *testing.T) {
	dd, _ := shuffleFiveVarCover(nativeLinearSifts[0].input)
	defer Cudd_Quit(dd)
	// Direct callers can hold any existing handle, including released nodes
	// in the unique table. Only the sifter can observe new trial handles.
	var original []*DdNode
	for _, n := range dd.zddUnique.entries {
		original = append(original, n)
	}
	firstTrial := dd.nextID
	if Cudd_zddLinearSifting(dd, 0, 4) == 0 {
		t.Fatal("sift")
	}
	reachable := make(map[*DdNode]bool)
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		if n == nil || n.manager != dd {
			t.Fatal("retired an original handle or a reachable cofactor")
		}
		if reachable[n] || Cudd_IsConstant(n) {
			return
		}
		reachable[n] = true
		visit(n.T)
		visit(n.E)
	}
	for _, n := range original {
		visit(n)
	}
	for _, n := range dd.zddUnique.entries {
		if !n.childrenReleased {
			visit(n)
		}
	}
	unused := 0
	for _, n := range dd.zddUnique.entries {
		if n.ID >= firstTrial && !reachable[n] {
			unused++
		}
		if reachable[n] && dd.zddUnique.get(bddUniqueKey{n.Index, n.T, n.E}) != n {
			t.Fatalf("reachable handle %d is no longer canonical", n.ID)
		}
	}
	if unused != 0 {
		t.Fatalf("retained %d unreachable trial nodes", unused)
	}
}

func TestZDDLinearSiftMatchesNativeFiveVarCovers(t *testing.T) {
	for _, v := range nativeLinearSifts {
		dd, f := siftFiveVarCover(v.input)
		var result uint32
		for mask := range 32 {
			var assignment [5]bool
			for i := range assignment {
				assignment[i] = mask&(1<<i) != 0
			}
			if zddTestContains(dd, f, assignment[:]) {
				result |= 1 << mask
			}
		}
		if result != v.result || !slices.Equal(dd.InvPermZ, v.order[:]) || zddHeapSize(dd) != v.keys {
			t.Errorf("input %08x gave %08x order %v size %d, want %08x %v %d", v.input, result, dd.InvPermZ, zddHeapSize(dd), v.result, v.order, v.keys)
		}
		Cudd_Quit(dd)
	}
}

func TestZDDLinearSiftsOnIndependentManagers(t *testing.T) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, v := range nativeLinearSifts {
		wg.Go(func() {
			<-start
			dd, _ := siftFiveVarCover(v.input)
			if !validZddOrder(dd) {
				t.Error("invalid order")
			}
			Cudd_Quit(dd)
		})
	}
	close(start)
	wg.Wait()
}

func BenchmarkZDDLinearSiftFiveVarCovers(b *testing.B) {
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		dd, _ := siftFiveVarCover(nativeLinearSifts[i%4].input)
		Cudd_Quit(dd)
		i++
	}
}

func TestWideLinearZDDTransformReusesWorkspace(t *testing.T) {
	dd := Cudd_Init(0, 9, 32, 128, 0)
	defer Cudd_Quit(dd)
	for mask := 1; mask <= 16; mask++ {
		tail := dd.One
		for i := 6; i >= 0; i-- {
			if mask&(1<<i) != 0 {
				tail = dd.zddNode(uint32(2+i), tail, dd.Zero)
			}
		}
		for _, f := range allTwoVarZDDFamilies(dd, 0, 1, tail) {
			Cudd_Ref(f)
		}
	}
	allocs := testing.AllocsPerRun(20, func() {
		if !linearZddAdjacentInPlace(dd, 0, 1) || !linearZddAdjacentInPlace(dd, 0, 1) {
			t.Fatal("wide transform")
		}
	})
	if allocs != 0 {
		t.Fatalf("warm transform pair across two wide subtables allocated %.0f times", allocs)
	}
}
