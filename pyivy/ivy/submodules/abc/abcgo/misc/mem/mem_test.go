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

func TestStepEntryCountAccessors(t *testing.T) {
	p := Mmr_StepStart(6, 3)
	if Mmr_StepEntryNum(p) != 0 || Mmr_StepEntryNumMax(p) != 0 {
		t.Fatal("new step manager counts are not zero")
	}
	a := Mmr_StepFetch(p, 2)
	b := Mmr_StepFetch(p, 3)
	if Mmr_StepEntryNum(p) != 2 || Mmr_StepEntryNumMax(p) != 2 {
		t.Fatal("fetch counts do not match live and high-water entries")
	}
	Mmr_StepRecycle(p, a)
	if Mmr_StepEntryNum(p) != 1 || Mmr_StepEntryNumMax(p) != 2 {
		t.Fatal("recycle did not decrement only the live-entry count")
	}
	Mmr_StepRecycle(p, b)
	if Mmr_StepEntryNum(p) != 0 || Mmr_StepEntryNumMax(p) != 2 {
		t.Fatal("recycling all entries changed the high-water count")
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

func TestArenaAppendCannotOverwriteNeighbor(t *testing.T) {
	p := Mem_FixedStart(8)
	a, b := Mem_FixedEntryFetch(p), Mem_FixedEntryFetch(p)
	b[0] = 42
	a = append(a, 9)
	if b[0] != 42 || len(a) != 9 {
		t.Fatal("fixed append overwrote neighbor")
	}
	f := Mem_FlexStart()
	a = Mem_FlexEntryFetch(f, 8)
	b = Mem_FlexEntryFetch(f, 8)
	b[0] = 42
	a = append(a, 9)
	if b[0] != 42 {
		t.Fatal("flex append overwrote neighbor")
	}
	m := Mmr_FixedStart(5, 3)
	h := Mmr_FixedFetch(m)
	next := Mmr_FixedFetch(m)
	words := Mmr_FixedEntry(m, h)
	neighbor := Mmr_FixedEntry(m, next)
	neighbor[0] = 42
	words = append(words, 9)
	if neighbor[0] != 42 {
		t.Fatal("word append overwrote neighbor")
	}
}
func TestRestartDropsChunkReferences(t *testing.T) {
	f := Mem_FlexStart()
	Mem_FlexEntryFetch(f, 4096)
	Mem_FlexEntryFetch(f, 8192)
	backing := f.chunks[:cap(f.chunks)]
	Mem_FlexRestart(f)
	for _, chunk := range backing[1:] {
		if chunk != nil {
			t.Fatal("flex retained released chunk")
		}
	}
	p := Mem_FixedStart(8)
	for i := 0; i < p.chunkSize+1; i++ {
		Mem_FixedEntryFetch(p)
	}
	backing = p.chunks[:cap(p.chunks)]
	Mem_FixedRestart(p)
	for _, chunk := range backing[1:] {
		if chunk != nil {
			t.Fatal("fixed retained released chunk")
		}
	}
}

func TestMaximumSizeHandleEntry(t *testing.T) {
	p := Mmr_FixedStart(5, 31)
	h := Mmr_FixedFetch(p)
	entry := Mmr_FixedEntry(p, h)
	if len(entry) != 31 {
		t.Fatal(len(entry))
	}
	entry[30] = 7
	h2 := Mmr_FixedFetch(p)
	if h2>>5 == h>>5 {
		t.Fatal("maximum entries must occupy separate pages")
	}
	if Mmr_FixedEntry(p, h)[30] != 7 {
		t.Fatal("entry moved")
	}
	Mmr_FixedRecycle(p, h)
	if Mmr_FixedFetch(p) != h {
		t.Fatal("maximum entry recycling")
	}
}

func TestFixedCustomChunkPolicy(t *testing.T) {
	p := Mem_FixedStartWithChunk(8, 3)
	for i := 0; i < 4; i++ {
		Mem_FixedEntryFetch(p)
	}
	if Mem_FixedReadMemUsage(p) != 48 {
		t.Fatal("custom growth")
	}
	Mem_FixedRestart(p)
	if Mem_FixedReadMemUsage(p) != 24 {
		t.Fatal("custom restart")
	}
}

func TestCustomFlexAndStepPolicies(t *testing.T) {
	f := Mem_FlexStartWithChunk(1 << 18)
	Mem_FlexEntryFetch(f, 10)
	if len(f.chunks[0]) != 1<<18 {
		t.Fatal("flex chunk policy")
	}
	p := Mem_StepStartWithChunk(4, 1024)
	Mem_StepEntryFetch(p, 40)
	if Mem_StepReadMemUsage(p) != 64*1024 {
		t.Fatal("step chunk policy")
	}
}

func TestStepRestart(t *testing.T) {
	p := Mem_StepStartWithChunk(2, 3)
	first := Mem_StepEntryFetch(p, 5)
	for i := 0; i < 4; i++ {
		Mem_StepEntryFetch(p, 5)
	}
	Mem_StepEntryFetch(p, 100)
	Mem_StepRestart(p)
	if len(p.large) != 0 || Mem_StepReadMemUsage(p) != 24 {
		t.Fatal("restart storage")
	}
	next := Mem_StepEntryFetch(p, 5)
	if &next[0] != &first[0] {
		t.Fatal("restart did not reuse first chunk")
	}
}
