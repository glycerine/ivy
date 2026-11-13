package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"reflect"
	"testing"
)

func TestOldGlaCexMapsBoundaryAcrossFrames(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	p := &Gla_Man_t{pGia: g, pPars: &Abs_Par_t{iFrame: 1}, pObjs: make([]Gla_Obj_t, 3), nObjs: 3, pObj2Obj: make([]uint32, gia.Gia_ManObjNum(g)), vAbs: vec.Vec_IntAllocArrayCopy([]int32{2}, 1), pSat: bsat.NewLabeledSolver2()}
	p.pObjs[1] = Gla_Obj_t{id: 1, iGiaObj: x >> 1, fPi: 1}
	p.pObjs[2] = Gla_Obj_t{id: 2, fAnd: 1, fAbs: 1, nFanins: 1, Fanins: [4]int32{1}}
	p.pObj2Obj[x>>1] = 1
	p.pSat.NewVariable()
	v0, v1 := p.pSat.NewVariable(), p.pSat.NewVariable()
	vec.Vec_IntSetEntry(&p.pObjs[1].vFrames, 0, v0)
	vec.Vec_IntSetEntry(&p.pObjs[1].vFrames, 1, v1)
	if p.pSat.Solve(bsat.Positive(v0), bsat.Negative(v1)) != bsat.Satisfiable {
		t.Fatal("SAT setup")
	}
	var cex *util.Abc_Cex_t
	var mapping *vec.Vec_Int_t
	Gia_GlaPrepareCexAndMap(p, &cex, &mapping)
	if got := vec.Vec_IntArray(mapping); !reflect.DeepEqual(got, []int32{x >> 1}) {
		t.Fatalf("mapping %v", got)
	}
	if !util.Abc_InfoHasBit(cex.PData, 0) || util.Abc_InfoHasBit(cex.PData, 1) {
		t.Fatalf("boundary cex %#v", cex.PData)
	}
	real := Gla_ManDeriveCex(p, mapping)
	if !util.Abc_InfoHasBit(real.PData, 0) || util.Abc_InfoHasBit(real.PData, 1) {
		t.Fatalf("real cex %#v", real.PData)
	}
}
