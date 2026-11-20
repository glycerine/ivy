package core

import (
	"slices"
	"testing"
)

func TestInteractionMatrixTracksSharedSupport(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := Cudd_bddIthVar(dd, 0), Cudd_bddIthVar(dd, 1), Cudd_bddIthVar(dd, 2)
	if Cudd_bddAnd(dd, x, y) == nil {
		t.Fatal("failed to construct shared-support node")
	}
	Cudd_ComputeInteractions(dd)
	if !Cudd_TestInteraction(dd, 0, 1) {
		t.Fatal("support variables should interact")
	}
	if Cudd_TestInteraction(dd, 0, 2) || Cudd_TestInteraction(dd, 1, 2) {
		t.Fatal("independent variable should not interact")
	}
	Cudd_SetInteraction(dd, 0, 2)
	if !Cudd_TestInteraction(dd, 2, 0) {
		t.Fatal("interaction relation should be symmetric")
	}
	_ = z
}

func TestInteractionsIncludeADDAndSeparateIndependentRoots(t *testing.T) {
	dd := NewBDDManager(75)
	defer Cudd_Quit(dd)
	// A shared tail occurs in two ADD roots, one of which is unreferenced.
	tail := dd.addNode(64, Cudd_addIthVar(dd, 74), dd.Zero)
	root := dd.addNode(0, tail, Cudd_addIthVar(dd, 2))
	Cudd_Ref(root)
	dd.addNode(1, tail, dd.One)
	dd.node(3, Cudd_Not(dd.Vars[70]), dd.One)
	Cudd_ComputeInteractions(dd)
	components := [][]int{{0, 2, 64, 74}, {1, 64, 74}, {3, 70}}
	for i := 0; i < 75; i++ {
		for j := i + 1; j < 75; j++ {
			want := false
			for _, group := range components {
				want = want || slices.Contains(group, i) && slices.Contains(group, j)
			}
			if Cudd_TestInteraction(dd, i, j) != want || Cudd_TestInteraction(dd, j, i) != want {
				t.Errorf("interaction (%d,%d), want %t", i, j, want)
			}
		}
	}
	Cudd_RecursiveDeref(dd, root)
	Cudd_GarbageCollect(dd, true)
	Cudd_ComputeInteractions(dd)
	if Cudd_TestInteraction(dd, 0, 74) || Cudd_TestInteraction(dd, 3, 70) {
		t.Fatal("rebuilding retained interactions from collected roots")
	}
}

func BenchmarkComputeInteractions(b *testing.B) {
	for _, tc := range []struct {
		name          string
		groups, width int
	}{{"IndependentCones", 32, 24}, {"DeepChain", 1, 128}} {
		b.Run(tc.name, func(b *testing.B) {
			dd := NewBDDManager(tc.groups * tc.width)
			defer Cudd_Quit(dd)
			for group := 0; group < tc.groups; group++ {
				first := group * tc.width
				root := dd.Vars[first+tc.width-1]
				for i := first + tc.width - 2; i >= first; i-- {
					root = dd.node(uint32(i), root, dd.LogicZero)
				}
				Cudd_Ref(root)
			}
			Cudd_ComputeInteractions(dd)
			for i := range len(dd.Vars) {
				for j := i + 1; j < len(dd.Vars); j++ {
					if Cudd_TestInteraction(dd, i, j) != (i/tc.width == j/tc.width) {
						b.Fatalf("wrong pair (%d,%d)", i, j)
					}
				}
			}
			b.ReportAllocs()
			for b.Loop() {
				Cudd_ComputeInteractions(dd)
				if !Cudd_TestInteraction(dd, 0, tc.width-1) || Cudd_TestInteraction(dd, 0, tc.width) {
					b.Fatal("incorrect cone support")
				}
			}
		})
	}
}

func TestInteractionsSurviveAddingVariables(t *testing.T) {
	dd := NewBDDManager(15)
	defer Cudd_Quit(dd)
	Cudd_SetInteraction(dd, 2, 4)
	Cudd_SetInteraction(dd, 9, 14)
	Cudd_bddIthVar(dd, 18)
	for i := range 19 {
		for j := i + 1; j < 19; j++ {
			want := i == 2 && j == 4 || i == 9 && j == 14
			if Cudd_TestInteraction(dd, i, j) != want {
				t.Errorf("after variable growth: pair (%d,%d), want %t", i, j, want)
			}
		}
	}
	if !Cudd_DestroyBDDSubtables(dd, 6) || Cudd_bddIthVar(dd, 18) == nil {
		t.Fatal("failed to remove and replace the last six variables")
	}
	if !Cudd_TestInteraction(dd, 2, 4) || Cudd_TestInteraction(dd, 9, 14) {
		t.Fatal("interaction with a destroyed variable survived its replacement")
	}
}

// Exhaustively compare the packed relation with each constructed support, including
// supports that cross word boundaries and the opposite variable order.
func TestPackedInteractionsMatchSupportPairs(t *testing.T) {
	const size = 139
	dd := NewBDDManager(size)
	defer Cudd_Quit(dd)
	order := make([]int, size)
	for i := range order {
		order[i] = size - 1 - i
	}
	if Cudd_ShuffleHeap(dd, order) == 0 {
		t.Fatal("shuffle")
	}
	want := make(map[[2]int]bool)
	for group := 0; group < 25; group++ {
		root := dd.One
		var support []int
		for i := 0; i < size; i++ {
			if (i*7+group*31)%43 >= group+2 && (i+3*group)%17 != 0 {
				continue
			}
			if group%2 == 0 {
				root = dd.node(uint32(i), root, dd.LogicZero)
			} else {
				root = dd.addNode(uint32(i), root, dd.Zero)
			}
			for _, j := range support {
				want[[2]int{j, i}] = true
			}
			support = append(support, i)
		}
		Cudd_Ref(root)
	}
	for range 2 {
		Cudd_ComputeInteractions(dd)
		for i := range size {
			for j := i + 1; j < size; j++ {
				if Cudd_TestInteraction(dd, i, j) != want[[2]int{i, j}] {
					t.Fatalf("wrong pair (%d,%d)", i, j)
				}
			}
		}
	}
}
