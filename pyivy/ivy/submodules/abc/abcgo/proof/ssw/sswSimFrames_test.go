package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestFrameCopyAndTransfer(t *testing.T) {
	g := aig.NewGraph()
	input := g.NewInput()
	output := g.AddOutput(input.Not())
	p := Ssw_SmlStart(g, 0, 2, 2)
	x := Ssw_SmlSimInfo(p, input)
	x[0], x[1] = 0x12, 0x34
	Ssw_SmlNodeCopyFanin(p, output, 0)
	y := Ssw_SmlSimInfo(p, output)
	if y[0] != ^uint32(0x12) || y[1] != ^uint32(0x34) {
		t.Fatalf("incorrect output: %x", y)
	}
	Ssw_SmlNodeTransferNext(p, output, input, 0)
	if Ssw_SmlNodesCompareInFrame(p, output, input, 0, 1) != 1 {
		t.Fatal("transfer failed")
	}
	y[2], y[3] = 55, 66
	Ssw_SmlNodeTransferFirst(p, output, input)
	if x[0] != 55 || x[1] != 66 {
		t.Fatal("wraparound transfer failed")
	}
}

func TestFrameAndSimulationMatchesBitwiseLogic(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a.Not(), b)
	p := Ssw_SmlStart(g, 0, 1, 1)
	Ssw_SmlSimInfo(p, a)[0] = 0xf0f0
	Ssw_SmlSimInfo(p, b)[0] = 0xcc00
	Ssw_SmlNodeSimulate(p, gate, 0)
	if got, want := Ssw_SmlSimInfo(p, gate)[0], ^uint32(0xf0f0)&uint32(0xcc00); got != want {
		t.Fatalf("got %x want %x", got, want)
	}
}
