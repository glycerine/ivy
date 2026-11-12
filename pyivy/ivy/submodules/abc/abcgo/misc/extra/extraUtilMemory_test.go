package extra

import "testing"

func TestFixedMemoryRecycleRestart(t *testing.T) {
	p := Extra_MmFixedStart(8)
	entry := Extra_MmFixedEntryFetch(p)
	entry[0] = 42
	Extra_MmFixedEntryRecycle(p, entry)
	if got := Extra_MmFixedEntryFetch(p); &got[0] != &entry[0] || got[0] != 42 {
		t.Fatal("recycle must reuse storage")
	}
	for i := int32(1); i <= p.nChunkSize; i++ {
		Extra_MmFixedEntryFetch(p)
	}
	if len(p.chunks) != 2 || Extra_MmFixedReadMaxEntriesUsed(p) != p.nChunkSize+1 {
		t.Fatal("growth statistics")
	}
	Extra_MmFixedRestart(p)
	if len(p.chunks) != 1 || p.nEntriesUsed != 0 || len(p.free) != int(p.nChunkSize) {
		t.Fatal("restart")
	}
	if Extra_MmFixedReadMemUsage(p) != p.nChunkSize*8 {
		t.Fatal("memory accounting")
	}
	Extra_MmFixedStop(p)
}
func TestFlexMemoryNonoverlapAndGrowth(t *testing.T) {
	p := Extra_MmFlexStart()
	a := Extra_MmFlexEntryFetch(p, 16)
	b := Extra_MmFlexEntryFetch(p, 16)
	a[0] = 7
	b[0] = 9
	if a[0] != 7 {
		t.Fatal("overlap")
	}
	c := Extra_MmFlexEntryFetch(p, 5000)
	if len(c) != 5000 || p.nChunkSize != 10000 || Extra_MmFlexReadMemUsage(p) != 14096 {
		t.Fatal("growth accounting")
	}
	if a[0] != 7 || b[0] != 9 {
		t.Fatal("old slices must survive growth")
	}
	Extra_MmFlexStop(p)
}
func TestStepMemoryClasses(t *testing.T) {
	p := Extra_MmStepStart(3)
	for _, size := range []int32{1, 4, 8, 9, 16, 17, 32} {
		entry := Extra_MmStepEntryFetch(p, size)
		if int32(len(entry)) != p.manager(size).nEntrySize {
			t.Fatal(size)
		}
		Extra_MmStepEntryRecycle(p, entry, size)
	}
	if Extra_MmStepEntryFetch(p, 0) != nil {
		t.Fatal("zero")
	}
	large := Extra_MmStepEntryFetch(p, 100)
	if len(large) != 100 || len(p.large) != 1 {
		t.Fatal("large")
	}
	usage := Extra_MmStepReadMemUsage(p)
	Extra_MmStepEntryRecycle(p, large, 100)
	if usage == 0 || Extra_MmStepReadMemUsage(p) != usage {
		t.Fatal("accounting")
	}
	Extra_MmStepStop(p)
}

func TestMemorySizeOverflowRejected(t *testing.T) {
	for _, run := range []func(){func() { Extra_MmFixedStart(1 << 30) }, func() { Extra_MmFlexEntryFetch(Extra_MmFlexStart(), 1<<30) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected size overflow panic")
				}
			}()
			run()
		}()
	}
}
