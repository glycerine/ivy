package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSequentialSimulationTransfersState(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	po := g.AddOutput(gate)
	li := g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 2, 1)
	Ssw_SmlSimInfo(p, pi)[0] = 0x0f
	Ssw_SmlSimInfo(p, pi)[1] = 0x33
	Ssw_SmlSimInfo(p, lo)[0] = 0x55
	Ssw_SmlSimulateOne(p)
	if got := Ssw_SmlSimInfo(p, po); got[0] != 0x05 || got[1] != 0x03 {
		t.Fatalf("outputs %x", got)
	}
	if Ssw_SmlSimInfo(p, li)[0] != 0x0f || Ssw_SmlSimInfo(p, lo)[1] != 0x0f || p.nSimRounds != 1 {
		t.Fatal("incorrect state transfer")
	}
}

func TestSimulationExplicitXor(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.NewXor(a, b.Not())
	out := g.AddOutput(x)
	p := Ssw_SmlStart(g, 0, 1, 1)
	Ssw_SmlSimInfo(p, a)[0] = 0xf0
	Ssw_SmlSimInfo(p, b)[0] = 0x33
	Ssw_SmlSimulateOne(p)
	if got, want := Ssw_SmlSimInfo(p, out)[0], uint32(0xf0)^^uint32(0x33); got != want {
		t.Fatalf("got %x want %x", got, want)
	}
}

func TestUnnormalizeConvertsFirstFrameConstant(t *testing.T) {
	g := aig.NewGraph()
	p := Ssw_SmlStart(g, 0, 2, 1)
	Ssw_SmlUnnormalize(p)
	words := Ssw_SmlSimInfo(p, g.Const1())
	if words[0] != ^uint32(0) || words[1] != 0 {
		t.Fatalf("got %x", words)
	}
}
