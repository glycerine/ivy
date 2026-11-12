package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfInitialReferences(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	z := Gia_ManAppendAnd(g, x, b)
	Gia_ManAppendCo(g, y)
	Gia_ManAppendCo(g, z)
	p := &Lf_Man_t{pGia: g}
	vec.Vec_IntFill(&p.vOffsets, int(g.nObjs), -1)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(x>>1), 0)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(y>>1), 1)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(z>>1), 2)
	Lf_ManSetFlowRefs(g, &p.vFlowRefs, &p.vOffsets)
	refs := vec.Vec_FltArray(&p.vFlowRefs)
	if refs[0] != 2 || refs[1] != 1 || refs[2] != 1 {
		t.Fatal("flow refs", refs)
	}
	Lf_ManSetCutRefs(p)
	if Gia_ManObj(g, x>>1).Value != 2 || Gia_ManObj(g, y>>1).Value != 0 {
		t.Fatal("cut refs")
	}
}
