package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
)

func TestCec4ManagerSATAndRecycle(t *testing.T) {
	for mode := int32(0); mode <= 2; mode++ {
		g := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendAnd(g, a, b)
		y := gia.Gia_ManAppendAnd(g, a, b)
		m := &Cec4_Man_t{pNew: g, pPars: &Cec_ParFra_t{jType: mode, nBTLimit: 1000, nCallsRecycle: 100, nSatVarMax: 100}, vFails: vec.Vec_BitStart(int(gia.Gia_ManObjNum(g)))}
		var easy int32
		if got := Cec4_ManSolveTwo(m, x>>1, y>>1, 0, &easy, 1); got != glucose2.GLUCOSE_UNSAT {
			t.Fatalf("mode %d equal gates: %d", mode, got)
		}
		if m.pSat != m.cnf.solver || m.nCallsSince != 1 {
			t.Fatal("manager solver was not initialized")
		}
		if got := Cec4_ManSolveTwo(m, a>>1, x>>1, 0, &easy, 1); got != glucose2.GLUCOSE_SAT {
			t.Fatalf("mode %d unequal gates: %d", mode, got)
		}
		old := m.pSat
		Cec4_ManSatSolverRecycle(m)
		if m.pSat == old || m.nCallsSince != 0 || m.nRecycles != 1 || len(m.cnf.ids) != 0 {
			t.Fatal("recycle retained old SAT state")
		}
		if got := Cec4_ManSolveTwo(m, x>>1, y>>1, 0, &easy, 1); got != glucose2.GLUCOSE_UNSAT {
			t.Fatalf("mode %d post-recycle: %d", mode, got)
		}
	}
}
