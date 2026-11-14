package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMfMappingRefs(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	refs := vec.Vec_IntAlloc(0)
	Mf_ManSetFlowRefs(g, refs)
	if vec.Vec_IntEntry(refs, int(x>>1)) != 1 {
		t.Fatal("missing output reference")
	}
	p := &Mf_Man_t{pGia: g, pPars: &Jf_Par_t{NLutSize: 2, DelayTarget: -1}, pLfObjs: make([]Mf_Obj_t, g.nObjs), iCur: 2}
	cut := &Mf_Cut_t{nLeaves: 2}
	cut.pLeaves[0], cut.pLeaves[1] = a>>1, b>>1
	p.pLfObjs[x>>1].iCutSet = Mf_ManSaveCuts(p, []*Mf_Cut_t{cut}, 1)
	p.pLfObjs[x>>1].Delay = 2
	if got := Mf_ManSetMapRefs(p); got != 1 || p.pPars.Edge != 2 || p.pPars.Delay != 2 {
		t.Fatalf("area %d edges %d delay %d", got, p.pPars.Edge, p.pPars.Delay)
	}
	if Mf_ObjMapRefNum(p, x>>1) != 1 || Mf_ObjMapRefNum(p, a>>1) != 1 {
		t.Fatal("incorrect mapping references")
	}
}
