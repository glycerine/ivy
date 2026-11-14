package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfArrivalCostAndBest(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, x^1)
	p := &Lf_Man_t{pGia: g, pPars: &Jf_Par_t{NLutSize: 4, FCutMin: 1, FOptEdge: 1, NAreaTuner: 1}, pObjBests: make([]Lf_Bst_t, 1)}
	vec.Vec_IntFill(&p.vOffsets, int(g.nObjs), -1)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(x>>1), 0)
	vec.Vec_IntFill(&p.vCiArrivals, 2, 0)
	Lf_ObjSetCiArrival(p, 0, 3)
	best := Lf_ObjReadBest(p, x>>1)
	best.Delay[0] = 6
	best.Flow[0] = 1.5
	best.Cut[0].fUsed = 1
	pool := vec.Vec_PtrAlloc(0)
	Lf_MemAlloc(&p.vStoreOld, 7, pool, LF_CUT_WORDS)
	c := Lf_Cut_t{nLeaves: 2, iFunc: 4}
	c.pLeaves[0], c.pLeaves[1] = a>>1, b>>1
	best.Cut[0].Handle = uint32(Lf_MemSaveCut(&p.vStoreOld, &c, x>>1))
	got := Lf_ObjCutBest(p, x>>1)
	if got.Delay != 6 || got.Flow != 1.5 || got.pLeaves[1] != b>>1 || Lf_ObjCoArrival(p, 0) != 6 || Lf_ObjCoArrival2(p, 0) != 6 {
		t.Fatal("best or arrival", got)
	}
	if Lf_CutArea(p, got) != 3 {
		t.Fatal("cut cost")
	}
	var drivers, inverts int32
	Lf_ManAnalyzeCoDrivers(g, &drivers, &inverts)
	if drivers != 1 || inverts != 1 {
		t.Fatal("drivers", drivers, inverts)
	}
}
