package mpm

import (
	"testing"

	"github.com/glycerine/abcgo/misc/mem"
)

func TestCutStoreCopiesAndReusesSourceHandles(t *testing.T) {
	p := &Mpm_Man_t{pManCuts: mem.Mmr_StepStart(6, 5), funcVar0: 3}
	defer mem.Mmr_StepStop(p.pManCuts)
	h := Mpm_CutCreateUnit(p, 7)
	original := Mpm_CutFetch(p, h)
	if original.nLeaves != 1 || original.pLeaves[0] != 14 || original.iFunc != 6 {
		t.Fatalf("unit cut = %+v", original)
	}
	original.fCompl, original.fUseless = 1, 1
	var copied *Mpm_Cut_t
	copyHandle := Mpm_CutCreate(p, original, &copied)
	dupHandle := Mpm_CutDup(p, original, 1)
	duplicate := Mpm_CutFetch(p, dupHandle)
	if copied.fCompl != 1 || copied.fUseless != 1 || duplicate.iFunc != 7 || duplicate.fCompl != 0 || duplicate.fUseless != 1 {
		t.Fatalf("copy=%+v duplicate=%+v", copied, duplicate)
	}
	original.pLeaves[0] = 18
	if copied.pLeaves[0] != 14 || duplicate.pLeaves[0] != 14 {
		t.Fatal("cut copies alias source leaves")
	}
	mem.Mmr_StepRecycle(p.pManCuts, copyHandle)
	reused := Mpm_CutCreateUnit(p, 4)
	if reused != copyHandle || Mpm_CutFetch(p, reused).pLeaves[0] != 8 {
		t.Fatalf("reused handle = %d, want %d with new contents", reused, copyHandle)
	}
}
