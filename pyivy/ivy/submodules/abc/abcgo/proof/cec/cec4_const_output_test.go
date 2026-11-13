package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SimulationFindsConstantOneMiter(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCo(g, 1)
	sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	cec4SimulateGraph(sim, g)
	output, cex := cec4FirstFailingOutput(sim, g)
	if output != 0 || cex == nil || cex.IPo != 0 {
		t.Fatalf("missed constant-one miter: output %d cex %#v", output, cex)
	}
}
