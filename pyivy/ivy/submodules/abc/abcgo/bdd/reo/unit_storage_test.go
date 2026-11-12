package reo

import "testing"

func TestUnitStorageChunksAndReuse(t *testing.T) {
	s := unitStorage{nMemChunksAlloc: 2}
	units := make([]*reo_unit, unitChunkSize+1)
	seen := map[*reo_unit]bool{}
	for i := range units {
		u := s.get()
		if seen[u] {
			t.Fatal("unit dispensed twice")
		}
		seen[u] = true
		units[i] = u
		u.Sign = int32(i + 1)
	}
	if s.nMemChunks != 2 || s.nUnitsUsed != unitChunkSize+1 {
		t.Fatal(s.nMemChunks, s.nUnitsUsed)
	}
	// A later chunk-directory append must not move existing node storage.
	for i, u := range units {
		if u.Sign != int32(i+1) {
			t.Fatal("unstable address", i)
		}
	}
	s.recycle(units[3])
	if s.get() != units[3] || s.nUnitsUsed != unitChunkSize+1 {
		t.Fatal("LIFO recycle failed")
	}
}

func TestPlaneRecyclePreservesOrderAndResetsStatistics(t *testing.T) {
	s := unitStorage{nMemChunksAlloc: 1}
	a, b := s.get(), s.get()
	plane := reo_plane{statsWidth: 9, statsApl: 4.5, fSifted: 1}
	addUnitToPlane(&plane, a)
	addUnitToPlane(&plane, b)
	if plane.pHead != b || b.Next != a || a.Next != nil || plane.statsNodes != 2 {
		t.Fatal("plane ordering")
	}
	count := s.nUnitsUsed
	s.recyclePlane(&plane)
	if plane != (reo_plane{}) || s.nUnitsUsed != count {
		t.Fatal("plane reset/counter compatibility")
	}
	if s.get() != b || s.get() != a {
		t.Fatal("bulk recycle order")
	}
	empty := reo_plane{statsWidth: 8}
	s.recyclePlane(&empty)
	if empty.statsWidth != 8 {
		t.Fatal("empty plane behavior differs from C")
	}
}

func TestUnitStorageStopAndRestart(t *testing.T) {
	s := unitStorage{nMemChunksAlloc: 1}
	old := s.get()
	s.stop()
	if s.pUnitFreeList != nil || s.pMemChunks != nil || s.nUnitsUsed != 0 || s.nMemChunks != 0 {
		t.Fatal("retained dispenser roots")
	}
	if next := s.get(); next == old || next == nil {
		t.Fatal("restart reused old storage")
	}
}

func TestUnitStorageCapacity(t *testing.T) {
	s := unitStorage{nMemChunksAlloc: 1}
	for i := 0; i < unitChunkSize; i++ {
		s.get()
	}
	defer func() {
		if recover() == nil {
			t.Fatal("missing capacity error")
		}
	}()
	s.get()
}
