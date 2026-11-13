package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// The C combinational entry point initializes one frame and simulates it.
func TestCombinationalSimulationEntryPoint(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a.Not(), b)
	out := g.AddOutput(gate)
	sim := Ssw_SmlSimulateComb(g, 2)
	defer Ssw_SmlStop(sim)
	if Ssw_SmlNumFrames(sim) != 1 || Ssw_SmlNumWordsTotal(sim) != 2 {
		t.Fatal("unexpected simulation dimensions")
	}
	x, y, z, o := Ssw_SmlSimInfo(sim, a), Ssw_SmlSimInfo(sim, b), Ssw_SmlSimInfo(sim, gate), Ssw_SmlSimInfo(sim, out)
	for i := range z {
		if z[i] != ^x[i]&y[i] || o[i] != z[i] {
			t.Fatalf("word=%d a=%x b=%x gate=%x out=%x", i, x[i], y[i], z[i], o[i])
		}
	}
}
