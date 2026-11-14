package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfCutSetReuse(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, a)
	p := &Lf_Man_t{pGia: g, pPars: &Jf_Par_t{NCutNum: 4}}
	vec.Vec_IntFill(&p.vOffsets, int(g.nObjs), -1)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(x>>1), 0)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(y>>1), 1)
	vec.Vec_IntFill(&p.vCutSets, 2, -1)
	xo := Gia_ManObj(g, x>>1)
	xo.Value = 2
	set := Lf_ManFetchSet(p, x>>1)
	set[0].nLeaves = 2
	var out []Lf_Cut_t
	if Lf_ManPrepareSet(p, x>>1, 0, &out) != 1 || xo.Value != 1 {
		t.Fatal("first visit")
	}
	if Lf_ManPrepareSet(p, x>>1, 0, &out) != 1 || xo.Value != 0 || vec.Vec_IntEntry(&p.vCutSets, 0) != -1 {
		t.Fatal("final visit")
	}
	yo := Gia_ManObj(g, y>>1)
	yo.Value = 1
	next := Lf_ManFetchSet(p, y>>1)
	if &next[0] != &set[0] {
		t.Fatal("set not reused")
	}
	if Lf_ManPrepareSet(p, a>>1, 1, &out) != 1 || out[0].pLeaves[0] != a>>1 {
		t.Fatal("unit input")
	}
}
