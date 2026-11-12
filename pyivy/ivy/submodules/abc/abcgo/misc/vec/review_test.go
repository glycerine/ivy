package vec

import "testing"

func TestSingleSliceTracksLogicalSizeAndCapacity(t *testing.T) {
	p := Vec_IntAllocExact(4)
	if len(p.data) != 0 || cap(p.data) != 4 {
		t.Fatal("allocation stores cached counters instead of slice bounds")
	}
	Vec_IntPush(p, 9)
	if len(p.data) != 1 || cap(p.data) != 4 {
		t.Fatal("push")
	}
	Vec_IntGrow(p, 20)
	if len(p.data) != 1 || cap(p.data) != 20 || p.data[0] != 9 {
		t.Fatal("grow changed live data")
	}
	Vec_IntClear(p)
	if len(p.data) != 0 || cap(p.data) != 20 {
		t.Fatal("clear")
	}
	Vec_IntPush(p, 7)
	released := Vec_IntReleaseArray(p)
	if len(released) != 1 || released[0] != 7 || len(p.data) != 0 || cap(p.data) != 0 {
		t.Fatal("release includes unused capacity")
	}
}
func TestFailedReservationsAreAtomic(t *testing.T) {
	ints := Vec_IntAllocExact(1)
	Vec_IntPush(ints, 4)
	if Vec_IntFetch(ints, 1) != nil || Vec_IntSize(ints) != 1 || Vec_IntEntry(ints, 0) != 4 {
		t.Fatal("int reservation changed state")
	}
	words := Vec_WrdAllocExact(1)
	Vec_WrdPush(words, 8)
	if Vec_WrdFetch(words, 1) != nil || Vec_WrdSize(words) != 1 || Vec_WrdEntry(words, 0) != 8 {
		t.Fatal("word reservation changed state")
	}
}
func TestPointerRemovalReleasesReferences(t *testing.T) {
	p := Vec_PtrStart(3)
	for i := range 3 {
		Vec_PtrWriteEntry(p, i, new(int))
	}
	Vec_PtrShrink(p, 1)
	if p.storage()[1] != nil || p.storage()[2] != nil {
		t.Fatal("shrink retained dead references")
	}
	Vec_PtrPush(p, new(int))
	Vec_PtrPop(p)
	if p.storage()[1] != nil {
		t.Fatal("pop retained dead reference")
	}
	Vec_PtrClear(p)
	for _, x := range p.storage() {
		if x != nil {
			t.Fatal("clear retained dead references")
		}
	}
}
func TestEmptyPointerUniqifyCounts(t *testing.T) {
	p := Vec_PtrAlloc(0)
	counts := Vec_IntAlloc(0)
	Vec_PtrUniqify2(p, nil, nil, counts)
	if Vec_IntSize(counts) != 0 {
		t.Fatal("empty input produced a count")
	}
}
func TestTypedNilAttributeIsAbsent(t *testing.T) {
	var missing *int
	starts := 0
	p := Vec_AttAlloc(1, nil, nil, func(any) any { starts++; return missing }, nil)
	Vec_AttEntry(p, 0)
	Vec_AttEntry(p, 0)
	if starts != 2 {
		t.Fatal("typed nil treated as initialized object")
	}
}
func TestVariableIntegerRejectsInvalidEncoding(t *testing.T) {
	mustPanic := func(f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatal("invalid encoding accepted")
			}
		}()
		f()
	}
	mustPanic(func() { Vec_StrPutI(Vec_StrAlloc(0), -1) })
	mustPanic(func() {
		p := Vec_StrAllocArray([]byte{0xff, 0xff, 0xff, 0xff, 0x7f}, 5)
		pos := 0
		Vec_StrGetI(p, &pos)
	})
}

func TestPagedMemoryWritesKeepHashConsistent(t *testing.T) {
	p := Vec_MemAlloc(1, 2)
	Vec_MemHashAlloc(p, 3)
	Vec_MemHashInsert(p, []uint64{1})
	Vec_MemHashInsert(p, []uint64{2})
	Vec_MemWriteEntry(p, 0, []uint64{2})
	if Vec_MemHashInsert(p, []uint64{2}) != 0 || p.nEntries != 2 {
		t.Fatal("write left stale key or duplicate record")
	}
	Vec_MemClean(p, 2)
	if Vec_MemHashInsert(p, []uint64{0}) != 0 || p.nEntries != 2 {
		t.Fatal("clean left stale hash")
	}
	Vec_MemPush(p, []uint64{3})
	if Vec_MemHashInsert(p, []uint64{3}) != 2 || p.nEntries != 3 {
		t.Fatal("raw append not indexed")
	}
	q := Vec_MemAlloc(1, 2)
	Vec_MemPush(q, []uint64{8})
	Vec_MemHashAlloc(q, 3)
	if Vec_MemHashInsert(q, []uint64{8}) != 0 || q.nEntries != 1 {
		t.Fatal("hash allocation ignored existing records")
	}
}
