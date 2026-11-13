package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestDynamicSimulationRecursesAcrossFrames(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	out := g.AddOutput(gate)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 2, 1)
	Ssw_SmlSimInfo(p, pi)[0] = 0xf
	Ssw_SmlSimInfo(p, pi)[1] = 0x33
	visited := make([]int32, len(g.Nodes())*2)
	Ssw_SmlSimulateOneDyn_rec(p, out, 1, visited, 1)
	if got := Ssw_SmlSimInfo(p, out)[1]; got != 3 {
		t.Fatalf("got %x", got)
	}
}
func TestSingleFrameSimulationTransfersNextLatchState(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 2, 1)
	Ssw_SmlSimInfo(p, pi)[0] = 0x35
	Ssw_SmlSimulateOneFrame(p)
	if got := Ssw_SmlSimInfo(p, lo)[1]; got != 0x35 || p.nSimRounds != 1 {
		t.Fatalf("next %x rounds %d", got, p.nSimRounds)
	}
}
