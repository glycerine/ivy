package aig

import "testing"

func TestFixedPoolReuseAndOwnership(t *testing.T) {
	p := Aig_MmFixedStart(16, 2)
	a := Aig_MmFixedEntryFetch(p)
	a[0] = 7
	b := Aig_MmFixedEntryFetch(p)
	if Aig_MmFixedReadMemUsage(p) != 32 || Aig_MmFixedReadMaxEntriesUsed(p) != 2 {
		t.Fatal("incorrect accounting")
	}
	Aig_MmFixedEntryRecycle(p, a)
	c := Aig_MmFixedEntryFetch(p)
	if &c[0] != &a[0] || c[0] != 7 {
		t.Fatal("entry not reused")
	}
	Aig_MmFixedEntryRecycle(p, b)
	Aig_MmFixedEntryRecycle(p, c)
	defer func() {
		if recover() == nil {
			t.Fatal("double recycle accepted")
		}
	}()
	Aig_MmFixedEntryRecycle(p, c)
}
func TestFlexibleAndSteppedPools(t *testing.T) {
	f := Aig_MmFlexStart()
	Aig_MmFlexEntryFetch(f, 100)
	Aig_MmFlexEntryFetch(f, 9)
	if Aig_MmFlexReadMemUsage(f) != 109 {
		t.Fatal("flex accounting")
	}
	Aig_MmFlexRestart(f)
	if Aig_MmFlexReadMemUsage(f) != 0 {
		t.Fatal("flex restart")
	}
	p := Aig_MmStepStart(3)
	a, b := Aig_MmStepEntryFetch(p, 9), Aig_MmStepEntryFetch(p, 80)
	if len(a) != 9 || len(b) != 80 || Aig_MmStepReadMemUsage(p) != 96 {
		t.Fatal("step allocation")
	}
	Aig_MmStepEntryRecycle(p, a, 9)
	Aig_MmStepEntryRecycle(p, b, 80)
	if Aig_MmStepReadMemUsage(p) != 96 {
		t.Fatal("step accounting")
	}
}
