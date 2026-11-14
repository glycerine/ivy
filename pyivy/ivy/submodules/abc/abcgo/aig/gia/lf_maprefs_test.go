package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfMappingReferencePropagation(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	pars := &Jf_Par_t{NLutSize: 4, DelayTarget: -1, FCutMin: 1}
	p := &Lf_Man_t{pGia: g, pPars: pars, pObjBests: make([]Lf_Bst_t, 1)}
	vec.Vec_IntFill(&p.vOffsets, int(g.nObjs), -1)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(x>>1), 0)
	vec.Vec_IntFill(&p.vCiArrivals, 2, 0)
	vec.Vec_FltFill(&p.vFlowRefs, 1, 1)
	pool := vec.Vec_PtrAlloc(0)
	Lf_MemAlloc(&p.vStoreOld, 7, pool, LF_CUT_WORDS)
	cut := Lf_Cut_t{nLeaves: 2, iFunc: 4}
	cut.pLeaves[0], cut.pLeaves[1] = a>>1, b>>1
	handle := Lf_MemSaveCut(&p.vStoreOld, &cut, x>>1)
	p.pObjBests[0].Cut[0].Handle = uint32(handle)
	p.pObjBests[0].Cut[1].Handle = uint32(handle)
	p.pObjBests[0].Delay[0] = 1
	p.pObjBests[0].Delay[1] = 1
	if Lf_ManSetMapRefs(p) != 1 || pars.Edge != 2 || pars.Delay != 1 || Lf_ObjMapRefNum(p, x>>1) != 1 || Lf_ObjRequired(p, a>>1) != 0 {
		t.Fatalf("area=%d edge=%d delay=%d", pars.Area, pars.Edge, pars.Delay)
	}
	Lf_ManCountMapRefs(p)
	if pars.Area != 1 || pars.Edge != 2 || pars.Delay != 1 {
		t.Fatal("recount", pars)
	}
}
