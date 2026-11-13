package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
	"testing"
)

func TestCec4SATReplayDistinguishesGatePairs(t *testing.T) {
	for mode := int32(0); mode < 3; mode++ {
		g := gia.Gia_ManStart(32)
		g.EnableMuxes()
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		roots := []int32{gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, c), gia.Gia_ManAppendXorReal(g, a, b), gia.Gia_ManAppendMuxReal(g, a, b, c)}
		gia.Gia_ManSetPhase(g)
		g.StartRepresentatives()
		for i := 1; i < len(roots); i++ {
			for j := 0; j < i; j++ {
				id, repr := roots[i]>>1, roots[j]>>1
				gia.Gia_ObjSetRepr(g, id, repr)
				copies := make([]int32, gia.Gia_ManObjNum(g))
				for k := range copies {
					copies[k] = int32(k) << 1
				}
				sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
				cnf := newCec4CNF(g, mode)
				bit := 0
				status := cec4SweepNode(g, g, sim, cnf, copies, id, repr, 1000, &bit)
				if status != glucose2.GLUCOSE_SAT {
					t.Fatalf("mode %d pair %d %d: status %d", mode, id, repr, status)
				}
				cec4SimulateGraph(sim, g)
				valueA := (sim.object(id)[0] >> uint(bit)) & 1
				valueB := (sim.object(repr)[0] >> uint(bit)) & 1
				phase := uint64(gia.Gia_ObjPhase(gia.Gia_ManObj(g, id)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(g, repr)))
				if valueA^valueB == phase {
					t.Fatalf("mode %d pair %d %d replay does not distinguish", mode, id, repr)
				}
			}
		}
	}
}
