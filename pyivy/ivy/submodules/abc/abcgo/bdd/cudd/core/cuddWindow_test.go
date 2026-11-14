package core

import (
	"reflect"
	"strings"
	"testing"
)

func TestWindow4CodesMatchSourceHamiltonianCircuit(t *testing.T) {
	want := []string{"ABCD", "BACD", "BADC", "ABDC", "ADBC", "ADCB", "DACB", "DABC", "DBAC", "BDAC", "BDCA", "DBCA", "DCBA", "DCAB", "CDAB", "CDBA", "CBDA", "BCDA", "BCAD", "CBAD", "CABD", "CADB", "ACDB", "ACBD"}
	patterns := window4Patterns()
	if len(patterns) != len(want) {
		t.Fatalf("got %d source permutations, want %d", len(patterns), len(want))
	}
	for i, item := range patterns {
		var got strings.Builder
		for _, index := range item.p {
			got.WriteByte("ABCD"[index])
		}
		if got.String() != want[i] {
			t.Fatalf("source code %d: got %s, want %s", item.code, got.String(), want[i])
		}
	}
	wantTies := []bool{false, false, false, true, false, true, false, false, false, true, true, false, true, true, true, true, true, true, true, true, true, true, true, true}
	for i, item := range patterns {
		if item.tie != wantTies[i] {
			t.Fatalf("source code %d tie rule: got %v, want %v", item.code, item.tie, wantTies[i])
		}
	}
}

func TestWindow4SourceTieSelection(t *testing.T) {
	// Code 13 follows code 7 and does not replace it on a tie in the C source.
	code, cost := window4Select(1, 10, 7, 5, false)
	code, cost = window4Select(code, cost, 13, 5, false)
	if code != 7 || cost != 5 {
		t.Fatalf("strict source case selected code %d at cost %d, want code 7", code, cost)
	}
	// Code 8 explicitly has the lower-code tie clause and replaces code 13.
	code, cost = window4Select(1, 10, 13, 5, false)
	code, cost = window4Select(code, cost, 8, 5, true)
	if code != 8 || cost != 5 {
		t.Fatalf("tie-enabled source case selected code %d at cost %d, want code 8", code, cost)
	}
}

func TestWindowCostOmitsIsolatedButCountsReferencedProjection(t *testing.T) {
	dd := NewBDDManager(2)
	base := windowCost(dd, dd.InvPerm)
	Cudd_Ref(dd.Vars[0])
	if got := windowCost(dd, dd.InvPerm); got != base+1 {
		t.Fatalf("referenced projection contribution: got %d, baseline %d", got, base)
	}
}

func TestWindowReorderBoundsAndPreservesRoots(t *testing.T) {
	dd := NewBDDManager(4)
	root := Cudd_bddIte(dd, dd.Vars[2], dd.Vars[1], dd.Vars[0])
	Cudd_Ref(root)
	before := reorderTruth(t, dd, root)
	start := append([]int(nil), dd.InvPerm...)
	if Cudd_WindowReorder(dd, 2, 2, CUDD_REORDER_WINDOW2) != 0 {
		t.Fatal("width-two window accepted a single level")
	}
	if Cudd_WindowReorder(dd, -1, 2, CUDD_REORDER_WINDOW2) != 0 || Cudd_WindowReorder(dd, 0, 4, CUDD_REORDER_WINDOW3) != 0 {
		t.Fatal("invalid window bounds accepted")
	}
	if !reflect.DeepEqual(start, dd.InvPerm) {
		t.Fatal("invalid window bounds changed the order")
	}
	if Cudd_WindowReorder(dd, 0, 3, CUDD_REORDER_WINDOW3_CONV) != 1 {
		t.Fatal("valid converging window reorder failed")
	}
	if !reflect.DeepEqual(before, reorderTruth(t, dd, root)) {
		t.Fatal("window reorder changed a referenced function")
	}
	for level, index := range dd.InvPerm {
		if dd.Perm[index] != level {
			t.Fatalf("Perm/InvPerm disagree: %v %v", dd.Perm, dd.InvPerm)
		}
	}
}

func TestWindow4SwapBoundary(t *testing.T) {
	dd := NewBDDManager(4)
	if Cudd_WindowReorder(dd, 0, 2, CUDD_REORDER_WINDOW4) != 1 {
		t.Fatal("width-four method did not fall back to width three")
	}
	if Cudd_WindowReorder(dd, 0, 3, CUDD_REORDER_WINDOW4) != 1 {
		t.Fatal("window ending at the final level failed")
	}
	if !reorderValidOrder(dd.InvPerm, len(dd.Vars)) {
		t.Fatalf("invalid resulting variable order: %v", dd.InvPerm)
	}
}

func TestReduceHeapDispatchesWindowAndHonorsMinsize(t *testing.T) {
	dd := NewBDDManager(4)
	before := append([]int(nil), dd.InvPerm...)
	if Cudd_ReduceHeap(dd, CUDD_REORDER_WINDOW4_CONV, windowPreprocessSize(dd)+1) != 1 {
		t.Fatal("window reorder failed at the minsize boundary")
	}
	if !reflect.DeepEqual(before, dd.InvPerm) || Cudd_ReadReorderings(dd) != 0 {
		t.Fatal("minsize skip changed ordering or reorder count")
	}
	if Cudd_ReduceHeap(dd, CUDD_REORDER_WINDOW4_CONV, 0) != 1 || Cudd_ReadReorderings(dd) != 1 {
		t.Fatal("window method was not dispatched by Cudd_ReduceHeap")
	}
}
