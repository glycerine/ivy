package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose"
	"testing"
)

func TestCec3SolveSourceEntryPoints(t *testing.T) {
	for _, cones := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCi(g)
		p := Cec3_ManCreate(g, &Cec3_Par_t{nConfLimit: 1000, fUseCones: cones})
		root := gia.Gia_ManAppendAnd(p.pNew, 2, 4)
		gia.Gia_ManExtendCopy2Array(p.pNew)
		if got := Cec3_ManSolveTwo(p, 0, root>>1, 0); got != glucose.GLUCOSE_SAT {
			t.Fatalf("cones=%d result=%d", cones, got)
		}
		pairs := vec.Vec_IntArray(p.vObjSatPairs)
		if len(pairs) != 4 {
			t.Fatalf("pairs=%v", pairs)
		}
		for i := 1; i < len(pairs); i += 2 {
			if glucose.Bmcg_sat_solver_read_cex_varvalue(p.pSat, pairs[i]) != 1 {
				t.Fatal("bad Glucose model")
			}
		}
		Cec3_ManDestroy(p)
	}
}
