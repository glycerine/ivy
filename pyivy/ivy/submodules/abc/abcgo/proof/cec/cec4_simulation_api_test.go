package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SimulationAndRefinementEntryPoints(t *testing.T) {
	g := gia.Gia_ManStart(16)
	g.EnableMuxes()
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	xor := gia.Gia_ManAppendXorReal(g, x, y)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, xor)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{})
	Cec4_ManSimAlloc(p, 2)
	p.pattern = 70
	Cec4_ObjSimSetInputBit(p, x>>1, 1)
	Cec4_ObjSimSetInputBit(p, y>>1, 1)
	if Cec4_ObjSimGetInputBit(p, x>>1) != 1 || p.packer.values != p.simState.sim {
		t.Fatal("simulation state not shared")
	}
	Cec4_ManSimulate(p)
	if Cec4_ObjSim(p, a>>1)[1]&(1<<6) == 0 || Cec4_ObjSim(p, xor>>1)[1]&(1<<6) != 0 {
		t.Fatal("AND/XOR simulation incorrect")
	}
	if Cec4_ManSimulateCos(p) != 0 || g.SequentialCounterexample() == nil {
		t.Fatal("missing witness")
	}
	Cec4_ObjSimSetInputBit(p, y>>1, 0)
	Cec4_ManSimulate(p)
	if Cec4_ObjSim(p, xor>>1)[1]&(1<<6) == 0 || Cec4_ObjSim(p, a>>1)[1]&(1<<6) != 0 {
		t.Fatal("incremental simulation incorrect")
	}
	cex := Cec4_ManDeriveCex(p, 1, 70)
	if cex.PData[0] != 1 {
		t.Fatalf("bad pattern %x", cex.PData[0])
	}
	Cec4_ManClearCis(p)
	if p.pattern != 0 || Cec4_ObjSim(p, x>>1)[1] != 0 {
		t.Fatal("CI clear failed")
	}
	Cec4_ManDestroy(p)
	if p.simState != nil || p.packer != nil {
		t.Fatal("simulation buffers retained after destroy")
	}
}

func TestCec4DirectOutputSimulationMatchesSourcePhase(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCo(g, 1)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{})
	defer Cec4_ManDestroy(p)
	Cec4_ManSimAlloc(p, 1)
	if got := Cec4_ManSimulateCos(p); got != 1 || g.SequentialCounterexample() != nil {
		t.Fatalf("source phase normalization changed: %d", got)
	}
}
