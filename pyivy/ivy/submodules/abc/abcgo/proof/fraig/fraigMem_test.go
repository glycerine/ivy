package fraig

import "testing"

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
