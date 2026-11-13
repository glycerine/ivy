package fraig

import (
	"fmt"
	"testing"
)

func TestFixedMemoryGrowthRecycleAndRestart(t *testing.T) {
	p := Fraig_MemFixedStart(8)
	if got := Fraig_MemFixedReadMemUsage(p); got != 0 {
		t.Fatalf("initial usage = %d", got)
	}
	entries := make([][]byte, 1025)
	for i := range entries {
		entries[i] = Fraig_MemFixedEntryFetch(p)
		entries[i][0] = byte(i)
	}
	if got := Fraig_MemFixedReadMemUsage(p); got != 2*1024*8 {
		t.Fatalf("usage after growth = %d", got)
	}
	Fraig_MemFixedEntryRecycle(p, entries[0])
	if got := Fraig_MemFixedEntryFetch(p); &got[0] != &entries[0][0] {
		t.Fatal("recycled entry was not reused")
	}
	if entries[1][0] != 1 {
		t.Fatal("entries overlap")
	}
	Fraig_MemFixedRestart(p)
	if got := Fraig_MemFixedReadMemUsage(p); got != 1024*8 {
		t.Fatalf("usage after restart = %d", got)
	}
	if got := len(Fraig_MemFixedEntryFetch(p)); got != 8 {
		t.Fatalf("entry size after restart = %d", got)
	}
	Fraig_MemFixedStop(p, 0)
}

func TestFixedMemoryRestartBeforeFirstFetch(t *testing.T) {
	p := Fraig_MemFixedStart(128)
	Fraig_MemFixedRestart(p)
	if got := len(Fraig_MemFixedEntryFetch(p)); got != 128 {
		t.Fatalf("entry size = %d", got)
	}
}

// The chunk policy and retained-first-chunk restart are from fraigMem.c:75-83
// and :202-226. Entry sizes are at least pointer-sized, as required by C's
// intrusive free list.
func TestFixedMemoryChunkBoundariesAndRestartOrder(t *testing.T) {
	for _, tc := range []struct{ size, entries int32 }{{8, 1024}, {63, 1024}, {64, 1024}, {1000, 65}, {9000, 8}} {
		t.Run(fmt.Sprint(tc.size), func(t *testing.T) {
			p := Fraig_MemFixedStart(tc.size)
			defer Fraig_MemFixedStop(p, 0)
			entries := make([][]byte, tc.entries+1)
			for i := range entries {
				entries[i] = Fraig_MemFixedEntryFetch(p)
			}
			if got, want := Fraig_MemFixedReadMemUsage(p), 2*tc.size*tc.entries; got != want {
				t.Fatalf("two chunks use %d bytes, want %d", got, want)
			}
			Fraig_MemFixedEntryRecycle(p, entries[1])
			Fraig_MemFixedEntryRecycle(p, entries[2])
			if got := Fraig_MemFixedEntryFetch(p); &got[0] != &entries[2][0] {
				t.Fatal("most recently recycled entry was not first")
			}
			if got := Fraig_MemFixedEntryFetch(p); &got[0] != &entries[1][0] {
				t.Fatal("second recycled entry was not next")
			}
			Fraig_MemFixedRestart(p)
			if got, want := Fraig_MemFixedReadMemUsage(p), tc.size*tc.entries; got != want {
				t.Fatalf("restart uses %d bytes, want %d", got, want)
			}
			if got := Fraig_MemFixedEntryFetch(p); &got[0] != &entries[0][0] {
				t.Fatal("restart did not reuse first entry of first chunk")
			}
		})
	}
}
