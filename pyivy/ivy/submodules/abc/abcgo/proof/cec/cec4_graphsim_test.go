package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4GraphSimulationProducesReplayableWitness(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b^1)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, gate)
	sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 2)
	sim.setInputBit(a>>1, 70, true)
	cec4SimulateGraph(sim, g)
	out, cex := cec4FirstFailingOutput(sim, g)
	if out != 1 || cex == nil || cex.IPo != 1 {
		t.Fatalf("output %d witness %+v", out, cex)
	}
	values := []bool{cex.PData[0]&1 != 0, cex.PData[0]&2 != 0}
	result, err := g.Evaluate(values)
	if err != nil {
		t.Fatal(err)
	}
	if !result[1] {
		t.Fatalf("witness %v did not fail output", values)
	}
}
