package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// Cec2/3_ManSweepNode must either substitute a proven representative or
// install a SAT pattern that separates an incorrect candidate (cecSat.c:864,
// cecSatG.c:861). Exercise both solver reset and cone-retention modes.
func TestCec23ManagerSweepNodeSourceOutcomes(t *testing.T) {
	for _, mode := range []int{2, 3} {
		for _, cones := range []int32{0, 1} {
			for _, equivalent := range []bool{false, true} {
				g := gia.Gia_ManStart(12)
				a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
				x := gia.Gia_ManAppendAnd(g, a, b)
				y := gia.Gia_ManAppendAnd(g, x, a)
				if !equivalent {
					y = gia.Gia_ManAppendAnd(g, a, b^1)
				}
				gia.Gia_ManSetPhase(g)
				g.StartRepresentatives()
				gia.Gia_ObjSetRepr(g, y>>1, x>>1)
				gia.Gia_ManCecSimAlloc(g, 1)
				var got int32
				if mode == 2 {
					p := Cec2_ManCreate(g, &Cec2_Par_t{nConfLimit: 1000, fUseCones: cones})
					p.pNew = g
					gia.Gia_ManExtendCopy2Array(g)
					for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
						gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), id<<1)
					}
					got = Cec2_ManSweepNode(p, y>>1)
				} else {
					p := Cec3_ManCreate(g, &Cec3_Par_t{nConfLimit: 1000, fUseCones: cones})
					p.pNew = g
					gia.Gia_ManExtendCopy2Array(g)
					for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
						gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), id<<1)
					}
					got = Cec3_ManSweepNode(p, y>>1)
				}
				if equivalent {
					if got != 1 || gia.Gia_ObjProved(g, y>>1) == 0 || gia.Gia_ObjValue(gia.Gia_ManObj(g, y>>1)) != x {
						t.Fatalf("mode=%d cones=%d equivalent: result=%d", mode, cones, got)
					}
				} else {
					if got != 0 || gia.Gia_ManCecSimPattern(g) != 1 {
						t.Fatalf("mode=%d cones=%d inequivalent: result=%d", mode, cones, got)
					}
					Cec2_ObjSimAnd(g, x>>1)
					Cec2_ObjSimAnd(g, y>>1)
					if (Cec2_ObjSim(g, x>>1)[0]^Cec2_ObjSim(g, y>>1)[0])&2 == 0 {
						t.Fatalf("mode=%d cones=%d SAT pattern did not separate", mode, cones)
					}
				}
			}
		}
	}
}
