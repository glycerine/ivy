package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestCec2ReplayCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, root)
	gia.Gia_ManCreateRefs(g)
	sim := newCec3Simulation(g.ObjectCount(), 2)
	s := satoko.New()
	cnf := newCec2CNF(g, s)
	status, inputs := cnf.solveTwoWithCone(0, root>>1, 0, 0, true)
	if status != satoko.SATOKO_SAT {
		t.Fatalf("status %d", status)
	}
	cec2ReplayModel(sim, g, s, inputs, 70)
	cec2SimulateAnds(sim, g)
	output, pattern := cec2FailingOutput(sim, g)
	if output != 0 || pattern != 70 {
		t.Fatalf("output=%d pattern=%d", output, pattern)
	}
}

func TestCec2RandomizePreservesZeroPattern(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	y := gia.Gia_ManAppendCi(g)
	sim := newCec3Simulation(g.ObjectCount(), 2)
	cec2RandomizeCIs(sim, g)
	if sim.object(x >> 1)[0]&1 != 0 || sim.object(y >> 1)[0]&1 != 0 {
		t.Fatal("pattern zero is not all-zero")
	}
}
