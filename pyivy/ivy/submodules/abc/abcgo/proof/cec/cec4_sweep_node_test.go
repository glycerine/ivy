package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
	"testing"
)

func TestCec4SweepNodeProvesOrReplaysCounterexample(t *testing.T) {
	for _, equivalent := range []bool{false, true} {
		g := gia.Gia_ManStart(10)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendAnd(g, a, b)
		y := gia.Gia_ManAppendAnd(g, a, b)
		if !equivalent {
			y = gia.Gia_ManAppendAnd(g, a, b^1)
		}
		gia.Gia_ManSetPhase(g)
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, y>>1, x>>1)
		copies := make([]int32, gia.Gia_ManObjNum(g))
		for id := range copies {
			copies[id] = int32(id) << 1
		}
		sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
		cnf := newCec4CNF(g, 0)
		bit := 0
		status := cec4SweepNode(g, g, sim, cnf, copies, y>>1, x>>1, 100, &bit)
		if equivalent {
			if status != glucose2.GLUCOSE_UNSAT || gia.Gia_ObjProved(g, y>>1) == 0 {
				t.Fatalf("equivalent status %d", status)
			}
		} else {
			if status != glucose2.GLUCOSE_SAT || bit != 1 {
				t.Fatalf("different status %d bit %d", status, bit)
			}
			cec4SimulateNode(sim, g, x>>1)
			cec4SimulateNode(sim, g, y>>1)
			if (sim.object(x >> 1)[0]^sim.object(y >> 1)[0])&2 == 0 {
				t.Fatal("SAT model replay did not distinguish pair")
			}
		}
	}
}
