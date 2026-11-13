package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SimulationSupportsBuffers(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a := gia.Gia_ManAppendCi(g)
	buf := gia.Gia_ManAppendBuf(g, a^1)
	gia.Gia_ManAppendCo(g, buf)
	sim := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	sim.setInputBit(a>>1, 0, true)
	cec4SimulateGraph(sim, g)
	if got := sim.object(buf >> 1)[0]; got != ^uint64(1) {
		t.Fatalf("buffer simulation %#x", got)
	}
	sim.setInputBit(a>>1, 1, true)
	stamps := make([]int32, gia.Gia_ManObjNum(g))
	cec4ResimulateNode(sim, g, buf>>1, 1, stamps)
	if got := sim.object(buf >> 1)[0]; got != ^uint64(3) {
		t.Fatalf("buffer local resimulation %#x", got)
	}
}
