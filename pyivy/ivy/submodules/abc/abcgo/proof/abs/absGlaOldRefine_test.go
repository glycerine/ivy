package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestOldGlaRefinementReportsTrueCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	p := &Gla_Man_t{pGia: g, pPars: &Abs_Par_t{}, pObjs: make([]Gla_Obj_t, 3), nObjs: 3, pObj2Obj: make([]uint32, gia.Gia_ManObjNum(g)), vAbs: vec.Vec_IntAllocArrayCopy([]int32{2}, 1), pSat: bsat.NewLabeledSolver2()}
	p.pObjs[1] = Gla_Obj_t{id: 1, iGiaObj: x >> 1, fPi: 1}
	p.pObjs[2] = Gla_Obj_t{id: 2, fAnd: 1, fAbs: 1, nFanins: 1, Fanins: [4]int32{1}}
	p.pObj2Obj[x>>1] = 1
	p.pSat.NewVariable()
	v := p.pSat.NewVariable()
	vec.Vec_IntSetEntry(&p.pObjs[1].vFrames, 0, v)
	if p.pSat.Solve(bsat.Positive(v)) != bsat.Satisfiable {
		t.Fatal("SAT setup")
	}
	p.pRnm = Rnm_ManStart(g)
	defer Rnm_ManStop(p.pRnm, 0)
	if Gla_ManRefinement(p) != nil {
		t.Fatal("expected genuine counterexample")
	}
	if g.SequentialCounterexample() == nil {
		t.Fatal("missing concrete counterexample")
	}
}
