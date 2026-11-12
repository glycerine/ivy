package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSleCutMergeAndContainment(t *testing.T) {
	a := []int32{Sle_CutSetSizeSign(2, (1<<1)|(1<<3)), 1, 3}
	b := []int32{Sle_CutSetSizeSign(2, (1<<3)|(1<<5)), 3, 5}
	out := make([]int32, 5)
	if Sle_CutMergeOrder(a, b, out, 3) != 1 || Sle_CutSize(out) != 3 || out[1] != 1 || out[2] != 3 || out[3] != 5 {
		t.Fatalf("merge %v", out)
	}
	if Sle_CutMergeOrder(a, b, out, 2) != 0 {
		t.Fatal("oversized cut accepted")
	}
	if Sle_SetCutIsContainedOrder(out, a) != 1 || Sle_SetCutIsContainedOrder(a, b) != 0 {
		t.Fatal("bad cut containment")
	}
	set := vec.Vec_IntStart(1)
	Sle_SetAddCut(set, out)
	if Sle_SetLastCutIsContained(set, a) != 0 {
		t.Fatal("superset contains candidate")
	}
	Sle_SetAddCut(set, a)
	if Sle_CutIsUsed(vec.Vec_IntArray(set)[1:]) != 0 || Sle_SetLastCutIsContained(set, out) != 1 {
		t.Fatal("dominance filtering failed")
	}
}

func TestSleComputeCuts(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	cuts := Sle_ManComputeCuts(p, 3, 0)
	list := sleList(cuts, y>>1, false)
	found := false
	for _, cut := range list {
		leaves := Sle_CutLeaves(cut)
		if len(leaves) == 3 && leaves[0] == a>>1 && leaves[1] == b>>1 && leaves[2] == c>>1 {
			found = true
		}
	}
	if !found || Sle_ManComputeDelay(p, cuts) != 1 {
		t.Fatalf("cuts %v delay %d", list, Sle_ManComputeDelay(p, cuts))
	}
	if got := Sle_ManComputeDelay(p, Sle_ManComputeCuts(p, 2, 0)); got != 2 {
		t.Fatalf("2-LUT delay %d", got)
	}
}

func TestSleCutFanins(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	cuts := Sle_ManComputeCuts(p, 3, 0)
	mask := Sle_ManInternalNodeMask(p)
	fanins := Sle_ManCollectCutFanins(p, cuts, mask)
	got := vec.Vec_IntArray(vec.Vec_WecEntry(fanins, int(y>>1)))
	if len(got) != 1 || got[0] != x>>1 {
		t.Fatalf("cut fanins %v", got)
	}
	var piOnly, internal bool
	for _, cut := range sleList(cuts, y>>1, false) {
		if Sle_ManCutHasPisOnly(cut, mask) != 0 {
			piOnly = true
		} else {
			internal = true
		}
	}
	if !piOnly || !internal {
		t.Fatal("expected both primary-input and internal cuts")
	}
}

func TestSleManagerVariables(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	m := Sle_ManAlloc(p, 0, 0)
	Sle_ManMarkupVariables(m)
	if m.nLevels != 1 || m.nNodeVars != p.nObjs || m.nCutVars != 3 || m.nEdgeVars != 1 || m.nDelayVars != 2 || m.nVarsTotal != p.nObjs+6 {
		t.Fatalf("bad variable counts: %+v", m)
	}
	if Sle_ManCheckContained([]int32{1, 3}, 2, []int32{3, 2, 1}, 3) != 1 || Sle_ManCheckContained([]int32{1, 4}, 2, []int32{3, 2, 1}, 3) != 0 {
		t.Fatal("bad LUT containment")
	}
	Sle_ManStop(m)
}
