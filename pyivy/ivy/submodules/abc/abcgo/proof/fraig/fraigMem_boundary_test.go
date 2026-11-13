package fraig

import "testing"

// fraigMem.c uses 1024 entries below 64 bytes, a 64 KiB target for
// intermediate sizes, and a minimum of eight entries per chunk.
func TestFixedMemorySourceChunkBoundaries(t *testing.T) {
	for _, tc := range []struct{ size, entries int32 }{{63, 1024}, {64, 1024}, {65, 1008}, {8192, 8}, {8193, 8}} {
		p := Fraig_MemFixedStart(tc.size)
		for i := int32(0); i < tc.entries; i++ {
			Fraig_MemFixedEntryFetch(p)
		}
		if got, want := Fraig_MemFixedReadMemUsage(p), tc.size*tc.entries; got != want {
			t.Fatalf("size=%d first chunk=%d want=%d", tc.size, got, want)
		}
		Fraig_MemFixedEntryFetch(p)
		if got, want := Fraig_MemFixedReadMemUsage(p), 2*tc.size*tc.entries; got != want {
			t.Fatalf("size=%d second chunk=%d want=%d", tc.size, got, want)
		}
		Fraig_MemFixedRestart(p)
		if got, want := Fraig_MemFixedReadMemUsage(p), tc.size*tc.entries; got != want {
			t.Fatalf("size=%d restart=%d want=%d", tc.size, got, want)
		}
		Fraig_MemFixedStop(p, 0)
	}
}

// The source recycle operation pushes entries onto an intrusive stack.
func TestFixedMemorySourceRecycleOrder(t *testing.T) {
	p := Fraig_MemFixedStart(16)
	a, b, c := Fraig_MemFixedEntryFetch(p), Fraig_MemFixedEntryFetch(p), Fraig_MemFixedEntryFetch(p)
	Fraig_MemFixedEntryRecycle(p, a)
	Fraig_MemFixedEntryRecycle(p, c)
	Fraig_MemFixedEntryRecycle(p, b)
	for i, want := range [][]byte{b, c, a} {
		got := Fraig_MemFixedEntryFetch(p)
		if &got[0] != &want[0] {
			t.Fatalf("fetch %d did not follow source LIFO order", i)
		}
	}
	Fraig_MemFixedStop(p, 0)
}
