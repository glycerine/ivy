package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestCec2SolveSourceEntryPoints(t *testing.T) {
	for _, cones := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCi(g)
		p := Cec2_ManCreate(g, &Cec2_Par_t{nConfLimit: 1000, fUseCones: cones})
		root := gia.Gia_ManAppendAnd(p.pNew, 2, 4)
		gia.Gia_ManExtendCopy2Array(p.pNew)
		if got := Cec2_ManSolveTwo(p, 0, root>>1, 0); got != satoko.SATOKO_SAT {
			t.Fatalf("cones=%d result=%d", cones, got)
		}
		pairs := vec.Vec_IntArray(p.vObjSatPairs)
		if len(pairs) != 4 {
			t.Fatalf("cones=%d pairs=%v", cones, pairs)
		}
		for i := 1; i < len(pairs); i += 2 {
			if p.pSat.ModelValue(int(pairs[i])) != 1 {
				t.Fatalf("cones=%d missing witness model", cones)
			}
		}
		if cones == 0 {
			p.pSat.Rollback()
		}
		Cec2_ManDestroy(p)
	}
}
