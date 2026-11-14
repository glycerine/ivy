package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestJfMappingHelpers(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	p := Jf_ManAlloc(g, &Jf_Par_t{NLutSize: 4, NCutNum: 4})
	defer Jf_ManFree(p)
	Jf_ObjAssignCut(p, Gia_ManObj(g, a>>1))
	Jf_ObjAssignCut(p, Gia_ManObj(g, b>>1))
	cut := []int32{Jf_CutSetAll(2, 1, 2), a, b}
	vec.Vec_IntWriteEntry(&p.vArr, int(a>>1), 3)
	vec.Vec_IntWriteEntry(&p.vArr, int(b>>1), 5)
	vec.Vec_FltWriteEntry(&p.vFlow, int(a>>1), 1.25)
	vec.Vec_FltWriteEntry(&p.vFlow, int(b>>1), 2.5)
	if Jf_CutArr(p, cut) != 6 || Jf_CutFlow(p, cut) != 3.75 {
		t.Fatal("arrival or flow")
	}
	oldA, oldB := g.Refs[a>>1], g.Refs[b>>1]
	Jf_CutRef(p, cut)
	if g.Refs[a>>1] != oldA+1 || g.Refs[b>>1] != oldB+1 {
		t.Fatal("ref")
	}
	Jf_CutDeref(p, cut)
	if g.Refs[a>>1] != oldA || g.Refs[b>>1] != oldB {
		t.Fatal("deref")
	}
	if Jf_CutArea(p, cut, 0) != 1 || g.Refs[a>>1] != oldA || g.Refs[b>>1] != oldB {
		t.Fatal("area or restoration")
	}
	if Jf_CutCheckMffc(p, cut, 3) != 1 || Jf_CutCheckMffc(p, cut, 2) != 0 {
		t.Fatal("MFFC limit")
	}
}
