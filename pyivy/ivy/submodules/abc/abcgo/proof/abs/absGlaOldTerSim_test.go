package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestOldGlaTernaryVerificationClearsMarks(t *testing.T) {
	g := gia.Gia_ManStart(10)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	and := gia.Gia_ManAppendAnd(g, x, y)
	out := gia.Gia_ManAppendCo(g, and)
	p := &Gla_Man_t{pGia: g, pPars: &Abs_Par_t{}, pObjs: make([]Gla_Obj_t, 3), nObjs: 3, pObj2Obj: make([]uint32, gia.Gia_ManObjNum(g)), pSat: bsat.NewLabeledSolver2()}
	p.pObjs[1].id = 1
	p.pObjs[2].id = 2
	p.pObj2Obj[x>>1] = 1
	p.pObj2Obj[y>>1] = 2
	p.pSat.NewVariable()
	vx, vy := p.pSat.NewVariable(), p.pSat.NewVariable()
	vec.Vec_IntSetEntry(&p.pObjs[1].vFrames, 0, vx)
	vec.Vec_IntSetEntry(&p.pObjs[2].vFrames, 0, vy)
	if p.pSat.Solve(bsat.Positive(vx), bsat.Positive(vy)) != bsat.Satisfiable {
		t.Fatal("SAT setup")
	}
	v := func(ids ...int32) *vec.Vec_Int_t { return vec.Vec_IntAllocArrayCopy(ids, len(ids)) }
	Gla_ManVerifyUsingTerSim(p, v(x>>1), v(y>>1), v(and>>1), v(out>>1), v(y>>1))
	for _, id := range []int32{0, x >> 1, y >> 1, and >> 1, out >> 1} {
		if gia.Gia_ObjTerSimGetC(gia.Gia_ManObj(g, id)) == 0 {
			t.Fatalf("object %d retains temporary ternary state", id)
		}
	}
}
