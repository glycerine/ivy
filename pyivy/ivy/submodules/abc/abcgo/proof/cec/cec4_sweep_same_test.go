package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
	"testing"
)

func TestCec4SweepNodeHandlesSameReducedObject(t *testing.T) {
	for _, different := range []bool{false, true} {
		g := gia.Gia_ManStart(10)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendAnd(g, a, b)
		y := gia.Gia_ManAppendAnd(g, a, b)
		gia.Gia_ManSetPhase(g)
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, y>>1, x>>1)
		copies := make([]int32, gia.Gia_ManObjNum(g))
		for i := range copies {
			copies[i] = int32(i) << 1
		}
		copies[y>>1] = x
		if different {
			copies[y>>1] ^= 1
		}
		sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
		cnf := newCec4CNF(g, 0)
		bit := 0
		if different {
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected invariant failure for opposite phase")
					}
				}()
				cec4SweepNode(g, g, sim, cnf, copies, y>>1, x>>1, 100, &bit)
			}()
			if bit != 0 {
				t.Fatal("invariant failure must not fabricate a pattern")
			}
		} else {
			got := cec4SweepNode(g, g, sim, cnf, copies, y>>1, x>>1, 100, &bit)
			if got != glucose2.GLUCOSE_UNSAT || gia.Gia_ObjProved(g, y>>1) == 0 {
				t.Fatalf("same relation: %d", got)
			}
		}
	}
}
