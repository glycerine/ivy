package core

import "testing"

func readCacheIndex(slots *computedSlots, slot uint32) computedEntry {
	if pos := slots.position(slot); pos != 0 {
		return *slots.record(pos)
	}
	return computedEntry{}
}

func TestFourComputedRecordsNeedNoHeapIndex(t *testing.T) {
	dd := NewBDDManager(0)
	defer Cudd_Quit(dd)
	entry := computedEntry{result: dd.One}
	if alloc := testing.AllocsPerRun(50, func() {
		var slots computedSlots
		for i, slot := range []uint32{0, 1 << 18, 7, 99} {
			slots.prepare(1<<20, i+1)
			slots.write(slot, entry)
			if readCacheIndex(&slots, slot) != entry {
				t.Fatalf("missing slot %d", slot)
			}
		}
	}); alloc != 0 {
		t.Fatalf("four inline records allocated %.0f times", alloc)
	}
}

func TestSparseComputedIndexWithStridedSlots(t *testing.T) {
	dd := NewBDDManager(0)
	defer Cudd_Quit(dd)
	var slots computedSlots
	const size = 1 << 23
	for i := range 300 {
		slot := uint32(i) << 12
		entry := computedEntry{key: computedKey{op: uintptr(i + 1)}, result: dd.One}
		slots.prepare(size, slots.used+1)
		if slots.write(slot, entry) {
			t.Fatalf("unexpected overwrite at %d", i)
		}
	}
	for i := range 300 {
		slot := uint32(i) << 12
		if got := readCacheIndex(&slots, slot); got.result != dd.One || got.key.op != uintptr(i+1) {
			t.Fatalf("lost record %d", i)
		}
		if got := readCacheIndex(&slots, slot+1); got.result != nil {
			t.Fatalf("false match for %d", slot+1)
		}
	}
	if removed := slots.invalidate(func(e computedEntry) bool { return e.key.op%2 == 0 }); removed != 150 {
		t.Fatal("removed", removed)
	}
	for i := range 300 {
		slot := uint32(i) << 12
		entry := computedEntry{key: computedKey{op: uintptr(600 - i)}, result: dd.Zero}
		if slots.write(slot, entry) != (i%2 == 0) || readCacheIndex(&slots, slot) != entry {
			t.Fatal("replacement", i)
		}
	}
	if slots.used != 300 || slots.index != nil {
		t.Fatal("sparse size", slots.used)
	}
}
