package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func TestSimulationCounterexampleReplays(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 2, 1)
	Ssw_SmlSimInfo(p, pi)[0] = 4
	Ssw_SmlSimulateOne(p)
	p.fNonConstOut = Ssw_SmlCheckNonConstOutputs(p)
	cex := Ssw_SmlGetCounterExample(p)
	if cex == nil || cex.IFrame != 1 || saig.Saig_ManVerifyCex(g, cex) == 0 {
		t.Fatal("counterexample failed replay")
	}
}

func TestSimulationCounterexampleRejectsFailedReplay(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	out := g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 1, 1)
	defer Ssw_SmlStop(p)
	// Simulate inconsistent output data: ABC verifies the reconstructed witness
	// against the graph, frees it on failed replay, and returns NULL.
	Ssw_SmlSimInfo(p, out)[0] = 1
	p.fNonConstOut = 1
	if cex := Ssw_SmlGetCounterExample(p); cex != nil {
		t.Fatal("expected a rejected witness when replay does not fail the output")
	}
}
func TestSimulationOutputPattern(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	out := g.AddOutput(a)
	p := Ssw_SmlStart(g, 0, 1, 1)
	Ssw_SmlSimInfo(p, a)[0] = 4
	model := Ssw_SmlCheckOutput(p)
	if len(model) != 2 || model[0] != 1 || model[1] != int32(out.NodeID()) {
		t.Fatalf("model %v", model)
	}
}

func TestSimulationOutputPatternRetainsSourcePrefixBehavior(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	p := Ssw_SmlStart(g, 1, 1, 1)
	Ssw_SmlSimInfo(p, a)[0] = 1
	Ssw_SmlSimInfo(p, a)[1] = 2
	Ssw_SmlSimInfo(p, b)[1] = 2
	model := Ssw_SmlCheckOutput(p)
	if len(model) != 3 || model[0] != 1 || model[1] != 0 {
		t.Fatalf("expected source prefix pattern, got %v", model)
	}
}
