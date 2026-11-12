package gia

import "testing"

func TestSupportOverlapDistanceAndTemporaryFanoutOwnership(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, c, d)
	root := Gia_ManAppendAnd(p, x, y)
	Gia_ManAppendCo(p, root)
	if Gia_ManCheckSuppOverlap(p, x>>1, y>>1) != 0 || Gia_ManCheckSuppOverlap(p, x>>1, root>>1) != 1 {
		t.Fatal("cone overlap")
	}
	Gia_ManCheckMark0(p)
	distances := Gia_ManComputeDistance(p, x>>1, nil, 0)
	if Vec_IntEntry(distances, int(x>>1)) != 1 || Vec_IntEntry(distances, int(root>>1)) != 2 || Vec_IntEntry(distances, int(y>>1)) != 3 || Vec_IntEntry(distances, int(c>>1)) != 4 {
		t.Fatal("ring distance", Vec_IntArray(distances))
	}
	if p.vFanoutNums != nil || p.vFanout != nil {
		t.Fatal("temporary fanout ownership leaked")
	}
	Gia_ManStaticFanoutStart(p)
	fanouts := p.vFanout
	Gia_ManComputeDistance(p, x>>1, nil, 0)
	if p.vFanout != fanouts {
		t.Fatal("borrowed static fanouts replaced")
	}
}
func TestValueSnapshotCopyRemapAndOutputSwap(t *testing.T) {
	p := Gia_ManStart(5)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b^1)
	p.pObjs[1].Value = 14
	p.pObjs[2].Value = ^uint32(0)
	saved := Gia_ManSaveValue(p)
	Gia_ManCleanValue(p)
	Gia_ManLoadValue(p, saved)
	copyMap := Vec_IntAlloc(3)
	for _, lit := range []int32{a ^ 1, b, -1} {
		Vec_IntPush(copyMap, lit)
	}
	Gia_ManUpdateCopy(copyMap, p)
	assertIDs(t, copyMap, 15, -1, -1)
	Gia_ManSwapPos(p, 1)
	out, err := p.Evaluate([]bool{true, true})
	if err != nil || out[0] || !out[1] {
		t.Fatal("PO swap", out, err)
	}
	if Gia_ManCountPisWithFanout(p) != 2 || Gia_ManCountPosWithNonZeroDrivers(p) != 2 {
		t.Fatal("interface fanout counts")
	}
	first := Gia_ManFirstFanouts(p)
	if Vec_IntEntry(first, int(a>>1)) != Gia_ObjId(p, Gia_ManCo(p, 1)) || Vec_IntEntry(first, int(b>>1)) != Gia_ObjId(p, Gia_ManCo(p, 0)) {
		t.Fatal("first fanouts")
	}
}
