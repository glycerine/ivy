package mem

import "testing"

func TestFixedReuseAndRestart(t *testing.T) {
	p := Mem_FixedStart(16)
	var entries [][]byte
	for i := 0; i < 1100; i++ {
		b := Mem_FixedEntryFetch(p)
		b[0] = byte(i)
		entries = append(entries, b)
	}
	if entries[12][0] != 12 || Mem_FixedReadMaxEntriesUsed(p) != 1100 {
		t.Fatal("unstable arena")
	}
	b := entries[10]
	Mem_FixedEntryRecycle(p, b)
	if &Mem_FixedEntryFetch(p)[0] != &b[0] {
		t.Fatal("not reused")
	}
	Mem_FixedRestart(p)
	if Mem_FixedReadMemUsage(p) != 16384 {
		t.Fatal("restart did not release chunks")
	}
}
func TestFlexGrowthAndRestart(t *testing.T) {
	p := Mem_FlexStart()
	a := Mem_FlexEntryFetch(p, 10)
	a[0] = 7
	Mem_FlexEntryFetch(p, 10000)
	if a[0] != 7 || Mem_FlexReadMemUsage(p) != 10010 {
		t.Fatal("growth")
	}
	Mem_FlexRestart(p)
	if Mem_FlexReadMemUsage(p) != 0 {
		t.Fatal("restart")
	}
	if len(Mem_FlexEntryFetch(p, 6000)) != 6000 {
		t.Fatal("large restart allocation")
	}
}
func TestStepClasses(t *testing.T) {
	p := Mem_StepStart(3)
	a := Mem_StepEntryFetch(p, 9)
	if len(a) != 16 {
		t.Fatal(len(a))
	}
	Mem_StepEntryRecycle(p, a, 9)
	b := Mem_StepEntryFetch(p, 9)
	if &a[0] != &b[0] {
		t.Fatal("reuse")
	}
	if len(Mem_StepEntryFetch(p, 100)) != 100 {
		t.Fatal("large")
	}
}
func TestHandleArenas(t *testing.T) {
	p := Mmr_FixedStart(5, 3)
	var hs []int32
	for i := 0; i < 40; i++ {
		h := Mmr_FixedFetch(p)
		Mmr_FixedEntry(p, h)[0] = uint64(i)
		hs = append(hs, h)
	}
	for i, h := range hs {
		if Mmr_FixedEntry(p, h)[0] != uint64(i) {
			t.Fatal("handle moved")
		}
	}
	Mmr_FixedRecycle(p, hs[3])
	h := Mmr_FixedFetch(p)
	if h != hs[3] || Mmr_FixedEntry(p, h)[0] != ^uint64(0) {
		t.Fatal("recycling")
	}
	s := Mmr_StepStart(6, 3)
	h = Mmr_StepFetch(s, 5)
	if len(Mmr_StepEntry(s, h)) != 5 {
		t.Fatal("encoded class")
	}
	Mmr_StepRecycle(s, h)
	if Mmr_StepFetch(s, 5) != h {
		t.Fatal("step recycle")
	}
}
func TestFlexHandlesAcrossPages(t *testing.T) {
	p := Mmr_FlexStart(4)
	h := Mmr_FlexFetch(p, 8)
	Mmr_FlexEntry(p, h)[0] = 9
	next := Mmr_FlexFetch(p, 8)
	if next>>4 == h>>4 || Mmr_FlexEntry(p, h)[0] != 9 {
		t.Fatal("page rollover")
	}
	Mmr_FlexRelease(p, next)
	if p.pages[0] != nil {
		t.Fatal("page release")
	}
}
