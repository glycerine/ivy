package bsat

import "testing"

func TestSATMemoryPoolAdapters(t *testing.T) {
	fixed := Sat_MmFixedStart(24)
	a := Sat_MmFixedEntryFetch(fixed)
	a[0] = 91
	Sat_MmFixedEntryRecycle(fixed, a)
	b := Sat_MmFixedEntryFetch(fixed)
	if &a[0] != &b[0] {
		t.Fatal("fixed recycle")
	}
	Sat_MmFixedRestart(fixed)
	if Sat_MmFixedReadMemUsage(fixed) <= 0 {
		t.Fatal("restart lost first chunk")
	}
	Sat_MmFixedStop(fixed, 0)
	flex := Sat_MmFlexStart()
	x := Sat_MmFlexEntryFetch(flex, 17)
	x[0] = 43
	y := Sat_MmFlexEntryFetch(flex, 5000)
	if len(y) != 5000 || x[0] != 43 {
		t.Fatal("flex ownership")
	}
	Sat_MmFlexStop(flex, 0)
	step := Sat_MmStepStart(4)
	p := Sat_MmStepEntryFetch(step, 9)
	if len(p) < 9 {
		t.Fatal("step bucket")
	}
	Sat_MmStepEntryRecycle(step, p, 9)
	q := Sat_MmStepEntryFetch(step, 9)
	if &p[0] != &q[0] {
		t.Fatal("step recycle")
	}
	Sat_MmStepEntryFetch(step, 500)
	Sat_MmStepRestart(step)
	r := Sat_MmStepEntryFetch(step, 9)
	if &p[0] != &r[0] {
		t.Fatal("step restart contract")
	}
	Sat_MmStepStop(step, 0)
}
