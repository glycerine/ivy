package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"reflect"
	"testing"
)

func TestOldGlaRefinement2SelectsPseudoInput(t *testing.T) {
	for _, prop := range []int32{0, 1} {
		g := gia.Gia_ManStart(12)
		x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		z := gia.Gia_ManAppendAnd(g, x, y)
		root := gia.Gia_ManAppendAnd(g, x, z)
		gia.Gia_ManAppendCo(g, root)
		p := &Gla_Man_t{pGia: g, pPars: &Abs_Par_t{FPropFanout: prop}, pObjs: make([]Gla_Obj_t, 4), nObjs: 4, pObj2Obj: make([]uint32, gia.Gia_ManObjNum(g)), pvRefis: make([][]Rfn_Obj_t, gia.Gia_ManObjNum(g)), vObjCounts: vec.Vec_IntAlloc(0), vAbs: vec.Vec_IntAllocArrayCopy([]int32{3}, 1), pSat: bsat.NewLabeledSolver2()}
		for i := range p.pObj2Obj {
			p.pObj2Obj[i] = ^uint32(0)
		}
		p.pObjs[1] = Gla_Obj_t{id: 1, iGiaObj: x >> 1, fPi: 1}
		p.pObjs[2] = Gla_Obj_t{id: 2, iGiaObj: z >> 1, fAnd: 1}
		p.pObjs[3] = Gla_Obj_t{id: 3, iGiaObj: root >> 1, fAnd: 1, fAbs: 1, nFanins: 2, Fanins: [4]int32{1, 2}}
		p.pObj2Obj[x>>1] = 1
		p.pObj2Obj[z>>1] = 2
		p.pObj2Obj[root>>1] = 3
		p.pSat.NewVariable()
		vx, vz := p.pSat.NewVariable(), p.pSat.NewVariable()
		vec.Vec_IntSetEntry(&p.pObjs[1].vFrames, 0, vx)
		vec.Vec_IntSetEntry(&p.pObjs[2].vFrames, 0, vz)
		if p.pSat.Solve(bsat.Positive(vx), bsat.Positive(vz)) != bsat.Satisfiable {
			t.Fatal("SAT setup")
		}
		if prop != 0 {
			gia.Gia_ManStaticFanoutStart(g)
		}
		for run := 0; run < 2; run++ {
			if got := vec.Vec_IntArray(Gla_ManRefinement2(p)); !reflect.DeepEqual(got, []int32{2}) {
				t.Fatalf("prop=%d run=%d selection %v", prop, run, got)
			}
		}
	}
}
