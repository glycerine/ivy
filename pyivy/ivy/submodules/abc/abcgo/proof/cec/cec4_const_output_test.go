package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SimulationMatchesSourceConstantPhase(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCo(g, 1)
	sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	cec4SimulateGraph(sim, g)
	output, cex := cec4FirstFailingOutput(sim, g)
	if output != -1 || cex != nil {
		t.Fatalf("source phase normalization changed: output %d cex %#v", output, cex)
	}
}
