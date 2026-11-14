package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfCutTimingAndArea(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	p := &Lf_Man_t{pGia: g, pPars: &Jf_Par_t{NLutSize: 4}, pObjBests: make([]Lf_Bst_t, 1)}
	vec.Vec_IntFill(&p.vOffsets, int(g.nObjs), -1)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(x>>1), 0)
	vec.Vec_IntFill(&p.vCiArrivals, 2, 0)
	vec.Vec_IntFill(&p.vRequired, int(g.nObjs), util.ABC_INFINITY)
	vec.Vec_IntFill(&p.vMapRefs, 1, 0)
	Lf_ObjSetCiArrival(p, 0, 2)
	Lf_ObjSetCiArrival(p, 1, 4)
	vec.Vec_IntWriteEntry(&p.vRequired, int(a>>1), 7)
	c := &Lf_Cut_t{nLeaves: 2}
	c.pLeaves[0], c.pLeaves[1] = a>>1, b>>1
	if r := Lf_CutRequired(p, c); r != 9 {
		t.Fatalf("required %d", r)
	}
	Lf_CutParams(p, c, 4, 2, nil)
	if c.Delay != 5 || c.fLate != 1 || c.Flow != 0.5 {
		t.Fatalf("params %+v", c)
	}
	// A referenced internal best cut is counted once and restored after costing.
	pool := vec.Vec_PtrAlloc(0)
	Lf_MemAlloc(&p.vStoreNew, 7, pool, LF_CUT_WORDS)
	c.iFunc = -1
	handle := Lf_MemSaveCut(&p.vStoreNew, c, x>>1)
	p.pObjBests[0].Cut[0] = Lf_Plc_t{fUsed: 1, Handle: uint32(handle)}
	parent := &Lf_Cut_t{nLeaves: 2}
	parent.pLeaves[0], parent.pLeaves[1] = a>>1, x>>1
	if area := Lf_CutAreaDerefed(p, parent); area != 2 || Lf_ObjMapRefNum(p, x>>1) != 0 {
		t.Fatalf("area %v refs %d", area, Lf_ObjMapRefNum(p, x>>1))
	}
}
