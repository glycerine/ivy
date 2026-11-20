package core

import "testing"

func TestCuddAllocNodeManagerOwnershipAndLimit(t *testing.T) {
	dd := Cudd_Init(0, 0, 0, 0, 0)
	if dd == nil {
		t.Fatal("Cudd_Init returned nil")
	}
	wantID := dd.nextID
	n := Cudd_AllocNode(dd)
	if n == nil || n.manager != dd || n.Ref != 0 || n.Index != 0 {
		t.Fatalf("unexpected allocated node: %#v", n)
	}
	if n.ID != wantID {
		t.Fatalf("allocated node ID = %d, want %d", n.ID, wantID)
	}
	limited := Cudd_Init(1, 0, 0, 0, 0)
	Cudd_SetMaxLive(limited, 0)
	if got := Cudd_AllocNode(limited); got != nil {
		t.Fatalf("allocation above max-live limit returned %#v, want nil", got)
	}
	if got := Cudd_ReadErrorCode(limited); got != CUDD_TOO_MANY_NODES {
		t.Fatalf("error code = %v, want CUDD_TOO_MANY_NODES", got)
	}
}

func TestCuddInitTableSizesSlotsAndPermutations(t *testing.T) {
	dd := Cudd_InitTable(3, 2, 9, 128)
	if dd == nil {
		t.Fatal("Cudd_InitTable returned nil")
	}
	if dd.tableBDDSize != 3 || dd.tableZddSize != 2 || dd.initSlots != 16 {
		t.Fatalf("table sizing = (%d, %d) with %d initial slots", dd.tableBDDSize, dd.tableZddSize, dd.initSlots)
	}
	if Cudd_ReadSlots(dd) != 96 {
		t.Fatalf("slot count = %d, want 96", Cudd_ReadSlots(dd))
	}
	if len(dd.PermZ) != 2 || dd.PermZ[1] != 1 || dd.InvPermZ[1] != 1 || Cudd_ReadLooseUpTo(dd) != 128 {
		t.Fatalf("ZDD table state or loose-up-to threshold was not initialized")
	}
	if Cudd_InitTable(1, 1, ^uint32(0), 0) != nil {
		t.Fatal("Cudd_InitTable accepted an unrepresentable slot table")
	}
}

func TestCuddAllocNodeLimitIncludesConstantsAndExcludesDead(t *testing.T) {
	dd := Cudd_Init(6, 0, 16, 64, 0)
	defer Cudd_Quit(dd)
	// Six isolated projections and four constants occupy live keys.
	Cudd_SetMaxLive(dd, 9)
	if got := Cudd_AllocNode(dd); got != nil || Cudd_ReadErrorCode(dd) != CUDD_TOO_MANY_NODES {
		t.Errorf("raw allocation exceeded nine live keys with result %p and status %d", got, Cudd_ReadErrorCode(dd))
	}
	for i, a := range dd.Vars {
		for _, b := range dd.Vars[i+1:] {
			n := Cudd_bddAnd(dd, a, b)
			Cudd_Ref(n)
			Cudd_RecursiveDeref(dd, n)
		}
	}
	Cudd_SetMaxLive(dd, 10)
	if got := Cudd_AllocNode(dd); got == nil {
		t.Fatal("fifteen dead conjunctions were charged against a ten-live-key limit")
	}
}
