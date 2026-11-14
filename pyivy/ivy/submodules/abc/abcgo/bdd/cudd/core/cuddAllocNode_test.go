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
	Cudd_SetMaxLive(dd, uint32(Cudd_ReadNodeCount(dd)))
	if got := Cudd_AllocNode(dd); got != nil {
		t.Fatalf("allocation at max-live limit returned %#v, want nil", got)
	}
	if got := Cudd_ReadErrorCode(dd); got != CUDD_TOO_MANY_NODES {
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
