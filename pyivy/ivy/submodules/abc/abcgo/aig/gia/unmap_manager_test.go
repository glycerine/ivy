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
