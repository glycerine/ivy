package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSplitMappingRoundTrip(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.vMapping, int(x>>1), int32(p.nObjs))
	vec.Vec_IntPush(p.vMapping, 2)
	vec.Vec_IntPush(p.vMapping, a>>1)
	vec.Vec_IntPush(p.vMapping, b>>1)
	vec.Vec_IntPush(p.vMapping, x>>1)
	p.vMapping2 = Spl_ManToWecMapping(p)
	p.vMapping = Spl_ManFromWecMapping(p, p.vMapping2)
	got := Gia_ObjLutFanins(p, x>>1)
	if len(got) != 2 || got[0] != a>>1 || got[1] != b>>1 {
		t.Fatalf("round trip %v", got)
	}
}

func TestSplitManagerMappingLifecycle(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.vMapping, int(x>>1), int32(p.nObjs))
	vec.Vec_IntPush(p.vMapping, 2)
	vec.Vec_IntPush(p.vMapping, a>>1)
	vec.Vec_IntPush(p.vMapping, b>>1)
	vec.Vec_IntPush(p.vMapping, x>>1)
	m := Spl_ManAlloc(p, 8, 0)
	if p.vMapping != nil || p.vMapping2 == nil || p.vFanout == nil {
		t.Fatal("window setup failed")
	}
	Spl_ManStop(m)
	if p.vMapping == nil || p.vMapping2 != nil || p.vFanout != nil || Gia_ObjLutSize(p, x>>1) != 2 {
		t.Fatal("window cleanup failed")
	}
}

func TestSplitWindowBoundary(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id int32
		in []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		vec.Vec_IntWriteEntry(p.vMapping, int(lut.id), int32(vec.Vec_IntSize(p.vMapping)))
		vec.Vec_IntPush(p.vMapping, int32(len(lut.in)))
		for _, id := range lut.in {
			vec.Vec_IntPush(p.vMapping, id)
		}
		vec.Vec_IntPush(p.vMapping, lut.id)
	}
	m := Spl_ManAlloc(p, 8, 0)
	defer Spl_ManStop(m)
	vec.Vec_IntPush(m.vNodes, x>>1)
	vec.Vec_IntPush(m.vAnds, x>>1)
	vec.Vec_BitWriteEntry(m.vMarksAnd, int(x>>1), 1)
	Spl_ManWinFindLeavesRoots(m)
	if vec.Vec_IntSize(m.vLeaves) != 2 || vec.Vec_IntSize(m.vRoots) != 1 || vec.Vec_IntEntry(m.vRoots, 0) != x>>1 {
		t.Fatalf("leaves %v roots %v", vec.Vec_IntArray(m.vLeaves), vec.Vec_IntArray(m.vRoots))
	}
	if Spl_ManLutFanouts(p, x>>1, m.vFanouts, m.vMarksNo, m.vMarksCIO) != 1 || vec.Vec_IntEntry(m.vFanouts, 0) != y>>1 {
		t.Fatal("bad LUT fanouts")
	}
}

func TestSplitComputePivotWindow(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id int32
		in []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		vec.Vec_IntWriteEntry(p.vMapping, int(lut.id), int32(vec.Vec_IntSize(p.vMapping)))
		vec.Vec_IntPush(p.vMapping, int32(len(lut.in)))
		for _, id := range lut.in {
			vec.Vec_IntPush(p.vMapping, id)
		}
		vec.Vec_IntPush(p.vMapping, lut.id)
	}
	Gia_ManComputeOneWinStart(p, 8, 0)
	var roots, nodes, leaves, ands *Vec_Int_t
	if n := Gia_ManComputeOneWin(p, y>>1, &roots, &nodes, &leaves, &ands); n != 2 || vec.Vec_IntSize(leaves) != 3 || vec.Vec_IntSize(roots) != 1 || vec.Vec_IntEntry(roots, 0) != y>>1 {
		t.Fatalf("window n=%d roots=%v leaves=%v", n, vec.Vec_IntArray(roots), vec.Vec_IntArray(leaves))
	}
	Gia_ManComputeOneWin(p, -1, nil, nil, nil, nil)
	if p.vMapping == nil || p.pSatlutWinman != nil {
		t.Fatal("window manager not closed")
	}
}
