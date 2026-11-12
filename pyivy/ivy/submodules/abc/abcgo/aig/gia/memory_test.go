package gia

import "testing"

func TestGiaFixedPoolChunkPolicy(t *testing.T) {
	p := Gia_MmFixedStart(12, 64)
	entries := make([][]byte, 9)
	for i := range entries {
		entries[i] = Gia_MmFixedEntryFetch(p)
		entries[i][0] = byte(i)
	}
	if Gia_MmFixedReadMemUsage(p) != 12*8*2 || Gia_MmFixedReadMaxEntriesUsed(p) != 9 {
		t.Fatal("GIA fixed chunk/stat policy")
	}
	Gia_MmFixedEntryRecycle(p, entries[0])
	reused := Gia_MmFixedEntryFetch(p)
	if &reused[0] != &entries[0][0] {
		t.Fatal("entry not recycled")
	}
	Gia_MmFixedRestart(p)
	if Gia_MmFixedReadMemUsage(p) != 12*8 {
		t.Fatal("restart did not retain one GIA chunk")
	}
	alias := entries[0]
	Gia_MmFixedStop(p, 0)
	alias[0] = 7
	if alias[0] != 7 {
		t.Fatal("release invalidated alias")
	}
}
func TestGiaFlexAndStepPools(t *testing.T) {
	flex := Gia_MmFlexStart()
	a := Gia_MmFlexEntryFetch(flex, 10)
	a[0] = 12
	b := Gia_MmFlexEntryFetch(flex, 1<<19)
	b[0] = 42
	if Gia_MmFlexReadMemUsage(flex) != 10+(1<<19) || a[0] != 12 {
		t.Fatal("flex accounting/stability")
	}
	Gia_MmFlexRestart(flex)
	if Gia_MmFlexReadMemUsage(flex) != 0 {
		t.Fatal("flex restart")
	}
	Gia_MmFlexStop(flex, 0)
	step := Gia_MmStepStart(3)
	small := Gia_MmStepEntryFetch(step, 7)
	if len(small) != 8 || cap(small) != 8 || Gia_MmStepReadMemUsage(step) != 8*1024 {
		t.Fatal("GIA step chunk policy")
	}
	Gia_MmStepEntryRecycle(step, small, 7)
	again := Gia_MmStepEntryFetch(step, 7)
	if &again[0] != &small[0] {
		t.Fatal("step recycling")
	}
	large := Gia_MmStepEntryFetch(step, 100)
	if len(large) != 100 || Gia_MmStepReadMemUsage(step) != 8*1024 {
		t.Fatal("large allocations must not alter pooled usage")
	}
	Gia_MmStepStop(step, 0)
}
