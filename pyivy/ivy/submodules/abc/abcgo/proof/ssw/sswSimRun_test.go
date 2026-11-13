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
