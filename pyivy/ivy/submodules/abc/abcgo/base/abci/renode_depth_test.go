package abci

import (
	"testing"

	ifpkg "github.com/glycerine/abcgo/map/if"
)

func TestRenodeAigCostForZeroLeafCut(t *testing.T) {
	p := ifpkg.If_ManStart(&ifpkg.If_Par_t{NLutSize: 6, NCutsMax: 8, FTruth: 1})
	defer ifpkg.If_ManStop(p)
	cut := new(ifpkg.If_Cut_t)
	ifpkg.If_CutSetup(p, cut)
	renodeState.Lock()
	defer renodeState.Unlock()
	if got := renodeCostAig(p, cut); got != 0 || len(ifpkg.If_CutPerm(cut)) != 0 {
		t.Fatalf("constant cut has cost %d and pin delays %v", got, ifpkg.If_CutPerm(cut))
	}
}

func TestRenodeConstantWithUnusedCutLeaves(t *testing.T) {
	p := ifpkg.If_ManStart(&ifpkg.If_Par_t{NLutSize: 6, NCutsMax: 8, FTruth: 1})
	defer ifpkg.If_ManStop(p)
	var a, b, cut ifpkg.If_Cut_t
	ifpkg.If_ManSetupCutTriv(p, &a, ifpkg.If_ManCreateCi(p).Id)
	ifpkg.If_ManSetupCutTriv(p, &b, ifpkg.If_ManCreateCi(p).Id)
	ifpkg.If_CutSetup(p, &cut)
	if ifpkg.If_CutMerge(p, &a, &b, &cut) == 0 || ifpkg.If_CutLeaveNum(&cut) != 2 || *ifpkg.If_CutTruth(p, &cut) != 0 {
		t.Fatal("expected a two-leaf cut with constant truth")
	}
	renodeState.Lock()
	defer renodeState.Unlock()
	if got := renodeCostAig(p, &cut); got != 0 {
		t.Fatalf("constant cut cost %d", got)
	}
	for _, delay := range ifpkg.If_CutPerm(&cut) {
		if delay != 156 {
			t.Fatalf("unused pin delay %d, want byte representation of -100", delay)
		}
	}
}
