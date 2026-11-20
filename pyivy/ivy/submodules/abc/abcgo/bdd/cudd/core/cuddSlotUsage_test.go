package core

import (
	"math"
	"testing"
)

func slotFixture(step func(string, *DdManager)) *DdManager {
	dd := Cudd_Init(9, 9, 8, 256, 0)
	step("initial", dd)
	for i := 1; i < 9; i++ {
		for j := 1; j < 9; j++ {
			if i != j {
				Cudd_Ref(dd.node(0, dd.Vars[i], dd.Vars[j]))
			}
		}
	}
	step("BDD", dd)
	for i := 1; i < 9; i++ {
		for j := 1; j < 9; j++ {
			if i != j {
				Cudd_Ref(dd.zddNode(0, dd.Univ[i], dd.Univ[j]))
			}
		}
	}
	step("ZDD", dd)
	for i := 0; i < 80; i++ {
		Cudd_Ref(Cudd_addConst(dd, float64(i+2)))
	}
	step("constants", dd)
	return dd
}

func TestUniqueTableOccupancyMatchesNative(t *testing.T) {
	// From test_vectors/perf_slots.c, including collisions at the two top levels.
	want := map[string]struct {
		slots, used uint32
		expected    float64
	}{
		"initial": {152, 22, .132027636514}, "BDD": {160, 32, .216714318172},
		"ZDD": {168, 41, .293335601578}, "constants": {192, 69, .394867469363},
	}
	dd := slotFixture(func(stage string, dd *DdManager) {
		w := want[stage]
		slots := Cudd_ReadSlots(dd)
		actual, expected := Cudd_ReadUsedSlots(dd), Cudd_ExpectedUsedSlots(dd)
		if slots != w.slots || math.Abs(actual-float64(w.used)/float64(w.slots)) > 1e-12 || math.Abs(expected-w.expected) > 1e-12 {
			t.Errorf("%s: slots %d occupied %.0f expected %.12f, want %d, %d, %.12f", stage, slots, actual*float64(slots), expected, w.slots, w.used, w.expected)
		}
	})
	Cudd_Quit(dd)
	if Cudd_ReadUsedSlots(dd) != 0 || Cudd_ExpectedUsedSlots(nil) != 0 {
		t.Error("occupancy for closed or missing manager")
	}
}

func TestSlotUsageCountsDeadKeysUntilCollection(t *testing.T) {
	dd := Cudd_Init(3, 2, 16, 64, 0)
	defer Cudd_Quit(dd)
	check := func(want float64, used float64) {
		t.Helper()
		if got := Cudd_ExpectedUsedSlots(dd); math.Abs(got-want) > 1e-12 {
			t.Errorf("expected occupancy %.12f want %.12f", got, want)
		}
		if got := Cudd_ReadUsedSlots(dd); math.Abs(got-used/96) > 1e-12 {
			t.Errorf("used slots %.0f want %.0f", got*96, used)
		}
	}
	check(.087355650477, 9)
	root := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	Cudd_Ref(root)
	check(.096841677182, 10)
	Cudd_RecursiveDeref(dd, root)
	check(.096841677182, 10)
	Cudd_GarbageCollect(dd, true)
	check(.087355650477, 9)
}

func BenchmarkReadUniqueTableOccupancy(b *testing.B) {
	dd := slotFixture(func(string, *DdManager) {})
	defer Cudd_Quit(dd)
	b.ReportAllocs()
	for b.Loop() {
		got := Cudd_ReadUsedSlots(dd) + Cudd_ExpectedUsedSlots(dd)
		if got <= 0 {
			b.Fatal(got)
		}
	}
}

func TestSlotGrowthSeparatesSharedADDAndZddLevels(t *testing.T) {
	dd := Cudd_Init(10, 3, 8, 64, 0)
	defer Cudd_Quit(dd)
	if Cudd_ShuffleHeap(dd, []int{1, 0, 2, 3, 4, 5, 6, 7, 8, 9}) == 0 {
		t.Fatal("failed to permute BDD levels")
	}
	top := uint32(1)
	// Neither family reaches the limit independently. Their shared native
	// level reaches 33 keys, and only the next fresh key grows it.
	for i := 2; i <= 5; i++ {
		for j := 6; j <= 9; j++ {
			Cudd_Ref(dd.node(top, dd.Vars[i], dd.Vars[j]))
			Cudd_Ref(dd.addNode(top, Cudd_addIthVar(dd, i), Cudd_addIthVar(dd, j)))
		}
	}
	width := dd.bddTableSlots[0]
	Cudd_Ref(dd.node(top, dd.Vars[2], dd.Vars[6]))
	if width != 8 || dd.bddTableSlots[0] != 8 {
		t.Fatalf("BDD hit grew shared level to %d", dd.bddTableSlots[0])
	}
	Cudd_Ref(dd.addNode(top, dd.One, dd.Zero))
	if dd.bddTableSlots[0] != 16 || dd.bddTableSlots[1] != 8 || dd.zddTableSlots[0] != 8 {
		t.Fatalf("BDD %v and ZDD %v after inserting at index 1", dd.bddTableSlots, dd.zddTableSlots)
	}
}

func TestSlotsSupportSparseWideTables(t *testing.T) {
	dd := Cudd_InitTable(1, 0, 1<<30, 0)
	defer Cudd_Quit(dd)
	if dd == nil {
		t.Fatal("could not create logical two-billion-slot table")
	}
	if used := Cudd_ReadUsedSlots(dd) * float64(Cudd_ReadSlots(dd)); used != 4 {
		t.Fatalf("four constants occupied %.0f buckets", used)
	}
}
