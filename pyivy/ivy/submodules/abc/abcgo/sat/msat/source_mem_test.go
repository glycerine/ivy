package msat

import "testing"

func TestFixedPoolUsesSourceChunkAllocation(t *testing.T) {
	p := Msat_MmFixedStart(8)
	for i := 0; i < 1025; i++ {
		Msat_MmFixedEntryFetch(p)
	}
	if got := Msat_MmFixedReadMemUsage(p); got != 16384 {
		t.Fatalf("two C chunks = 16384 bytes, got %d", got)
	}
	Msat_MmFixedRestart(p)
	if got := Msat_MmFixedReadMemUsage(p); got != 8192 {
		t.Fatalf("restart retains first C chunk, got %d", got)
	}
}
func TestFlexPoolUsesSourceChunkAccounting(t *testing.T) {
	p := Msat_MmFlexStart()
	Msat_MmFlexEntryFetch(p, 1)
	if got := Msat_MmFlexReadMemUsage(p); got != 4096 {
		t.Fatalf("first chunk usage %d", got)
	}
	Msat_MmFlexEntryFetch(p, 5000)
	if got := Msat_MmFlexReadMemUsage(p); got != 14096 {
		t.Fatalf("oversize chunk usage %d", got)
	}
}
func TestStepPoolExcludesLargeAllocationsFromUsage(t *testing.T) {
	p := Msat_MmStepStart(2)
	if Msat_MmStepEntryFetch(p, 0) != nil {
		t.Fatal("C returns nil for zero bytes")
	}
	Msat_MmStepEntryFetch(p, 1000)
	if got := Msat_MmStepReadMemUsage(p); got != 0 {
		t.Fatalf("large C allocations are not counted, got %d", got)
	}
	Msat_MmStepEntryFetch(p, 8)
	if got := Msat_MmStepReadMemUsage(p); got != 8192 {
		t.Fatalf("pool chunk usage %d", got)
	}
}

func TestSolverClauseUsesAndRecyclesSourceArena(t *testing.T) {
	p := New(2)
	if !p.AddClause([]int{0, 2}) {
		t.Fatal("first clause")
	}
	if got := Msat_MmStepReadMemUsage(Msat_SolverReadMem(p)); got != 16384 {
		t.Fatalf("C allocates 1024 16-byte entries, got %d bytes", got)
	}
	first := p.vClauses.data[0]
	firstEntry := &first.arena[0]
	Msat_ClauseFree(p, first, 1)
	p.vClauses.data = p.vClauses.data[:0]
	if !p.AddClause([]int{1, 3}) {
		t.Fatal("second clause")
	}
	if &p.vClauses.data[0].arena[0] != firstEntry {
		t.Fatal("C recycles the freed arena entry")
	}
	if got := Msat_MmStepReadMemUsage(p.pMem); got != 16384 {
		t.Fatalf("recycled entry allocated another chunk, got %d", got)
	}
}
