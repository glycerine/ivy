package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfEnumerateTwoLevel(t *testing.T) {
	for _, min := range []int32{0, 1} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, y)
		p := &Lf_Man_t{pGia: g, pPars: &Jf_Par_t{NLutSize: 3, NCutNum: 4, FCutMin: min, DelayTarget: -1}, pObjBests: make([]Lf_Bst_t, 2)}
		vec.Vec_IntFill(&p.vOffsets, int(g.nObjs), -1)
		vec.Vec_IntWriteEntry(&p.vOffsets, int(x>>1), 0)
		vec.Vec_IntWriteEntry(&p.vOffsets, int(y>>1), 1)
		vec.Vec_IntFill(&p.vCutSets, 2, -1)
		vec.Vec_IntFill(&p.vRequired, int(g.nObjs), util.ABC_INFINITY)
		vec.Vec_FltFill(&p.vFlowRefs, 2, 1)
		vec.Vec_IntFill(&p.vCiArrivals, 3, 0)
		vec.Vec_IntFill(&p.vMapRefs, 2, 0)
		if min != 0 {
			p.vTtMem = vec.Vec_MemAllocForTT(3, 0)
		}
		Lf_MemAlloc(&p.vStoreNew, 7, &p.vFreePages, LF_CUT_WORDS)
		Lf_ManSetCutRefs(p)
		Lf_ObjMergeOrder(p, x>>1)
		Lf_ObjMergeOrder(p, y>>1)
		best := Lf_ObjReadBest(p, y>>1)
		var cut Lf_Cut_t
		Lf_MemLoadCut(&p.vStoreNew, int32(best.Cut[0].Handle), y>>1, &cut, min, 0)
		if cut.nLeaves != 3 || cut.pLeaves[0] != a>>1 || cut.pLeaves[1] != b>>1 || cut.pLeaves[2] != c>>1 || best.Delay[0] != 1 {
			t.Fatalf("min=%d cut=%+v best=%+v", min, cut, best)
		}
		p.vStoreOld, p.vStoreNew = p.vStoreNew, p.vStoreOld
		Lf_MemAlloc(&p.vStoreNew, 7, &p.vFreePages, LF_CUT_WORDS)
		Lf_ManSetMapRefs(p)
		p.Iter = 1
		Lf_ManSetCutRefs(p)
		Lf_ObjMergeOrder(p, x>>1)
		Lf_ObjMergeOrder(p, y>>1)
		best = Lf_ObjReadBest(p, y>>1)
		Lf_MemLoadCut(&p.vStoreNew, int32(best.Cut[0].Handle), y>>1, &cut, min, 0)
		if cut.nLeaves != 3 || best.Delay[0] != 1 {
			t.Fatalf("second round min=%d cut=%+v", min, cut)
		}
	}
}

func TestLfEnumerateRealMux(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendMuxReal(g, c, a, b)
	Gia_ManAppendCo(g, x)
	p := &Lf_Man_t{pGia: g, pPars: &Jf_Par_t{NLutSize: 4, NCutNum: 4, FCutMin: 1}, pObjBests: make([]Lf_Bst_t, 1), vTtMem: vec.Vec_MemAllocForTT(4, 0)}
	vec.Vec_IntFill(&p.vOffsets, int(g.nObjs), -1)
	vec.Vec_IntWriteEntry(&p.vOffsets, int(x>>1), 0)
	vec.Vec_IntFill(&p.vCutSets, 1, -1)
	vec.Vec_IntFill(&p.vRequired, int(g.nObjs), util.ABC_INFINITY)
	vec.Vec_FltFill(&p.vFlowRefs, 1, 1)
	vec.Vec_IntFill(&p.vCiArrivals, 3, 0)
	Lf_MemAlloc(&p.vStoreNew, 7, &p.vFreePages, LF_CUT_WORDS)
	Lf_ObjMergeOrder(p, x>>1)
	best := Lf_ObjReadBest(p, x>>1)
	var cut Lf_Cut_t
	Lf_MemLoadCut(&p.vStoreNew, int32(best.Cut[0].Handle), x>>1, &cut, 1, 0)
	if cut.nLeaves != 3 || best.Delay[0] != 1 {
		t.Fatalf("MUX cut %+v", cut)
	}
	truth := Lf_CutTruth(p, &cut)[0]
	if cut.iFunc&1 != 0 {
		truth = ^truth
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := g.Evaluate(in)
		if err != nil || (truth>>uint(mask)&1 != 0) != out[0] {
			t.Fatalf("mask %d truth %x out %v error %v", mask, truth, out, err)
		}
	}
}
