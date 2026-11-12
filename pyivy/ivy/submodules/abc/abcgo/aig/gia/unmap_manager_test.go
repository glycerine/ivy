package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestUnmapUsedTruth(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	Gia_ManAppendCo(g, x)
	g.vMapping = vec.Vec_IntStart(int(g.nObjs))
	vec.Vec_IntWriteEntry(g.vMapping, int(x>>1), int32(g.nObjs))
	vec.Vec_IntPush(g.vMapping, 2)
	vec.Vec_IntPush(g.vMapping, a>>1)
	vec.Vec_IntPush(g.vMapping, b>>1)
	vec.Vec_IntPush(g.vMapping, x>>1)
	pairs := vec.Vec_IntAlloc(6)
	vec.Vec_IntPushThree(pairs, x>>1, a>>1, b>>1)
	vec.Vec_IntPushThree(pairs, x>>1, a>>1, b>>1)
	p := Unm_ManAlloc(g)
	p.vUsed = Unm_ManFindUsedNodes(pairs, g.nObjs)
	result := Unm_ManComputeTruths(p)
	if vec.Vec_IntSize(p.vUsed) != 1 || vec.Vec_WrdEntry(result, 0) != 0x2222222222222222 {
		t.Fatal("incorrect used node truth")
	}
	Unm_ManFree(p)
}

func TestUnmapRepeatedDecomposablePairs(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	g.vMapping = vec.Vec_IntStart(int(g.nObjs))
	for _, lut := range []struct {
		id   int32
		fans []int32
	}{
		{x >> 1, []int32{a >> 1, b >> 1}},
		{y >> 1, []int32{a >> 1, b >> 1, c >> 1}},
	} {
		vec.Vec_IntWriteEntry(g.vMapping, int(lut.id), int32(vec.Vec_IntSize(g.vMapping)))
		vec.Vec_IntPush(g.vMapping, int32(len(lut.fans)))
		for _, fan := range lut.fans {
			vec.Vec_IntPush(g.vMapping, fan)
		}
		vec.Vec_IntPush(g.vMapping, lut.id)
	}
	p := Unm_ManAlloc(g)
	defer Unm_ManFree(p)
	pairs := Unm_ManComputePairs(p, 0)
	if vec.Vec_IntSize(pairs) != 6 {
		t.Fatalf("expected two repeated pairs, got %v", vec.Vec_IntArray(pairs))
	}
	p.vUsed = Unm_ManFindUsedNodes(pairs, g.nObjs)
	p.vId2Used = vec.Vec_IntStartFull(int(g.nObjs))
	for i, id := range vec.Vec_IntArray(p.vUsed) {
		vec.Vec_IntWriteEntry(p.vId2Used, int(id), int32(i))
	}
	p.vTruths = Unm_ManComputeTruths(p)
	decomp := Unm_ManCollectDecomp(p, pairs, 0)
	if vec.Vec_IntSize(decomp) != 6 {
		t.Fatalf("expected two decomposable pairs, got %v", vec.Vec_IntArray(decomp))
	}
}
