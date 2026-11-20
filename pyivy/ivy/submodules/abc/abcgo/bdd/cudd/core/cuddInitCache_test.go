package core

import (
	"math/bits"
	"testing"
)

func TestCuddInitUsesNativeTableAndCacheSizes(t *testing.T) {
	for _, tc := range []struct {
		bdd, zdd, slots, cache        uint32
		memory                        uint64
		wantSlots, wantCache, wantMax uint32
	}{
		{3, 2, 9, 70, 1 << 20, 96, 64, 384},
		{0, 0, 0, 0, 0, 8, 4, 32},
		{36, 0, 256, 16384, 0, 9472, 16384, 37888},
		{9, 0, 16, 1, 63, 160, 64, 0},
	} {
		dd := Cudd_Init(tc.bdd, tc.zdd, tc.slots, tc.cache, tc.memory)
		if dd == nil {
			t.Fatal("could not initialize manager")
		}
		if s, c, m := Cudd_ReadSlots(dd), Cudd_ReadCacheSlots(dd), Cudd_ReadMaxCache(dd); s != tc.wantSlots || c != tc.wantCache || m != tc.wantMax {
			t.Errorf("init(%d,%d,%d,%d,%d) slots, cache, soft limit = %d, %d, %d, want %d, %d, %d", tc.bdd, tc.zdd, tc.slots, tc.cache, tc.memory, s, c, m, tc.wantSlots, tc.wantCache, tc.wantMax)
		}
		if len(dd.Vars) != int(tc.bdd) || len(dd.Univ) != int(tc.zdd) {
			t.Error("missing projection or universe")
		}
		if bits.UintSize == 64 {
			// Original ABC's 64-bit DdNode and DdCache records both occupy 40 bytes.
			memory := tc.memory
			if memory == 0 {
				memory = 64 << 20
			}
			if h, l := Cudd_ReadMaxCacheHard(dd), Cudd_ReadLooseUpTo(dd); h != uint32(memory/(40*3)) || l != uint32(memory/(40*5)) {
				t.Errorf("memory hint %d gave hard cache limit %d and loose limit %d", memory, h, l)
			}
			if got := Cudd_ReadMaxMemory(dd); got != ^uint64(0) {
				t.Errorf("initial hard memory limit = %d, want unlimited", got)
			}
		}
		Cudd_Quit(dd)
	}
}

func TestCuddCacheAndUniqueLimitDefaults(t *testing.T) {
	dd := Cudd_Init(3, 2, 9, 70, 1<<20)
	defer Cudd_Quit(dd)
	Cudd_SetLooseUpTo(dd, 15)
	Cudd_SetMaxCacheHard(dd, 29)
	if Cudd_ReadLooseUpTo(dd) != 15 || Cudd_ReadMaxCacheHard(dd) != 29 {
		t.Fatal("explicit limits were lost")
	}
	Cudd_SetLooseUpTo(dd, 0)
	Cudd_SetMaxCacheHard(dd, 0)
	if bits.UintSize == 64 && (Cudd_ReadLooseUpTo(dd) != 335544 || Cudd_ReadMaxCacheHard(dd) != 559240) {
		t.Fatalf("reset yielded limits %d/%d, want ABC defaults 335544/559240", Cudd_ReadLooseUpTo(dd), Cudd_ReadMaxCacheHard(dd))
	}
}

func TestCuddCustomSlotsAfterVariableGrowth(t *testing.T) {
	dd := Cudd_Init(1, 1, 9, 16, 0)
	defer Cudd_Quit(dd)
	Cudd_bddIthVar(dd, 3)
	if got := Cudd_ReadSlots(dd); got != 96 {
		t.Errorf("growth to four BDD variables left %d slots, want 96", got)
	}
	if !dd.ensureZddVar(2) || Cudd_ReadSlots(dd) != 128 {
		t.Errorf("growth to three ZDD variables left %d slots, want 128", Cudd_ReadSlots(dd))
	}
	if !Cudd_InsertBDDSubtables(dd, 2, 1) || Cudd_ReadSlots(dd) != 160 {
		t.Errorf("inserting two subtables left %d slots, want 160", Cudd_ReadSlots(dd))
	}
}

func TestVariableInsertedAtLevelRetainsSubtableCapacities(t *testing.T) {
	for _, insert := range []func(*DdManager, int) *DdNode{Cudd_bddNewVarAtLevel, Cudd_addNewVarAtLevel} {
		dd := Cudd_InitTable(2, 0, 16, 128)
		Cudd_bddIthVar(dd, 1)
		if !Cudd_RehashBDDSubtable(dd, 0) {
			t.Fatal("rehash failed")
		}
		v := insert(dd, 0)
		if v == nil || len(dd.bddTableSlots) != 3 || dd.bddTableSlots[0] != 16 || dd.bddTableSlots[1] != 32 || dd.bddTableSlots[2] != 16 {
			t.Errorf("new variable at root left subtable capacities %v, want [16 32 16]", dd.bddTableSlots)
		}
		Cudd_Quit(dd)
	}
}

func TestZddVariableAlignmentAccountsForAllSubtables(t *testing.T) {
	dd := Cudd_Init(3, 0, 9, 32, 0)
	defer Cudd_Quit(dd)
	if Cudd_zddVarsFromBddVars(dd, 2) != 1 {
		t.Fatal("alignment failed")
	}
	if slots := Cudd_ReadSlots(dd); slots != 160 || len(dd.zddTableSlots) != 6 {
		t.Fatalf("six aligned ZDD variables have %d subtables and %d total slots, want six and 160", len(dd.zddTableSlots), slots)
	}
}

func TestCuddSubtableResizeDoublesCacheOnce(t *testing.T) {
	for _, zdd := range []bool{false, true} {
		dd := Cudd_Init(0, 0, 8, 4, 0)
		if zdd {
			Cudd_ResizeZddTable(dd, 9)
		} else {
			Cudd_ResizeBDDTable(dd, 9)
		}
		if got := Cudd_ReadCacheSlots(dd); got != 8 {
			t.Errorf("zdd=%v, grew cache to %d slots, want one doubling to 8", zdd, got)
		}
		Cudd_Quit(dd)
	}
}
