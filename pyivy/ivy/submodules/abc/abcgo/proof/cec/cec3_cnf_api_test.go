package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose"
	"testing"
)

func TestCec3GlucoseCNF(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		a, b := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, x^1, z)
		root := gia.Gia_ManAppendAnd(g, a^1, b^1)
		if mux {
			gia.Gia_ObjSetMark0(gia.Gia_ManObj(g, root>>1), 1)
		}
		p := &Cec3_Man_t{pNew: g, pSat: glucose.Bmcg_sat_solver_start(), vFrontier: vec.Vec_PtrAlloc(8), vFanins: vec.Vec_PtrAlloc(8)}
		rv := Cec3_ObjGetCnfVar(p, root>>1)
		vars := []int32{Cec3_ObjGetCnfVar(p, x>>1), Cec3_ObjGetCnfVar(p, y>>1), Cec3_ObjGetCnfVar(p, z>>1), rv}
		for mask := 0; mask < 16; mask++ {
			assumptions := make([]int32, 4)
			for i, v := range vars {
				assumptions[i] = v * 2
				if mask&(1<<i) == 0 {
					assumptions[i]++
				}
			}
			want := !((mask&1 != 0 && mask&2 != 0) || (mask&1 == 0 && mask&4 != 0))
			got := glucose.Bmcg_sat_solver_solve(p.pSat, assumptions, 4) == glucose.GLUCOSE_SAT
			if got != (want == (mask&8 != 0)) {
				t.Fatalf("mux=%v mask=%04b", mux, mask)
			}
		}
	}
}
